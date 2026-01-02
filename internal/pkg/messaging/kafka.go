package messaging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	// ErrKafkaTopicRequired is returned when the topic is empty.
	ErrKafkaTopicRequired = errors.New("pkgmessage: kafka topic is required")
	// ErrKafkaHandlerRequired is returned when Consume is called with a nil handler.
	ErrKafkaHandlerRequired = errors.New("pkgmessage: kafka handler is required")
	// ErrKafkaBrokersRequired is returned when no Kafka brokers are configured.
	ErrKafkaBrokersRequired = errors.New("pkgmessage: kafka brokers are required")
	// ErrKafkaGroupRequired is returned when a consumer group is required but not provided.
	ErrKafkaGroupRequired = errors.New("pkgmessage: kafka consumer group is required")
)

// KafkaConfig configures the Kafka implementation.
type KafkaConfig struct {
	// Brokers lists Kafka broker addresses.
	Brokers []string

	// Dialer configures broker connections.
	Dialer *kafka.Dialer

	// WriterConfig overrides the default writer configuration.
	WriterConfig *kafka.WriterConfig
	// ReaderConfig overrides the default reader configuration.
	ReaderConfig *kafka.ReaderConfig
}

// Kafka is a messaging implementation backed by kafka-go.
type Kafka struct {
	brokers []string
	dialer  *kafka.Dialer

	writerConfig *kafka.WriterConfig
	readerConfig *kafka.ReaderConfig

	mu      sync.Mutex
	writers map[string]*kafka.Writer
	readers []*kafka.Reader
	closed  bool
}

// NewKafka constructs a Kafka messaging client.
func NewKafka(cfg KafkaConfig) (*Kafka, error) {
	if len(cfg.Brokers) == 0 {
		return nil, ErrKafkaBrokersRequired
	}

	return &Kafka{
		brokers: append([]string{}, cfg.Brokers...),
		dialer:  cfg.Dialer,

		writerConfig: cfg.WriterConfig,
		readerConfig: cfg.ReaderConfig,

		writers: map[string]*kafka.Writer{},
	}, nil
}

// Close shuts down all Kafka readers and writers.
func (k *Kafka) Close() error {
	k.mu.Lock()
	if k.closed {
		k.mu.Unlock()
		return nil
	}
	k.closed = true
	writers := make([]*kafka.Writer, 0, len(k.writers))
	for _, w := range k.writers {
		writers = append(writers, w)
	}
	k.writers = nil
	readers := append([]*kafka.Reader{}, k.readers...)
	k.readers = nil
	k.mu.Unlock()

	var closeErr error
	for _, r := range readers {
		closeErr = errors.Join(closeErr, r.Close())
	}
	for _, w := range writers {
		closeErr = errors.Join(closeErr, w.Close())
	}
	return closeErr
}

// Publish sends a message to a Kafka topic.
func (k *Kafka) Publish(ctx context.Context, destination string, msg OutgoingMessage) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if destination == "" {
		return PublishResult{}, ErrKafkaTopicRequired
	}
	if msg.Delay > 0 {
		return PublishResult{}, ErrUnsupported
	}
	if err := k.ensureOpen(); err != nil {
		return PublishResult{}, err
	}

	writer := k.getWriter(destination)
	kmsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Body,
		Time:  time.Now(),
	}

	for _, h := range msg.Headers {
		if h.Key == "" {
			continue
		}
		kmsg.Headers = append(kmsg.Headers, kafka.Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	if err := writer.WriteMessages(ctx, kmsg); err != nil {
		return PublishResult{}, fmt.Errorf("pkgmessage: kafka publish: %w", err)
	}

	return PublishResult{
		Topic:     destination,
		Timestamp: kmsg.Time,
	}, nil
}

// Consume starts consuming messages from a Kafka topic.
func (k *Kafka) Consume(ctx context.Context, source string, handler Handler, opts ...ConsumeOption) error {
	co := newConsumeOptions(opts...)
	if err := validateKafkaConsume(ctx, source, handler, co); err != nil {
		return err
	}
	if err := k.ensureOpen(); err != nil {
		return err
	}
	if k.isClosed() {
		return io.ErrClosedPipe
	}

	consumeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	reader := k.newReader(source, co)
	if err := k.addReader(reader); err != nil {
		return errors.Join(err, reader.Close())
	}

	autoAck := co.autoAck
	concurrency := concurrencyOrDefault(co.concurrency, 1)

	msgCh := make(chan kafka.Message)
	errCh := make(chan error, 1)

	go kafkaFetchLoop(consumeCtx, reader, msgCh, errCh)

	wg := startKafkaWorkers(consumeCtx, cancel, reader, handler, autoAck, concurrency, msgCh, errCh)
	waitErr := waitKafkaConsume(ctx, errCh, wg)
	k.removeReader(reader)
	closeErr := reader.Close()
	if closeErr != nil {
		return errors.Join(waitErr, closeErr)
	}
	return waitErr
}

func (k *Kafka) ensureOpen() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return io.ErrClosedPipe
	}
	return nil
}

