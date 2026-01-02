package messaging

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	nsq "github.com/nsqio/go-nsq"
)

type nsqMessage struct {
	topic string
	msg   *nsq.Message

	responded atomic.Bool
}

func newNSQMessage(topic string, msg *nsq.Message) *nsqMessage {
	return &nsqMessage{
		topic: topic,
		msg:   msg,
	}
}

func (m *nsqMessage) hasResponded() bool {
	return m.responded.Load()
}

func (m *nsqMessage) Body() []byte { return m.msg.Body }
func (m *nsqMessage) Key() []byte  { return nil }

func (m *nsqMessage) Headers() []Header { return nil }

func (m *nsqMessage) Attributes() map[string]string { return nil }

func (m *nsqMessage) ID() string { return fmt.Sprintf("%x", m.msg.ID) }

func (m *nsqMessage) Topic() string   { return m.topic }
func (m *nsqMessage) Subject() string { return "" }

func (m *nsqMessage) Timestamp() time.Time {
	return time.Unix(0, m.msg.Timestamp)
}

func (m *nsqMessage) Ack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	m.msg.Finish()
	return nil
}

func (m *nsqMessage) Nack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	m.msg.Requeue(0)
	return nil
}

func (m *nsqMessage) Extend(ctx context.Context, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.msg.Touch()
	return nil
}

func (m *nsqMessage) Metadata() map[string]any {
	return map[string]any{
		"attempts":      m.msg.Attempts,
		"nsqd_address":  m.msg.NSQDAddress,
		"raw_timestamp": m.msg.Timestamp,
	}
}

func (m *nsqMessage) Raw() any { return m.msg }
