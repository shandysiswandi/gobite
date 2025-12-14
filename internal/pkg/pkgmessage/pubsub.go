package pkgmessage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	"google.golang.org/api/option"
)

var (
	ErrPubSubProjectIDRequired    = errors.New("pkgmessage: pubsub project id is required")
	ErrPubSubClientRequired       = errors.New("pkgmessage: pubsub client is required")
	ErrPubSubTopicRequired        = errors.New("pkgmessage: pubsub topic is required")
	ErrPubSubSubscriptionRequired = errors.New("pkgmessage: pubsub subscription is required")
	ErrPubSubHandlerRequired      = errors.New("pkgmessage: pubsub handler is required")
)

type PubSubConfig struct {
	ProjectID string

	Client        *pubsub.Client
	ClientOptions []option.ClientOption
}

type PubSub struct {
	client *pubsub.Client

	mu     sync.Mutex
	closed bool

	publishers map[string]*pubsub.Publisher
}

func NewPubSub(ctx context.Context, cfg PubSubConfig) (*PubSub, error) {
	if cfg.Client != nil {
		return &PubSub{client: cfg.Client, publishers: map[string]*pubsub.Publisher{}}, nil
	}
	if cfg.ProjectID == "" {
		return nil, ErrPubSubProjectIDRequired
	}

	c, err := pubsub.NewClient(ctx, cfg.ProjectID, cfg.ClientOptions...)
	if err != nil {
		return nil, fmt.Errorf("pkgmessage: pubsub new client: %w", err)
	}

	return &PubSub{client: c, publishers: map[string]*pubsub.Publisher{}}, nil
}

func (p *PubSub) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	pubs := make([]*pubsub.Publisher, 0, len(p.publishers))
	for _, pub := range p.publishers {
		pubs = append(pubs, pub)
	}
	p.publishers = nil
	p.mu.Unlock()

	for _, pub := range pubs {
		pub.Stop()
	}

	if p.client == nil {
		return nil
	}
	return p.client.Close()
}

func (p *PubSub) Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if destination == "" {
		return PublishResult{}, ErrPubSubTopicRequired
	}
	if err := p.ensurePubSubOpen(); err != nil {
		return PublishResult{}, err
	}
	if msg.Delay > 0 {
		return PublishResult{}, ErrUnsupported
	}

	pub := p.getPublisher(destination)
	res := pub.Publish(ctx, &pubsub.Message{
		Data:        msg.Body,
		Attributes:  msg.Attributes,
		OrderingKey: msg.OrderingKey,
	})
	id, err := res.Get(ctx)
	if err != nil {
		return PublishResult{}, fmt.Errorf("pkgmessage: pubsub publish: %w", err)
	}

	return PublishResult{
		MessageID: id,
		Topic:     destination,
	}, nil
}

func (p *PubSub) Consume(ctx context.Context, source string, handler Handler, opts ConsumeOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if source == "" {
		return ErrPubSubSubscriptionRequired
	}
	if handler == nil {
		return ErrPubSubHandlerRequired
	}
	if err := p.ensurePubSubOpen(); err != nil {
		return err
	}

	sub := p.client.Subscriber(source)
	applyPubSubReceiveSettings(sub, opts)

	autoAck := autoAckFromConsumeOptions(opts)
	return sub.Receive(ctx, makePubSubHandler(source, handler, autoAck))
}

func (p *PubSub) getPublisher(topicNameOrID string) *pubsub.Publisher {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.publishers == nil {
		p.publishers = map[string]*pubsub.Publisher{}
	}
	if pub, ok := p.publishers[topicNameOrID]; ok {
		return pub
	}
	pub := p.client.Publisher(topicNameOrID)
	p.publishers[topicNameOrID] = pub
	return pub
}

func (p *PubSub) ensurePubSubOpen() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return ErrPubSubClientRequired
	}
	if p.closed {
		return io.ErrClosedPipe
	}
	return nil
}

func autoAckFromConsumeOptions(opts ConsumeOptions) bool {
	autoAck := opts.AutoAck
	if opts.Params == nil {
		return autoAck
	}
	v, ok := opts.Params["auto_ack"]
	if !ok {
		return autoAck
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return autoAck
}

func makePubSubHandler(subscription string, handler Handler, autoAck bool) func(context.Context, *pubsub.Message) {
	return func(ctx context.Context, m *pubsub.Message) {
		wrapped := newPubSubMessage(subscription, m)
		herr := handler(ctx, wrapped)

		if wrapped.hasResponded() || !autoAck {
			return
		}

		if herr == nil {
			if err := wrapped.Ack(ctx); err != nil {
				return
			}
			return
		}

		if err := wrapped.Nack(ctx); err != nil {
			return
		}
	}
}

func applyPubSubReceiveSettings(sub *pubsub.Subscriber, opts ConsumeOptions) {
	if opts.Concurrency > 0 {
		sub.ReceiveSettings.NumGoroutines = opts.Concurrency
	}
	if opts.MaxInFlight > 0 {
		sub.ReceiveSettings.MaxOutstandingMessages = opts.MaxInFlight
	}
}
