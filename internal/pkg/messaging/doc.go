// Package messaging provides a broker-agnostic API for publishing and
// consuming messages.
//
// The goal is to keep business code independent from the underlying messaging
// system (Kafka, NATS, NSQ, Google Pub/Sub, etc). You can swap implementations
// without changing your use-case code, as long as it relies on the interfaces
// in this package.
package messaging
