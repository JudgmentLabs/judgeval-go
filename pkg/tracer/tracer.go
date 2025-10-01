package tracer

import (
	"context"

	"github.com/JudgmentLabs/judgeval-go/pkg/logger"
	"github.com/JudgmentLabs/judgeval-go/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	TRACER_NAME = "judgeval"
)

type Tracer struct {
	*BaseTracer
	tracerProvider *sdktrace.TracerProvider
}

type TracerOptions func(*TracerConfig)

type TracerConfig struct {
	Configuration TracerConfiguration
	Serializer    ISerializer
	Initialize    bool
}

func WithConfiguration(config TracerConfiguration) TracerOptions {
	return func(tc *TracerConfig) {
		tc.Configuration = config
	}
}

func WithSerializer(serializer ISerializer) TracerOptions {
	return func(tc *TracerConfig) {
		tc.Serializer = serializer
	}
}

func WithInitialize(initialize bool) TracerOptions {
	return func(tc *TracerConfig) {
		tc.Initialize = initialize
	}
}

func NewTracer(options ...TracerOptions) *Tracer {
	config := &TracerConfig{
		Serializer: NewJSONSerializer(),
		Initialize: true,
	}

	for _, option := range options {
		option(config)
	}

	if config.Configuration.APIURL == "" {
		panic("Configuration is required")
	}

	baseTracer := NewBaseTracer(config.Configuration, config.Serializer, false)

	tracer := &Tracer{
		BaseTracer: baseTracer,
	}

	if config.Initialize {
		tracer.Initialize()
	}

	return tracer
}

func (t *Tracer) Initialize() {
	spanExporter := t.GetSpanExporter()

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(t.configuration.ProjectName),
			semconv.TelemetrySDKName(TRACER_NAME),
			semconv.TelemetrySDKVersion(version.Version),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		logger.Error("Failed to create resource: %v", err)
		return
	}

	t.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(spanExporter)),
	)

	otel.SetTracerProvider(t.tracerProvider)
}

func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.tracerProvider != nil {
		return t.tracerProvider.Shutdown(ctx)
	}
	return nil
}

func (t *Tracer) Flush(ctx context.Context) error {
	if t.tracerProvider != nil {
		return t.tracerProvider.ForceFlush(ctx)
	}
	return nil
}

func (t *Tracer) GetTracerProvider() *sdktrace.TracerProvider {
	return t.tracerProvider
}
