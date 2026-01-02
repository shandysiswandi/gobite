package messaging

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
)

type natsMessage struct {
	msg        *nats.Msg
	receivedAt time.Time

	responded atomic.Bool
}

func newNATSMessage(msg *nats.Msg, receivedAt time.Time) *natsMessage {
	return &natsMessage{
		msg:        msg,
		receivedAt: receivedAt,
	}
}

func (m *natsMessage) hasResponded() bool {
	return m.responded.Load()
}

func (m *natsMessage) Body() []byte { return m.msg.Data }
func (m *natsMessage) Key() []byte  { return nil }

func (m *natsMessage) Headers() []Header {
	if len(m.msg.Header) == 0 {
		return nil
	}

	var headers []Header
	for k, values := range m.msg.Header {
		for _, v := range values {
			headers = append(headers, Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}
	return headers
}

func (m *natsMessage) Attributes() map[string]string {
	if len(m.msg.Header) == 0 {
		return nil
	}

	attrs := make(map[string]string, len(m.msg.Header))
	for k, values := range m.msg.Header {
		if len(values) > 0 {
			attrs[k] = values[0]
		}
	}
	return attrs
}

func (m *natsMessage) ID() string { return "" }

func (m *natsMessage) Topic() string   { return "" }
func (m *natsMessage) Subject() string { return m.msg.Subject }

func (m *natsMessage) Timestamp() time.Time { return m.receivedAt }

func (m *natsMessage) Ack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	if err := m.msg.Ack(); err != nil && !isNATSAckUnsupported(err) {
		return err
	}
	return nil
}

func (m *natsMessage) Nack(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.responded.Swap(true) {
		return nil
	}
	if err := m.msg.Nak(); err != nil && !isNATSAckUnsupported(err) {
		return err
	}
	return nil
}

func (m *natsMessage) Extend(ctx context.Context, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := m.msg.InProgress(); err != nil && !isNATSAckUnsupported(err) {
		return err
	}
	return nil
}

func (m *natsMessage) Metadata() map[string]any {
	meta := map[string]any{
		"reply": m.msg.Reply,
	}

	if md, err := m.msg.Metadata(); err == nil && md != nil {
		meta["sequence_stream"] = md.Sequence.Stream
		meta["sequence_consumer"] = md.Sequence.Consumer
		meta["num_delivered"] = md.NumDelivered
		meta["num_pending"] = md.NumPending
		meta["timestamp"] = md.Timestamp
		meta["domain"] = md.Domain
	}

	return meta
}

func (m *natsMessage) Raw() any { return m.msg }

func (m *natsMessage) String() string {
	return fmt.Sprintf("nats subject=%q", m.msg.Subject)
}

func isNATSAckUnsupported(err error) bool {
	return errors.Is(err, nats.ErrMsgNoReply) || errors.Is(err, nats.ErrMsgNotBound)
}