func (k *Kafka) isClosed() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.closed
}

func (k *Kafka) getWriter(topic string) *kafka.Writer {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.writers == nil {
		k.writers = map[string]*kafka.Writer{}
	}
	if w, ok := k.writers[topic]; ok {
		return w
	}

	cfg := kafka.WriterConfig{
		Brokers:  k.brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Dialer:   k.dialer,
	}
	if k.writerConfig != nil {
		cfg = *k.writerConfig
		cfg.Topic = topic
		if len(cfg.Brokers) == 0 {
			cfg.Brokers = k.brokers
		}
		if cfg.Dialer == nil {
			cfg.Dialer = k.dialer
		}
		if cfg.Balancer == nil {
			cfg.Balancer = &kafka.LeastBytes{}
		}
	}

	w := kafka.NewWriter(cfg)
	k.writers[topic] = w
	return w
}

func (k *Kafka) newReader(topic string, opts consumeOptions) *kafka.Reader {
	cfg := kafka.ReaderConfig{
		Brokers:  k.brokers,
		GroupID:  opts.group,
		Topic:    topic,
		MaxBytes: 10e6,
		Dialer:   k.dialer,
	}
	if k.readerConfig != nil {
		cfg = *k.readerConfig
		cfg.Topic = topic
		cfg.GroupID = opts.group
		if len(cfg.Brokers) == 0 {
			cfg.Brokers = k.brokers
		}
		if cfg.Dialer == nil {
			cfg.Dialer = k.dialer
		}
		if cfg.MaxBytes == 0 {
			cfg.MaxBytes = 10e6
		}
	}

	return kafka.NewReader(cfg)
}

func (k *Kafka) addReader(reader *kafka.Reader) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return io.ErrClosedPipe
	}
	k.readers = append(k.readers, reader)
	return nil
}

func (k *Kafka) removeReader(reader *kafka.Reader) {
	if reader == nil {
		return
	}
	k.mu.Lock()
	defer k.mu.Unlock()

	for i := range k.readers {
		if k.readers[i] == reader {
			k.readers = append(k.readers[:i], k.readers[i+1:]...)
			return
		}
	}
}

func handleKafkaMessage(
	ctx context.Context,
	reader *kafka.Reader,
	m kafka.Message,
	handler Handler,
	autoAck bool,
) error {
	wrapped := newKafkaMessage(reader, m)
	herr := callHandlerWithRecover(ctx, "kafka", func() error {
		return handler(ctx, wrapped)
	})

	if wrapped.hasResponded() || !autoAck {
		return nil
	}

	if herr == nil {
		return wrapped.Ack(ctx)
	}

	if err := wrapped.Nack(ctx); err != nil {
		return err
	}
	return nil
}

func trySendErr(ch chan<- error, err error) {
	if err == nil {
		return
	}
	select {
	case ch <- err:
	default:
	}
}

func validateKafkaConsume(ctx context.Context, topic string, handler Handler, opts consumeOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if topic == "" {
		return ErrKafkaTopicRequired
	}
	if handler == nil {
		return ErrKafkaHandlerRequired
	}
	if opts.group == "" {
		return ErrKafkaGroupRequired
	}
	return nil
}

func kafkaFetchLoop(ctx context.Context, reader *kafka.Reader, msgCh chan<- kafka.Message, errCh chan<- error) {
	defer close(msgCh)

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			trySendErr(errCh, err)
			return
		}

		select {
		case msgCh <- m:
		case <-ctx.Done():
			trySendErr(errCh, ctx.Err())
			return
		}
	}
}

func startKafkaWorkers(
	ctx context.Context,
	cancel context.CancelFunc,
	reader *kafka.Reader,
	handler Handler,
	autoAck bool,
	concurrency int,
	msgCh <-chan kafka.Message,
	errCh chan<- error,
) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for range concurrency {
		go func() {
			defer wg.Done()
			for m := range msgCh {
				if err := handleKafkaMessage(ctx, reader, m, handler, autoAck); err != nil {
					trySendErr(errCh, err)
					cancel()
					return
				}
			}
		}()
	}
	return &wg
}

func waitKafkaConsume(ctx context.Context, errCh <-chan error, wg *sync.WaitGroup) error {
	select {
	case err := <-errCh:
		wg.Wait()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("pkgmessage: kafka consume: %w", err)
	case <-ctx.Done():
		wg.Wait()
		return ctx.Err()
	}
}
