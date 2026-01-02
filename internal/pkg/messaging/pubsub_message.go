package messaging

import (
	"context"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

type pubSubMessage struct {
	topic        string
	subscription string
	msg          *pubsub.Message

	responded atomic.Bool
}

func newPubSubMessage(topic, subscription string, msg *pubsub.Message) *pubSubMessage {
	return &pubSubMessage{
		topic:        topic,
		subscription: subscription,
		msg:          msg,
	}
}

func (m *pubSubMessage) hasResponded() bool {
	return m.responded.Load()
}

func (m *pubSubMessage) Body() []byte { return m.msg.Data }
func (m *pubSubMessage) Key() []byte  { return nil }

func (m *pubSubMessage) Headers() []Header { return nil }

func (m *pubSubMessage) Attributes() map[string]string { return m.msg.Attributes }

func (m *pubSubMessage) ID() string { return m.msg.ID }

func (m *pubSubMessage) Topic() string   { return m.topic }
func (m *pubSubMessage) Subject() string { return "" }

func (m *pubSubMessage) Timestamp() time.Time { return m.msg.PublishTime }

func (m *pubSubMessage) Ack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	m.msg.Ack()
	return nil
}

func (m *pubSubMessage) Nack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	m.msg.Nack()
	return nil
}

func (m *pubSubMessage) Extend(ctx context.Context, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrUnsupported
}

func (m *pubSubMessage) Metadata() map[string]any {
	meta := map[string]any{
		"topic":        m.topic,
		"subscription": m.subscription,
		"ordering_key": m.msg.OrderingKey,
	}
	if m.msg.DeliveryAttempt != nil {
		meta["delivery_attempt"] = *m.msg.DeliveryAttempt
	}
	return meta
}

func (m *pubSubMessage) Raw() any { return m.msg }
