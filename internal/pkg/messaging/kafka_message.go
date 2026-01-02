package messaging

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaMessage struct {
	reader *kafka.Reader
	msg    kafka.Message

	responded atomic.Bool
}

func newKafkaMessage(reader *kafka.Reader, msg kafka.Message) *kafkaMessage {
	return &kafkaMessage{
		reader: reader,
		msg:    msg,
	}
}

func (m *kafkaMessage) hasResponded() bool {
	return m.responded.Load()
}

func (m *kafkaMessage) Body() []byte { return m.msg.Value }
func (m *kafkaMessage) Key() []byte  { return m.msg.Key }

func (m *kafkaMessage) Headers() []Header {
	if len(m.msg.Headers) == 0 {
		return nil
	}
	out := make([]Header, 0, len(m.msg.Headers))
	for _, h := range m.msg.Headers {
		out = append(out, Header{Key: h.Key, Value: h.Value})
	}
	return out
}

func (m *kafkaMessage) Attributes() map[string]string {
	if len(m.msg.Headers) == 0 {
		return nil
	}
	attrs := make(map[string]string, len(m.msg.Headers))
	for _, h := range m.msg.Headers {
		if _, ok := attrs[h.Key]; ok {
			continue
		}
		attrs[h.Key] = string(h.Value)
	}
	return attrs
}

func (m *kafkaMessage) ID() string {
	return fmt.Sprintf("%s/%d/%d", m.msg.Topic, m.msg.Partition, m.msg.Offset)
}

func (m *kafkaMessage) Topic() string   { return m.msg.Topic }
func (m *kafkaMessage) Subject() string { return "" }

func (m *kafkaMessage) Timestamp() time.Time { return m.msg.Time }

func (m *kafkaMessage) Ack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	return m.reader.CommitMessages(ctx, m.msg)
}

func (m *kafkaMessage) Nack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.responded.Store(true)
	return nil
}

func (m *kafkaMessage) Extend(ctx context.Context, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrUnsupported
}

func (m *kafkaMessage) Metadata() map[string]any {
	return map[string]any{
		"partition": m.msg.Partition,
		"offset":    m.msg.Offset,
		"topic":     m.msg.Topic,
	}
}

func (m *kafkaMessage) Raw() any { return m.msg }
