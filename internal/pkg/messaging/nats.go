package messaging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	// ErrNATSSubjectRequired is returned when the subject is empty.
	ErrNATSSubjectRequired = errors.New("pkgmessage: nats subject is required")
	// ErrNATSURLRequired is returned when the NATS server URL is missing.
	ErrNATSURLRequired = errors.New("pkgmessage: nats url is required")
	// ErrNATSHandlerRequired is returned when Consume is called with a nil handler.
	ErrNATSHandlerRequired = errors.New("pkgmessage: nats handler is required")
)

// NATSConfig configures the NATS implementation.
type NATSConfig struct {
	// URL is the NATS server address.
	URL string

	// Options are passed to the NATS client.
	Options []nats.Option
}

// NATS is a messaging implementation backed by NATS.
type NATS struct {
	conn *nats.Conn

	mu     sync.Mutex
	subs   []*nats.Subscription
	closed bool
}

// NewNATS constructs a NATS messaging client.
func NewNATS(cfg NATSConfig) (*NATS, error) {
	if cfg.URL == "" {
		return nil, ErrNATSURLRequired
	}

	conn, err := nats.Connect(cfg.URL, cfg.Options...)
	if err != nil {
		return nil, fmt.Errorf("pkgmessage: nats connect: %w", err)
	}

	return &NATS{
		conn: conn,
	}, nil
}

// Close drains subscriptions and closes the NATS connection.
func (n *NATS) Close() error {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return nil
	}
	n.closed = true
	subs := append([]*nats.Subscription{}, n.subs...)
	n.mu.Unlock()

	var closeErr error
	for _, sub := range subs {
		if err := sub.Drain(); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}

	if err := n.conn.Drain(); err != nil {
		closeErr = errors.Join(closeErr, err)
	}
	n.conn.Close()
	return closeErr
}

// Publish sends a message to a NATS subject.
func (n *NATS) Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if destination == "" {
		return PublishResult{}, ErrNATSSubjectRequired
	}
	if msg.Delay > 0 {
		return PublishResult{}, ErrUnsupported
	}

	nmsg := nats.NewMsg(destination)
	nmsg.Data = msg.Body

	for _, h := range msg.Headers {
		if h.Key == "" {
			continue
		}
		nmsg.Header.Add(h.Key, string(h.Value))
	}

	if err := n.conn.PublishMsg(nmsg); err != nil {
		return PublishResult{}, fmt.Errorf("pkgmessage: nats publish: %w", err)
	}
	if err := n.conn.Flush(); err != nil {
		return PublishResult{}, fmt.Errorf("pkgmessage: nats flush: %w", err)
	}

	return PublishResult{
		Topic:     destination,
		Timestamp: time.Now(),
	}, nil
}

// Consume starts consuming messages from a NATS subject.
func (n *NATS) Consume(ctx context.Context, source string, handler Handler, opts ...ConsumeOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if source == "" {
		return ErrNATSSubjectRequired
	}
	if handler == nil {
		return ErrNATSHandlerRequired
	}

	co := newConsumeOptions(opts...)
	sub, wg, msgCh, err := n.subscribeNATS(ctx, source, handler, co)
	if err != nil {
		return err
	}

	if err := n.addNATSSub(sub); err != nil {
		uerr := sub.Drain()
		close(msgCh)
		wg.Wait()
		if uerr != nil {
			return errors.Join(err, uerr)
		}
		return err
	}

	if err := n.conn.Flush(); err != nil {
		ferr := fmt.Errorf("pkgmessage: nats flush: %w", err)
		uerr := sub.Drain()
		close(msgCh)
		wg.Wait()
		if uerr != nil {
			return errors.Join(ferr, uerr)
		}
		return ferr
	}

	return n.waitNATSConsume(ctx, sub, msgCh, wg)
}

func (n *NATS) addNATSSub(sub *nats.Subscription) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return io.ErrClosedPipe
	}
	n.subs = append(n.subs, sub)
	return nil
}

func (n *NATS) subscribeNATS(ctx context.Context, subject string, handler Handler, opts consumeOptions) (*nats.Subscription, *sync.WaitGroup, chan *nats.Msg, error) {
	queueGroup := queueGroupFromConsumeOptions(opts)
	concurrency := concurrencyOrDefault(opts.concurrency, 1)
	autoAck := opts.autoAck

	msgCh := make(chan *nats.Msg, concurrency)
	var wg sync.WaitGroup

	sub, err := n.conn.QueueSubscribe(subject, queueGroup, func(m *nats.Msg) {
		select {
		case msgCh <- m:
		case <-ctx.Done():
		}
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("pkgmessage: nats subscribe: %w", err)
	}

	for range concurrency {
		wg.Go(func() {
			for msg := range msgCh {
				wrapped := newNATSMessage(msg, time.Now())
				herr := callHandlerWithRecover(ctx, "nats", func() error {
					return handler(ctx, wrapped)
				})
				if wrapped.hasResponded() || !autoAck {
					continue
				}
				//nolint:errcheck // ignore for now
				_ = autoAckNATS(ctx, wrapped, herr)
			}
		})
	}

	return sub, &wg, msgCh, nil
}

func (n *NATS) waitNATSConsume(ctx context.Context, sub *nats.Subscription, msgCh chan *nats.Msg, wg *sync.WaitGroup) error {
	<-ctx.Done()

	uerr := sub.Drain()
	close(msgCh)
	wg.Wait()

	return errors.Join(ctx.Err(), uerr)
}

func queueGroupFromConsumeOptions(opts consumeOptions) string {
	queueGroup := opts.queueGroup
	if opts.params != nil {
		if v, ok := opts.params["queue_group"]; ok && v != "" {
			queueGroup = v
		}
	}
	return queueGroup
}

func concurrencyOrDefault(n, def int) int {
	if n <= 0 {
		return def
	}
	return n
}

func autoAckNATS(ctx context.Context, msg interface {
	Ack(context.Context) error
	Nack(context.Context) error
}, handlerErr error) error {
	if handlerErr == nil {
		return msg.Ack(ctx)
	}
	return msg.Nack(ctx)
}
