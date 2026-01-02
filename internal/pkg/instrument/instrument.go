package instrument

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Instrumentation exposes tracing and metrics providers for dependency injection.
type Instrumentation interface {
	Tracer(name string) trace.Tracer
	Meter(name string) metric.Meter
	Shutdown(ctx context.Context) error
}

// Config drives OpenTelemetry initialization.
type Config struct {
	// Enabled toggles OpenTelemetry initialization.
	Enabled bool
	// ServiceName is the service.name resource attribute.
	ServiceName string
	// ServiceVersion is the service.version resource attribute.
	ServiceVersion string
	// Environment is the deployment environment name.
	Environment string
	// OTLPEndpoint is the OTLP collector endpoint.
	OTLPEndpoint string
	// OTLPSecure controls TLS usage for OTLP exporters.
	OTLPSecure bool
	// TraceSampleRatio controls trace sampling probability.
	TraceSampleRatio float64
	// MetricsInterval configures the metrics export interval.
	MetricsInterval time.Duration
	// MaskFields lists log field names to mask in output.
	MaskFields []string
}

type otelInstrumentation struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
}

// New builds an OpenTelemetry-backed implementation or returns a noop instance when disabled.
func New(ctx context.Context, cfg *Config) (Instrumentation, error) {
	if cfg == nil || !cfg.Enabled {
		return NewNoop(), nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("env", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	traceExporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
	}
	metricExporterOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
	}
	logExporterOpts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(cfg.OTLPEndpoint),
	}
	if !cfg.OTLPSecure {
		traceExporterOpts = append(traceExporterOpts, otlptracegrpc.WithInsecure())
		metricExporterOpts = append(metricExporterOpts, otlpmetricgrpc.WithInsecure())
		logExporterOpts = append(logExporterOpts, otlploggrpc.WithInsecure())
	}

	traceExporter, err := otlptracegrpc.New(ctx, traceExporterOpts...)
	if err != nil {
		return nil, err
	}

	logExporter, err := otlploggrpc.New(ctx, logExporterOpts...)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetricgrpc.New(ctx, metricExporterOpts...)
	if err != nil {
		return nil, err
	}

	ratio := cfg.TraceSampleRatio
	if ratio <= 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(traceExporter),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(cfg.MetricsInterval))),
	)

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)

	initLogging(cfg.ServiceName, lp, cfg.MaskFields)

	return &otelInstrumentation{
		tracerProvider: tp,
		meterProvider:  mp,
		loggerProvider: lp,
	}, nil
}

// Tracer returns a tracer for the given name.
func (o *otelInstrumentation) Tracer(name string) trace.Tracer {
	return o.tracerProvider.Tracer(name)
}

// Meter returns a meter for the given name.
func (o *otelInstrumentation) Meter(name string) metric.Meter {
	return o.meterProvider.Meter(name)
}

// Shutdown flushes and stops tracing, metrics, and logs.
func (o *otelInstrumentation) Shutdown(ctx context.Context) error {
	return errors.Join([]error{
		o.tracerProvider.Shutdown(ctx),
		o.meterProvider.Shutdown(ctx),
		o.loggerProvider.Shutdown(ctx),
	}...)
}

// NewNoop returns a no-op implementation suitable for unit tests.
func NewNoop() Instrumentation {
	return &noopInstrumentation{
		tracerProvider: tracenoop.NewTracerProvider(),
		meterProvider:  metricnoop.NewMeterProvider(),
	}
}

type noopInstrumentation struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
}

// Tracer returns a no-op tracer.
func (n *noopInstrumentation) Tracer(name string) trace.Tracer {
	return n.tracerProvider.Tracer(name)
}

// Meter returns a no-op meter.
func (n *noopInstrumentation) Meter(name string) metric.Meter {
	return n.meterProvider.Meter(name)
}

// Shutdown is a no-op for the noop instrumentation.
func (n *noopInstrumentation) Shutdown(ctx context.Context) error {
	return nil
}
