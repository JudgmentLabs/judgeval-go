package judgeval

import (
	"context"

	"github.com/JudgmentLabs/judgeval-go/logger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	"go.opentelemetry.io/otel/trace/noop"
)

type FilterTracerFunc func(name string, opts ...trace.TracerOption) bool

type JudgmentTracerProvider struct {
	embedded.TracerProvider
	delegate     *sdktrace.TracerProvider
	filterTracer FilterTracerFunc
}

var _ trace.TracerProvider = (*JudgmentTracerProvider)(nil)

type JudgmentTracerProviderConfig struct {
	FilterTracer FilterTracerFunc
}

func NewJudgmentTracerProvider(config *JudgmentTracerProviderConfig, opts ...sdktrace.TracerProviderOption) *JudgmentTracerProvider {
	filterTracer := func(name string, opts ...trace.TracerOption) bool { return true }
	if config != nil && config.FilterTracer != nil {
		filterTracer = config.FilterTracer
	}

	provider := sdktrace.NewTracerProvider(opts...)

	return &JudgmentTracerProvider{
		delegate:     provider,
		filterTracer: filterTracer,
	}
}

func (j *JudgmentTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if name == TracerName {
		return j.delegate.Tracer(name, opts...)
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("[JudgmentTracerProvider] Failed to filter tracer %s: %v.", name, r)
		}
	}()

	if j.filterTracer(name, opts...) {
		return j.delegate.Tracer(name, opts...)
	}

	logger.Debug("[JudgmentTracerProvider] Returning NoOpTracer for tracer %s as it is disallowed by the filterTracer callback.", name)
	return noop.NewTracerProvider().Tracer(name, opts...)
}

func (j *JudgmentTracerProvider) Shutdown(ctx context.Context) error {
	return j.delegate.Shutdown(ctx)
}

func (j *JudgmentTracerProvider) ForceFlush(ctx context.Context) error {
	return j.delegate.ForceFlush(ctx)
}
