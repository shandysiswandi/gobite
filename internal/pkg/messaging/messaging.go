package messaging

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrUnsupported is returned when a feature is not supported by the selected broker.
//
// For example, not all brokers support delayed delivery.
var ErrUnsupported = errors.New("pkgmessage: unsupported operation")

// Messaging is a broker-agnostic client that can publish and consume messages.
//
// Implementations can wrap Google Pub/Sub, NSQ, Kafka, NATS
// or any other messaging system.
type Messaging interface {
	io.Closer

	Publisher
	Consumer
}

// Publisher publishes messages to a destination (topic/subject/exchange/queue).
type Publisher interface {
	// Publish sends a message to the destination.
	Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error)
}

// Consumer consumes messages from a source (subscription/channel/queue/subject).
type Consumer interface {
	// Consume starts consuming messages from the source.
	Consume(ctx context.Context, source string, handler Handler, opts ...ConsumeOption) error
}

// Handler processes a received message.
//
// Returning a non-nil error does not imply any particular broker behavior.
// Implementations may choose to ack, nack/requeue, or leave the message unacked.
type Handler func(ctx context.Context, msg Message) error

// OutgoingMessage represents a broker-agnostic message to be published.
type OutgoingMessage struct {
	// Body is the message payload.
	Body []byte

	// Key is commonly used by Kafka for partitioning.
	Key []byte

	// Headers support arbitrary binary values and duplicate keys.
	Headers []Header

	// Attributes is a convenience for brokers that model string attributes (e.g. Pub/Sub).
	Attributes map[string]string

	// OrderingKey is commonly used by Google Pub/Sub.
	OrderingKey string

	// Delay is used for deferred delivery (when supported).
	Delay time.Duration

	// Metadata carries broker-specific publish settings (e.g. partition, message group id).
	Metadata map[string]any
}

// Header is a key/value pair used for message headers.
type Header struct {
	// Key is the header name.
	Key string
	// Value is the header value.
	Value []byte
}

// PublishResult carries optional broker-specific publish metadata.
type PublishResult struct {
	// MessageID is the broker-assigned message ID.
	MessageID string

	// Topic is the topic used for publishing (Kafka-like brokers).
	Topic string
	// Partition is the partition used for publishing (Kafka-like brokers).
	Partition int32
	// Offset is the publish offset (Kafka-like brokers).
	Offset int64

	// Sequence is commonly used by some NATS/JetStream APIs.
	Sequence uint64

	// Timestamp is when the broker accepted the message.
	Timestamp time.Time

	// Raw holds the underlying broker-specific publish result, if exposed.
	Raw any
}

// Message is a broker-agnostic received message.
type Message interface {
	// Body returns the message payload.
	Body() []byte
	// Key returns the message key.
	Key() []byte
	// Headers returns message headers.
	Headers() []Header
	// Attributes returns broker string attributes.
	Attributes() map[string]string

	// ID returns the broker message ID.
	ID() string
	// Topic returns the topic name when applicable.
	Topic() string
	// Subject returns the subject name when applicable.
	Subject() string
	// Timestamp returns the broker timestamp.
	Timestamp() time.Time

	// Ack acknowledges successful processing (delete/commit/ack).
	Ack(ctx context.Context) error
}

// Nackable can request a message redelivery (nack/requeue/negative ack).
type Nackable interface {
	// Nack requests a message redelivery.
	Nack(ctx context.Context) error
}

// Extendable can extend an ack deadline / lease when supported.
type Extendable interface {
	// Extend updates the message deadline/lease.
	Extend(ctx context.Context, d time.Duration) error
}

// MetadataCarrier exposes broker-specific metadata (delivery tags, receipt handles, etc).
type MetadataCarrier interface {
	// Metadata returns broker-specific metadata.
	Metadata() map[string]any
}

// RawCarrier exposes the underlying broker message type.
type RawCarrier interface {
	// Raw returns the underlying broker message type.
	Raw() any
}
