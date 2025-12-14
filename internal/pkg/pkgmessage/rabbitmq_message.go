package pkgmessage

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMessage struct {
	delivery amqp.Delivery

	responded atomic.Bool
}

func newRabbitMessage(d amqp.Delivery) *rabbitMessage {
	return &rabbitMessage{delivery: d}
}

func (m *rabbitMessage) hasResponded() bool {
	return m.responded.Load()
}

func (m *rabbitMessage) Body() []byte { return m.delivery.Body }
func (m *rabbitMessage) Key() []byte  { return nil }

func (m *rabbitMessage) Headers() []Header {
	if len(m.delivery.Headers) == 0 {
		return nil
	}
	out := make([]Header, 0, len(m.delivery.Headers))
	for k, v := range m.delivery.Headers {
		out = append(out, Header{
			Key:   k,
			Value: tableValueToBytes(v),
		})
	}
	return out
}

func (m *rabbitMessage) Attributes() map[string]string {
	if len(m.delivery.Headers) == 0 {
		return nil
	}
	attrs := make(map[string]string, len(m.delivery.Headers))
	for k, v := range m.delivery.Headers {
		switch vv := v.(type) {
		case string:
			attrs[k] = vv
		case []byte:
			attrs[k] = string(vv)
		}
	}
	return attrs
}

func (m *rabbitMessage) ID() string {
	if m.delivery.MessageId != "" {
		return m.delivery.MessageId
	}
	return fmt.Sprintf("%d", m.delivery.DeliveryTag)
}

func (m *rabbitMessage) Topic() string   { return m.delivery.Exchange }
func (m *rabbitMessage) Subject() string { return m.delivery.RoutingKey }

func (m *rabbitMessage) Timestamp() time.Time { return m.delivery.Timestamp }

func (m *rabbitMessage) Ack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	return m.delivery.Ack(false)
}

func (m *rabbitMessage) Nack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	return m.delivery.Nack(false, true)
}

func (m *rabbitMessage) Extend(ctx context.Context, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrUnsupported
}

func (m *rabbitMessage) Metadata() map[string]any {
	return map[string]any{
		"delivery_tag": m.delivery.DeliveryTag,
		"redelivered":  m.delivery.Redelivered,
		"consumer_tag": m.delivery.ConsumerTag,
		"exchange":     m.delivery.Exchange,
		"routing_key":  m.delivery.RoutingKey,
		"content_type": m.delivery.ContentType,
		"app_id":       m.delivery.AppId,
		"type":         m.delivery.Type,
		"priority":     m.delivery.Priority,
	}
}

func (m *rabbitMessage) Raw() any { return m.delivery }

func tableValueToBytes(v any) []byte {
	switch vv := v.(type) {
	case []byte:
		return vv
	case string:
		return []byte(vv)
	default:
		return []byte(fmt.Sprint(v))
	}
}
