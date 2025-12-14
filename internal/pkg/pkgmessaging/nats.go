package pkgmessaging

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
	ErrNATSSubjectRequired = errors.New("pkgmessage: nats subject is required")
	ErrNATSURLRequired     = errors.New("pkgmessage: nats url is required")
	ErrNATSHandlerRequired = errors.New("pkgmessage: nats handler is required")
)

type NATSConfig struct {
	URL string

	Options []nats.Option
}

type NATS struct {
	conn *nats.Conn

	mu     sync.Mutex
	subs   []*nats.Subscription
	closed bool
}

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
	sub, wg, err := n.subscribeNATS(ctx, source, handler, co)
	if err != nil {
		return err
	}

	if err := n.addNATSSub(sub); err != nil {
		if uerr := sub.Unsubscribe(); uerr != nil {
			return errors.Join(err, uerr)
		}
		return err
	}

	if err := n.conn.Flush(); err != nil {
		if uerr := sub.Unsubscribe(); uerr != nil {
			return errors.Join(fmt.Errorf("pkgmessage: nats flush: %w", err), uerr)
		}
		return fmt.Errorf("pkgmessage: nats flush: %w", err)
	}

	return n.waitNATSConsume(ctx, sub, wg)
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

func (n *NATS) subscribeNATS(ctx context.Context, subject string, handler Handler, opts consumeOptions) (*nats.Subscription, *sync.WaitGroup, error) {
	queueGroup := queueGroupFromConsumeOptions(opts)
	concurrency := concurrencyOrDefault(opts.concurrency, 1)
	autoAck := opts.autoAck

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	sub, err := n.conn.QueueSubscribe(subject, queueGroup, func(m *nats.Msg) {
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return
		}

		wg.Add(1)
		go func(msg *nats.Msg) {
			defer wg.Done()
			defer func() { <-sem }()

			wrapped := newNATSMessage(msg, time.Now())
			herr := handler(ctx, wrapped)

			if wrapped.hasResponded() || !autoAck {
				return
			}

			if err := autoAckNATS(ctx, wrapped, herr); err != nil {
				return
			}
		}(m)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("pkgmessage: nats subscribe: %w", err)
	}

	return sub, &wg, nil
}

func (n *NATS) waitNATSConsume(ctx context.Context, sub *nats.Subscription, wg *sync.WaitGroup) error {
	<-ctx.Done()

	uerr := sub.Unsubscribe()
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
