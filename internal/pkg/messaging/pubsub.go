package messaging

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
	// ErrPubSubProjectIDRequired is returned when a ProjectID is required but missing.
	ErrPubSubProjectIDRequired = errors.New("pkgmessage: pubsub project id is required")
	// ErrPubSubClientRequired is returned when the Pub/Sub client is nil or closed.
	ErrPubSubClientRequired = errors.New("pkgmessage: pubsub client is required")
	// ErrPubSubTopicRequired is returned when the publish topic is empty.
	ErrPubSubTopicRequired = errors.New("pkgmessage: pubsub topic is required")
	// ErrPubSubSubscriptionRequired is returned when the subscription name is empty.
	ErrPubSubSubscriptionRequired = errors.New("pkgmessage: pubsub subscription is required")
	// ErrPubSubHandlerRequired is returned when Consume is called with a nil handler.
	ErrPubSubHandlerRequired = errors.New("pkgmessage: pubsub handler is required")
)

// PubSubConfig configures the Google Pub/Sub implementation.
type PubSubConfig struct {
	// ProjectID is the Google Cloud project ID.
	ProjectID string

	// Client provides an existing Pub/Sub client.
	Client *pubsub.Client
	// ClientOptions are used when creating a new client.
	ClientOptions []option.ClientOption
}

// PubSub is a messaging implementation backed by Google Pub/Sub.
type PubSub struct {
	client *pubsub.Client

	mu     sync.Mutex
	closed bool

	publishers map[string]*pubsub.Publisher
}

// NewPubSub constructs a PubSub messaging client.
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

// Close stops publishers and closes the Pub/Sub client.
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

// Publish sends a message to a Pub/Sub topic.
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

// Consume starts consuming messages from a Pub/Sub subscription.
func (p *PubSub) Consume(ctx context.Context, source string, handler Handler, opts ...ConsumeOption) error {
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

	co := newConsumeOptions(opts...)
	topic := ""
	subscription := source
	if subName, ok := subscriptionFromConsumeOptions(co); ok {
		topic = source
		subscription = subName
	}

	sub := p.client.Subscriber(subscription)
	applyPubSubReceiveSettings(sub, co)

	autoAck := autoAckFromConsumeOptions(co)
	return sub.Receive(ctx, makePubSubHandler(topic, subscription, handler, autoAck))
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

func autoAckFromConsumeOptions(opts consumeOptions) bool {
	autoAck := opts.autoAck
	if opts.params == nil {
		return autoAck
	}
	v, ok := opts.params["auto_ack"]
	if !ok {
		return autoAck
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return autoAck
}

func subscriptionFromConsumeOptions(opts consumeOptions) (string, bool) {
	if opts.subscription != "" {
		return opts.subscription, true
	}
	if opts.params != nil {
		if v, ok := opts.params["subscription"]; ok && v != "" {
			return v, true
		}
	}
	return "", false
}

func makePubSubHandler(topic, subscription string, handler Handler, autoAck bool) func(context.Context, *pubsub.Message) {
	return func(ctx context.Context, m *pubsub.Message) {
		wrapped := newPubSubMessage(topic, subscription, m)
		herr := callHandlerWithRecover(ctx, "pubsub", func() error {
			return handler(ctx, wrapped)
		})

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

func applyPubSubReceiveSettings(sub *pubsub.Subscriber, opts consumeOptions) {
	if opts.concurrency > 0 {
		sub.ReceiveSettings.NumGoroutines = opts.concurrency
	}
	if opts.maxInFlight > 0 {
		sub.ReceiveSettings.MaxOutstandingMessages = opts.maxInFlight
	}
}
