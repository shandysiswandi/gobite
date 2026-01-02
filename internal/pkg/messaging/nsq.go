package messaging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	nsq "github.com/nsqio/go-nsq"
)

var (
	// ErrNSQTopicRequired is returned when the topic is empty.
	ErrNSQTopicRequired = errors.New("pkgmessage: nsq topic is required")
	// ErrNSQChannelRequired is returned when the channel is empty.
	ErrNSQChannelRequired = errors.New("pkgmessage: nsq channel is required")
	// ErrNSQHandlerRequired is returned when Consume is called with a nil handler.
	ErrNSQHandlerRequired = errors.New("pkgmessage: nsq handler is required")
	// ErrNSQProducerAddrRequired is returned when the producer address is missing.
	ErrNSQProducerAddrRequired = errors.New("pkgmessage: nsq producer address is required")
	// ErrNSQConsumerAddrsRequired is returned when no NSQD/lookupd consumer addresses are configured.
	ErrNSQConsumerAddrsRequired = errors.New("pkgmessage: nsq consumer nsqd/lookupd addresses are required")
)

// NSQConfig configures the NSQ implementation.
type NSQConfig struct {
	// ProducerAddr is the NSQD address for publishing.
	ProducerAddr string

	// ConsumerNSQDAddrs lists NSQD addresses for consumers.
	ConsumerNSQDAddrs []string
	// ConsumerLookupdAddrs lists lookupd addresses for consumers.
	ConsumerLookupdAddrs []string

	// ProducerConfig overrides the default producer config.
	ProducerConfig *nsq.Config
	// ConsumerConfig overrides the default consumer config.
	ConsumerConfig *nsq.Config
}

// NSQ is a messaging implementation backed by NSQ.
type NSQ struct {
	producer *nsq.Producer

	consumerNSQDAddrs    []string
	consumerLookupdAddrs []string
	consumerConfig       *nsq.Config

	mu        sync.Mutex
	consumers []*nsq.Consumer
	closed    bool
}

// NewNSQ constructs an NSQ messaging client.
func NewNSQ(cfg NSQConfig) (*NSQ, error) {
	var producer *nsq.Producer
	if cfg.ProducerAddr != "" {
		pcfg := cfg.ProducerConfig
		if pcfg == nil {
			pcfg = nsq.NewConfig()
		}

		p, err := nsq.NewProducer(cfg.ProducerAddr, pcfg)
		if err != nil {
			return nil, fmt.Errorf("pkgmessage: nsq new producer: %w", err)
		}
		p.SetLoggerLevel(nsq.LogLevelError)

		producer = p
	}

	ccfg := cfg.ConsumerConfig
	if ccfg == nil {
		ccfg = nsq.NewConfig()
	}

	return &NSQ{
		producer: producer,

		consumerNSQDAddrs:    append([]string{}, cfg.ConsumerNSQDAddrs...),
		consumerLookupdAddrs: append([]string{}, cfg.ConsumerLookupdAddrs...),
		consumerConfig:       ccfg,
	}, nil
}

// Close stops NSQ consumers and the producer.
func (n *NSQ) Close() error {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return nil
	}
	n.closed = true
	consumers := append([]*nsq.Consumer{}, n.consumers...)
	n.mu.Unlock()

	for _, c := range consumers {
		c.Stop()
		<-c.StopChan
	}

	if n.producer != nil {
		n.producer.Stop()
	}
	return nil
}

// Publish sends a message to an NSQ topic.
func (n *NSQ) Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if destination == "" {
		return PublishResult{}, ErrNSQTopicRequired
	}
	if n.producer == nil {
		return PublishResult{}, ErrNSQProducerAddrRequired
	}

	body := msg.Body
	if msg.Delay > 0 {
		if err := n.producer.DeferredPublish(destination, msg.Delay, body); err != nil {
			return PublishResult{}, fmt.Errorf("pkgmessage: nsq deferred publish: %w", err)
		}
		return PublishResult{
			Topic:     destination,
			Timestamp: time.Now(),
		}, nil
	}

	if err := n.producer.Publish(destination, body); err != nil {
		return PublishResult{}, fmt.Errorf("pkgmessage: nsq publish: %w", err)
	}

	return PublishResult{
		Topic:     destination,
		Timestamp: time.Now(),
	}, nil
}

