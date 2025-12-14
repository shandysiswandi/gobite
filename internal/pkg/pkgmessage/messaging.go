package pkgmessage

import (
	"context"
	"errors"
	"io"
	"time"
)

var ErrUnsupported = errors.New("pkgmessage: unsupported operation")

// Messaging is a broker-agnostic client that can publish and consume messages.
//
// Implementations can wrap Google Pub/Sub, NSQ, Kafka, RabbitMQ, NATS
// or any other messaging system.
type Messaging interface {
	io.Closer

	Publisher
	Consumer
}

// Publisher publishes messages to a destination (topic/subject/exchange/queue).
type Publisher interface {
	Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error)
}

// Consumer consumes messages from a source (subscription/channel/queue/subject).
type Consumer interface {
	Consume(ctx context.Context, source string, handler Handler, opts ConsumeOptions) error
}

// Handler processes a received message.
//
// Returning a non-nil error does not imply any particular broker behavior.
// Implementations may choose to ack, nack/requeue, or leave the message unacked.
type Handler func(ctx context.Context, msg Message) error

// ConsumeOptions provides common consumption settings across brokers.
// Fields that do not apply to a given broker should be ignored.
type ConsumeOptions struct {
	Concurrency int

	// Group is commonly used for Kafka consumer groups.
	Group string

	// Channel is commonly used for NSQ channels.
	Channel string

	// QueueGroup is commonly used for NATS queue subscriptions.
	QueueGroup string

	// Exchange and RoutingKey are commonly used for RabbitMQ bindings.
	Exchange   string
	RoutingKey string

	// AutoAck indicates messages are acknowledged automatically by the broker/client.
	AutoAck bool

	// MaxInFlight is a generic "outstanding/unacked messages" limit.
	MaxInFlight int

	// AckDeadline is commonly used for Google Pub/Sub.
	AckDeadline time.Duration

	// Params carries broker-specific settings (e.g. "auto_commit", "prefetch").
	Params map[string]string
}

// OutgoingMessage represents a broker-agnostic message to be published.
type OutgoingMessage struct {
	Body []byte

	// Key is commonly used by Kafka for partitioning.
	Key []byte

	// Headers support arbitrary binary values and duplicate keys.
	Headers []Header

	// Attributes is a convenience for brokers that model string attributes (e.g. Pub/Sub).
	Attributes map[string]string

	// OrderingKey is commonly used by Google Pub/Sub.
	OrderingKey string

	// RoutingKey is commonly used by RabbitMQ.
	RoutingKey string

	// Delay is used for deferred delivery (when supported).
	Delay time.Duration

	// Metadata carries broker-specific publish settings (e.g. partition, message group id).
	Metadata map[string]any
}

// Header is a key/value pair used for message headers.
type Header struct {
	Key   string
	Value []byte
}

// PublishResult carries optional broker-specific publish metadata.
type PublishResult struct {
	MessageID string

	// Kafka-like metadata.
	Topic     string
	Partition int32
	Offset    int64

	// Sequence is commonly used by some NATS/JetStream APIs.
	Sequence uint64

	Timestamp time.Time

	// Raw holds the underlying broker-specific publish result, if exposed.
	Raw any
}

// Message is a broker-agnostic received message.
type Message interface {
	Body() []byte
	Key() []byte
	Headers() []Header
	Attributes() map[string]string

	ID() string
	Topic() string
	Subject() string
	Timestamp() time.Time

	// Ack acknowledges successful processing (delete/commit/ack).
	Ack(ctx context.Context) error
}

// Nackable can request a message redelivery (nack/requeue/negative ack).
type Nackable interface {
	Nack(ctx context.Context) error
}

// Extendable can extend an ack deadline / lease when supported.
type Extendable interface {
	Extend(ctx context.Context, d time.Duration) error
}

// MetadataCarrier exposes broker-specific metadata (delivery tags, receipt handles, etc).
type MetadataCarrier interface {
	Metadata() map[string]any
}

// RawCarrier exposes the underlying broker message type.
type RawCarrier interface {
	Raw() any
}
