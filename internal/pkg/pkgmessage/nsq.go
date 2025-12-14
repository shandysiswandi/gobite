package pkgmessage

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
	ErrNSQTopicRequired         = errors.New("pkgmessage: nsq topic is required")
	ErrNSQChannelRequired       = errors.New("pkgmessage: nsq channel is required")
	ErrNSQHandlerRequired       = errors.New("pkgmessage: nsq handler is required")
	ErrNSQProducerAddrRequired  = errors.New("pkgmessage: nsq producer address is required")
	ErrNSQConsumerAddrsRequired = errors.New("pkgmessage: nsq consumer nsqd/lookupd addresses are required")
)

type NSQConfig struct {
	ProducerAddr string

	ConsumerNSQDAddrs   []string
	ConsumerLookupdAddr []string

	ProducerConfig *nsq.Config
	ConsumerConfig *nsq.Config
}

type NSQ struct {
	producer *nsq.Producer

	consumerNSQDAddrs   []string
	consumerLookupdAddr []string
	consumerConfig      *nsq.Config

	mu        sync.Mutex
	consumers []*nsq.Consumer
	closed    bool
}

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
		producer = p
	}

	ccfg := cfg.ConsumerConfig
	if ccfg == nil {
		ccfg = nsq.NewConfig()
	}

	return &NSQ{
		producer: producer,

		consumerNSQDAddrs:   append([]string{}, cfg.ConsumerNSQDAddrs...),
		consumerLookupdAddr: append([]string{}, cfg.ConsumerLookupdAddr...),
		consumerConfig:      ccfg,
	}, nil
}

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

func (n *NSQ) Consume(ctx context.Context, source string, handler Handler, opts ConsumeOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if source == "" {
		return ErrNSQTopicRequired
	}
	if handler == nil {
		return ErrNSQHandlerRequired
	}
	if len(n.consumerNSQDAddrs) == 0 && len(n.consumerLookupdAddr) == 0 {
		return ErrNSQConsumerAddrsRequired
	}

	consumer, concurrency, autoAck, err := n.newNSQConsumer(source, opts)
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

func (n *NSQ) newNSQConsumer(topic string, opts ConsumeOptions) (*nsq.Consumer, int, bool, error) {
	channel := opts.Channel
	if channel == "" {
		return nil, 0, false, ErrNSQChannelRequired
	}

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	autoAck := opts.AutoAck
	if opts.Params != nil {
		if v, ok := opts.Params["auto_ack"]; ok {
			if b, err := strconv.ParseBool(v); err == nil {
				autoAck = b
			}
		}
	}

	ccfg := *n.consumerConfig
	if opts.MaxInFlight > 0 {
		ccfg.MaxInFlight = opts.MaxInFlight
	} else if ccfg.MaxInFlight < concurrency {
		ccfg.MaxInFlight = concurrency
	}

	consumer, err := nsq.NewConsumer(topic, channel, &ccfg)
	if err != nil {
		return nil, 0, false, fmt.Errorf("pkgmessage: nsq new consumer: %w", err)
	}

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
	if len(n.consumerLookupdAddr) > 0 {
		if err := consumer.ConnectToNSQLookupds(n.consumerLookupdAddr); err != nil {
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
		herr := handler(ctx, wrapped)

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
