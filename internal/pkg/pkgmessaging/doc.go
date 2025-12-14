// Package pkgmessaging provides a broker-agnostic publish/consume API plus concrete
// adapters for multiple messaging systems (NSQ, NATS, Kafka, Google Pub/Sub).
//
// All adapters implement the Messaging interface:
//
//	type Messaging interface {
//		io.Closer
//		Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error)
//		Consume(ctx context.Context, source string, handler Handler, opts ...ConsumeOption) error
//	}
//
// Consume blocks until the context is canceled, so it's commonly started in a goroutine.
//
// # Acks
//
// When consuming, you can choose either:
//   - Auto-ack: pass WithAutoAck(true). If your handler returns nil => Ack; non-nil => Nack (when supported).
//   - Manual ack: omit WithAutoAck(true) and call msg.Ack(ctx) / msg.(Nackable).Nack(ctx) yourself.
//
// Note: ack/nack semantics differ per broker (e.g. NATS core messages don't support ack/nack; JetStream does).
//
// # NSQ
//
// Creating a client with a producer (nsqd TCP address, usually port 4150):
//
//	import (
//		"context"
//		"time"
//
//		"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
//	)
//
//	ctx := context.Background()
//
//	nsqClient, err := pkgmessaging.NewNSQ(pkgmessaging.NSQConfig{
//		ProducerAddr: "127.0.0.1:4150",
//	})
//	if err != nil {
//		// handle error
//	}
//	defer nsqClient.Close()
//
//	_, err = nsqClient.Publish(ctx, "user.created", pkgmessaging.OutgoingMessage{
//		Body: []byte(`{"id":"123"}`),
//	})
//
// Deferred publish (when you want a delivery delay):
//
//	_, err = nsqClient.Publish(ctx, "user.created", pkgmessaging.OutgoingMessage{
//		Body:  []byte(`{"id":"123"}`),
//		Delay: 10 * time.Second,
//	})
//
// Creating a client for consuming (use either lookupd or direct nsqd addresses):
//
//	nsqClient, err := pkgmessaging.NewNSQ(pkgmessaging.NSQConfig{
//		ConsumerLookupdAddr: []string{"127.0.0.1:4161"},
//		// Or:
//		// ConsumerNSQDAddrs: []string{"127.0.0.1:4150"},
//	})
//
// Consuming blocks until the context is canceled, so it's commonly started in a goroutine.
// To auto-acknowledge based on your handler's return value, pass WithAutoAck(true):
//
//	err = nsqClient.Consume(ctx, "user.created",
//		func(ctx context.Context, msg pkgmessaging.Message) error {
//			// ...process msg.Body()
//			return nil // nil => Ack, error => Nack (when WithAutoAck(true))
//		},
//		pkgmessaging.WithChannel("email-service"),
//		pkgmessaging.WithConcurrency(8),
//		pkgmessaging.WithAutoAck(true),
//	)
//
// If WithAutoAck(true) is not set, you must Ack/Nack manually (NSQ auto-responses are disabled):
//
//	err = nsqClient.Consume(ctx, "user.created",
//		func(ctx context.Context, msg pkgmessaging.Message) error {
//			if err := process(msg.Body()); err != nil {
//				if n, ok := msg.(pkgmessaging.Nackable); ok {
//					return n.Nack(ctx)
//				}
//				return err
//			}
//			return msg.Ack(ctx)
//		},
//		pkgmessaging.WithChannel("email-service"),
//	)
//
// # Kafka
//
// Creating a client with broker addresses:
//
//	import (
//		"context"
//
//		"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
//	)
//
//	ctx := context.Background()
//
//	kafkaClient, err := pkgmessaging.NewKafka(pkgmessaging.KafkaConfig{
//		Brokers: []string{"127.0.0.1:9092"},
//	})
//	if err != nil {
//		// handle error
//	}
//	defer kafkaClient.Close()
//
//	_, err = kafkaClient.Publish(ctx, "user.created", pkgmessaging.OutgoingMessage{
//		Key:  []byte("user:123"),
//		Body: []byte(`{"id":"123"}`),
//	})
//
// Consuming requires a consumer group (WithGroup). WithAutoAck(true) commits offsets
// when the handler returns nil; when the handler returns an error, offsets are not
// committed and the message may be received again:
//
//	err = kafkaClient.Consume(ctx, "user.created",
//		func(ctx context.Context, msg pkgmessaging.Message) error {
//			// ...process msg.Body()
//			return nil // nil => Ack (commit), error => Nack (no commit) when WithAutoAck(true)
//		},
//		pkgmessaging.WithGroup("email-service"),
//		pkgmessaging.WithConcurrency(8),
//		pkgmessaging.WithAutoAck(true),
//	)
//
// If WithAutoAck(true) is not set, you must Ack/Nack manually:
//
//	err = kafkaClient.Consume(ctx, "user.created",
//		func(ctx context.Context, msg pkgmessaging.Message) error {
//			if err := process(msg.Body()); err != nil {
//				return msg.Nack(ctx)
//			}
//			return msg.Ack(ctx)
//		},
//		pkgmessaging.WithGroup("email-service"),
//	)
//
// # NATS
//
// Creating a client:
//
//	natsClient, err := pkgmessaging.NewNATS(pkgmessaging.NATSConfig{
//		URL: "nats://127.0.0.1:4222",
//	})
//	if err != nil {
//		// handle error
//	}
//	defer natsClient.Close()
//
// Publish to a subject:
//
//	_, err = natsClient.Publish(ctx, "user.created", pkgmessaging.OutgoingMessage{
//		Body: []byte(`{"id":"123"}`),
//	})
//
// Consume from a subject using a queue group (competing consumers). WithQueueGroup controls delivery
// semantics (one message per group member):
//
//	err = natsClient.Consume(ctx, "user.created",
//		func(ctx context.Context, msg pkgmessaging.Message) error {
//			// ...process msg.Body()
//			return nil
//		},
//		pkgmessaging.WithQueueGroup("email-service"),
//		pkgmessaging.WithConcurrency(8),
//		pkgmessaging.WithAutoAck(true),
//	)
//
// Note: Ack/Nack are meaningful for JetStream messages; for core NATS messages they are effectively no-ops.
//
// # Google Pub/Sub
//
// Creating a client:
//
//	pubsubClient, err := pkgmessaging.NewPubSub(ctx, pkgmessaging.PubSubConfig{
//		ProjectID: "my-gcp-project",
//	})
//	if err != nil {
//		// handle error
//	}
//	defer pubsubClient.Close()
//
// Publish to a topic:
//
//	_, err = pubsubClient.Publish(ctx, "user.created", pkgmessaging.OutgoingMessage{
//		Body: []byte(`{"id":"123"}`),
//	})
//
// Consume from a subscription (the subscription must already exist in GCP):
//
//	err = pubsubClient.Consume(ctx, "user-created-email-sub",
//		func(ctx context.Context, msg pkgmessaging.Message) error {
//			// ...process msg.Body()
//			return nil
//		},
//		pkgmessaging.WithConcurrency(8),
//		pkgmessaging.WithMaxInFlight(100),
//		pkgmessaging.WithAutoAck(true),
//	)
package pkgmessaging
