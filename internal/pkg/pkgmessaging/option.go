package pkgmessaging

import "time"

type consumeOptions struct {
	// concurrency specifies the number of concurrent message handlers
	// processing messages in parallel.
	concurrency int

	// autoAck indicates whether messages are acknowledged automatically
	// by the broker or client upon delivery.
	autoAck bool

	// group identifies the consumer group name.
	// Commonly used for Kafka consumer groups.
	group string

	// channel specifies the channel name.
	// Commonly used for NSQ consumers.
	channel string

	// queueGroup specifies the queue group name.
	// Commonly used for NATS queue subscriptions.
	queueGroup string

	// maxInFlight limits the maximum number of outstanding (unacknowledged)
	// messages that can be in flight at any given time.
	maxInFlight int

	// ackDeadline defines the maximum amount of time the consumer has
	// to acknowledge a message before it is redelivered.
	// Commonly used for Google Pub/Sub.
	ackDeadline time.Duration

	// params contains broker-specific configuration options such as
	// "auto_commit", "prefetch", or other implementation-defined settings.
	params map[string]string
}

type ConsumeOption func(*consumeOptions)

func newConsumeOptions(opts ...ConsumeOption) consumeOptions {
	var co consumeOptions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&co)
	}
	return co
}

func WithConcurrency(n int) ConsumeOption {
	return func(o *consumeOptions) { o.concurrency = n }
}

func WithGroup(group string) ConsumeOption {
	return func(o *consumeOptions) { o.group = group }
}

func WithChannel(channel string) ConsumeOption {
	return func(o *consumeOptions) { o.channel = channel }
}

func WithQueueGroup(queueGroup string) ConsumeOption {
	return func(o *consumeOptions) { o.queueGroup = queueGroup }
}

func WithAutoAck(autoAck bool) ConsumeOption {
	return func(o *consumeOptions) { o.autoAck = autoAck }
}

// WithAutoAct is an alias for WithAutoAck.
func WithAutoAct(autoAck bool) ConsumeOption {
	return WithAutoAck(autoAck)
}

func WithMaxInFlight(maxInFlight int) ConsumeOption {
	return func(o *consumeOptions) { o.maxInFlight = maxInFlight }
}

func WithAckDeadline(d time.Duration) ConsumeOption {
	return func(o *consumeOptions) { o.ackDeadline = d }
}

func WithParams(params map[string]string) ConsumeOption {
	return func(o *consumeOptions) {
		if len(params) == 0 {
			return
		}
		if o.params == nil {
			o.params = make(map[string]string, len(params))
		}
		for k, v := range params {
			o.params[k] = v
		}
	}
}

func WithParam(key, value string) ConsumeOption {
	return func(o *consumeOptions) {
		if key == "" {
			return
		}
		if o.params == nil {
			o.params = make(map[string]string, 1)
		}
		o.params[key] = value
	}
}