// Consume starts consuming messages from an NSQ topic/channel.
func (n *NSQ) Consume(ctx context.Context, source string, handler Handler, opts ...ConsumeOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if source == "" {
		return ErrNSQTopicRequired
	}
	if handler == nil {
		return ErrNSQHandlerRequired
	}
	if len(n.consumerNSQDAddrs) == 0 && len(n.consumerLookupdAddrs) == 0 {
		return ErrNSQConsumerAddrsRequired
	}

	co := newConsumeOptions(opts...)
	consumer, concurrency, autoAck, err := n.newNSQConsumer(source, co)
	if err != nil {
		return err
	}

	consumer.AddConcurrentHandlers(n.makeNSQHandler(ctx, source, handler, autoAck), concurrency)

	if err := n.addNSQConsumer(consumer); err != nil {
		stopNSQConsumer(consumer)
		return err
	}

	if err := n.connectNSQConsumer(consumer); err != nil {
		stopNSQConsumer(consumer)
		return err
	}

	return waitNSQConsumer(ctx, consumer)
}

func (n *NSQ) newNSQConsumer(topic string, opts consumeOptions) (*nsq.Consumer, int, bool, error) {
	channel := opts.channel
	if channel == "" {
		return nil, 0, false, ErrNSQChannelRequired
	}

	concurrency := opts.concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	autoAck := opts.autoAck
	if opts.params != nil {
		if v, ok := opts.params["auto_ack"]; ok {
			if b, err := strconv.ParseBool(v); err == nil {
				autoAck = b
			}
		}
	}

	ccfg := *n.consumerConfig
	if opts.maxInFlight > 0 {
		ccfg.MaxInFlight = opts.maxInFlight
	} else if ccfg.MaxInFlight < concurrency {
		ccfg.MaxInFlight = concurrency
	}

	consumer, err := nsq.NewConsumer(topic, channel, &ccfg)
	if err != nil {
		return nil, 0, false, fmt.Errorf("pkgmessage: nsq new consumer: %w", err)
	}
	consumer.SetLoggerLevel(nsq.LogLevelError)

	return consumer, concurrency, autoAck, nil
}

func (n *NSQ) addNSQConsumer(consumer *nsq.Consumer) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return io.ErrClosedPipe
	}
	n.consumers = append(n.consumers, consumer)
	return nil
}

func (n *NSQ) connectNSQConsumer(consumer *nsq.Consumer) error {
	if len(n.consumerLookupdAddrs) > 0 {
		if err := consumer.ConnectToNSQLookupds(n.consumerLookupdAddrs); err != nil {
			return fmt.Errorf("pkgmessage: nsq connect lookupd: %w", err)
		}
		return nil
	}

	if err := consumer.ConnectToNSQDs(n.consumerNSQDAddrs); err != nil {
		return fmt.Errorf("pkgmessage: nsq connect nsqd: %w", err)
	}
	return nil
}

func (n *NSQ) makeNSQHandler(ctx context.Context, topic string, handler Handler, autoAck bool) nsq.HandlerFunc {
	return func(m *nsq.Message) error {
		m.DisableAutoResponse()

		wrapped := newNSQMessage(topic, m)
		herr := callHandlerWithRecover(ctx, "nsq", func() error {
			return handler(ctx, wrapped)
		})

		if wrapped.hasResponded() || !autoAck {
			return herr
		}

		if herr == nil {
			if err := wrapped.Ack(ctx); err != nil {
				return err
			}
			return nil
		}

		if err := wrapped.Nack(ctx); err != nil {
			return err
		}
		return nil
	}
}

func stopNSQConsumer(consumer *nsq.Consumer) {
	consumer.Stop()
	<-consumer.StopChan
}

func waitNSQConsumer(ctx context.Context, consumer *nsq.Consumer) error {
	select {
	case <-ctx.Done():
		stopNSQConsumer(consumer)
		return ctx.Err()
	case <-consumer.StopChan:
		return nil
	}
}
