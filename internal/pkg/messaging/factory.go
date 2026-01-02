package messaging

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	// DriverNSQ selects the NSQ backend.
	DriverNSQ = "nsq"
	// DriverNATS selects the NATS backend.
	DriverNATS = "nats"
	// DriverKafka selects the Kafka backend.
	DriverKafka = "kafka"
	// DriverGooglePubSub selects the Google Pub/Sub backend.
	DriverGooglePubSub = "google-pubsub"
)

// ErrUnknownDriver indicates an unsupported messaging driver.
var ErrUnknownDriver = errors.New("messaging: unknown driver")

// FactoryOptions groups config for supported messaging backends.
type FactoryOptions struct {
	// NSQ provides configuration for the NSQ driver.
	NSQ NSQConfig
	// Kafka provides configuration for the Kafka driver.
	Kafka KafkaConfig
	// NATS provides configuration for the NATS driver.
	NATS NATSConfig
	// PubSub provides configuration for the Google Pub/Sub driver.
	PubSub PubSubConfig
}

// NewFromDriver constructs a Messaging implementation by driver name.
func NewFromDriver(ctx context.Context, driver string, opts FactoryOptions) (Messaging, error) {
	switch strings.TrimSpace(driver) {
	case DriverNSQ:
		return NewNSQ(opts.NSQ)
	case DriverKafka:
		return NewKafka(opts.Kafka)
	case DriverNATS:
		return NewNATS(opts.NATS)
	case DriverGooglePubSub:
		return NewPubSub(ctx, opts.PubSub)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownDriver, driver)
	}
}
