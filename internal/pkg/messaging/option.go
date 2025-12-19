package messaging

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

	// subscription specifies the subscription name.
	// Commonly used for Google Pub/Sub consumers.
	subscription string

	// maxInFlight limits the maximum number of outstanding (unacknowledged)
	// messages that can be in flight at any given time.
	maxInFlight int

	// params contains broker-specific configuration options such as
	// "auto_commit", "prefetch", or other implementation-defined settings.
	params map[string]string
}

// ConsumeOption configures consumer behavior (concurrency, auto-ack, and broker-specific parameters).
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

// WithConcurrency sets how many handler goroutines process messages in parallel.
func WithConcurrency(n int) ConsumeOption {
	return func(o *consumeOptions) { o.concurrency = n }
}

// WithGroup sets the consumer group name (Kafka).
func WithGroup(group string) ConsumeOption {
	return func(o *consumeOptions) { o.group = group }
}

// WithChannel sets the channel name (NSQ).
func WithChannel(channel string) ConsumeOption {
	return func(o *consumeOptions) { o.channel = channel }
}

// WithQueueGroup sets the queue group name (NATS).
func WithQueueGroup(queueGroup string) ConsumeOption {
	return func(o *consumeOptions) { o.queueGroup = queueGroup }
}

// WithSubscription sets the subscription name (Google Pub/Sub).
func WithSubscription(subscription string) ConsumeOption {
	return func(o *consumeOptions) { o.subscription = subscription }
}

// WithAutoAck controls whether the wrapper should ack/nack automatically after the handler returns.
func WithAutoAck(autoAck bool) ConsumeOption {
	return func(o *consumeOptions) { o.autoAck = autoAck }
}

// WithAutoAct is an alias for WithAutoAck.
func WithAutoAct(autoAck bool) ConsumeOption {
	return WithAutoAck(autoAck)
}

// WithMaxInFlight limits the maximum number of unacknowledged messages in flight.
func WithMaxInFlight(maxInFlight int) ConsumeOption {
	return func(o *consumeOptions) { o.maxInFlight = maxInFlight }
}

// WithParams sets broker-specific parameters in bulk.
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

// WithParam sets a single broker-specific parameter.
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
