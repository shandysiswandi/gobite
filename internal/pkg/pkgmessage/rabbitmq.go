package pkgmessage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrRabbitMQURLRequired   = errors.New("pkgmessage: rabbitmq url is required")
	ErrRabbitMQHandlerNeeded = errors.New("pkgmessage: rabbitmq handler is required")
	ErrRabbitMQQueueRequired = errors.New("pkgmessage: rabbitmq queue is required")
)

type RabbitMQConfig struct {
	URL string

	Connection *amqp.Connection

	DialConfig *amqp.Config
}

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	mu     sync.Mutex
	closed bool
}

func NewRabbitMQ(cfg RabbitMQConfig) (*RabbitMQ, error) {
	var conn *amqp.Connection
	if cfg.Connection != nil {
		conn = cfg.Connection
	} else {
		if cfg.URL == "" {
			return nil, ErrRabbitMQURLRequired
		}
		c, err := dialRabbit(cfg.URL, cfg.DialConfig)
		if err != nil {
			return nil, err
		}
		conn = c
	}

	ch, err := conn.Channel()
	if err != nil {
		closeErr := conn.Close()
		return nil, errors.Join(fmt.Errorf("pkgmessage: rabbitmq open channel: %w", err), closeErr)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	ch := r.channel
	conn := r.conn
	r.channel = nil
	r.conn = nil
	r.mu.Unlock()

	var closeErr error
	if ch != nil {
		closeErr = errors.Join(closeErr, ch.Close())
	}
	if conn != nil {
		closeErr = errors.Join(closeErr, conn.Close())
	}
	return closeErr
}

func (r *RabbitMQ) Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if destination == "" {
		return PublishResult{}, ErrRabbitMQQueueRequired
	}
	if msg.Delay > 0 {
		return PublishResult{}, ErrUnsupported
	}
	if err := r.ensureRabbitOpen(); err != nil {
		return PublishResult{}, err
	}

	exchange, routingKey := resolveRabbitPublish(destination, msg)
	pub := amqp.Publishing{
		Timestamp: time.Now(),
		Body:      msg.Body,
		Headers:   headersToTable(msg.Headers, msg.Attributes),
	}

	if err := r.channel.PublishWithContext(ctx, exchange, routingKey, false, false, pub); err != nil {
		return PublishResult{}, fmt.Errorf("pkgmessage: rabbitmq publish: %w", err)
	}

	return PublishResult{
		Topic:     exchange,
		Timestamp: pub.Timestamp,
	}, nil
}

func (r *RabbitMQ) Consume(ctx context.Context, source string, handler Handler, opts ConsumeOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if source == "" {
		return ErrRabbitMQQueueRequired
	}
	if handler == nil {
		return ErrRabbitMQHandlerNeeded
	}
	if err := r.ensureRabbitOpen(); err != nil {
		return err
	}

	consumeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	tag := fmt.Sprintf("gobite-%d", time.Now().UnixNano())
	deliveries, err := r.startConsume(source, tag, opts)
	if err != nil {
		return err
	}

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	autoAck := opts.AutoAck
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for range concurrency {
		go func() {
			defer wg.Done()
			for d := range deliveries {
				if err := handleRabbitDelivery(consumeCtx, d, handler, autoAck); err != nil {
					trySendErr(errCh, err)
					cancel()
					return
				}
			}
		}()
	}

	select {
	case <-consumeCtx.Done():
		cancelErr := r.channel.Cancel(tag, false)
		wg.Wait()
		return errors.Join(consumeCtx.Err(), cancelErr)
	case err := <-errCh:
		cancelErr := r.channel.Cancel(tag, false)
		wg.Wait()
		return errors.Join(err, cancelErr)
	}
}

func (r *RabbitMQ) ensureRabbitOpen() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return io.ErrClosedPipe
	}
	if r.channel == nil {
		return io.ErrClosedPipe
	}
	return nil
}

func dialRabbit(url string, cfg *amqp.Config) (*amqp.Connection, error) {
	if cfg == nil {
		conn, err := amqp.Dial(url)
		if err != nil {
			return nil, fmt.Errorf("pkgmessage: rabbitmq dial: %w", err)
		}
		return conn, nil
	}

	conn, err := amqp.DialConfig(url, *cfg)
	if err != nil {
		return nil, fmt.Errorf("pkgmessage: rabbitmq dial: %w", err)
	}
	return conn, nil
}

func (r *RabbitMQ) startConsume(queue, consumerTag string, opts ConsumeOptions) (<-chan amqp.Delivery, error) {
	if opts.Exchange != "" && opts.RoutingKey != "" && opts.Params != nil {
		if v, ok := opts.Params["bind"]; ok {
			bind, err := strconv.ParseBool(v)
			if err == nil && bind {
				if err := r.channel.QueueBind(queue, opts.RoutingKey, opts.Exchange, false, nil); err != nil {
					return nil, fmt.Errorf("pkgmessage: rabbitmq bind: %w", err)
				}
			}
		}
	}

	autoAck := false
	exclusive := false
	noLocal := false
	noWait := false

	msgs, err := r.channel.Consume(queue, consumerTag, autoAck, exclusive, noLocal, noWait, nil)
	if err != nil {
		return nil, fmt.Errorf("pkgmessage: rabbitmq consume: %w", err)
	}
	return msgs, nil
}

func handleRabbitDelivery(ctx context.Context, d amqp.Delivery, handler Handler, autoAck bool) error {
	wrapped := newRabbitMessage(d)
	herr := handler(ctx, wrapped)

	if wrapped.hasResponded() || !autoAck {
		return nil
	}

	if herr == nil {
		return wrapped.Ack(ctx)
	}
	if err := wrapped.Nack(ctx); err != nil {
		return err
	}
	return nil
}

func resolveRabbitPublish(destination string, msg OutgoingMessage) (exchange string, routingKey string) {
	exchange = destination
	routingKey = msg.RoutingKey

	if msg.Metadata != nil {
		if ex, ok := msg.Metadata["exchange"].(string); ok && ex != "" {
			exchange = ex
		}
		if rk, ok := msg.Metadata["routing_key"].(string); ok && rk != "" {
			routingKey = rk
		}
	}

	if routingKey == "" {
		return "", destination
	}
	return exchange, routingKey
}

func headersToTable(headers []Header, attrs map[string]string) amqp.Table {
	if len(headers) == 0 && len(attrs) == 0 {
		return nil
	}
	t := amqp.Table{}
	for k, v := range attrs {
		if k == "" {
			continue
		}
		t[k] = v
	}
	for _, h := range headers {
		if h.Key == "" {
			continue
		}
		t[h.Key] = h.Value
	}
	return t
}
