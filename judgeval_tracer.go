package judgeval

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type JudgevalTracer interface {
	Initialize(ctx context.Context) error
	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error

	GetTracer() trace.Tracer
	Span(ctx context.Context, spanName string) (context.Context, trace.Span)
	SetSpanKind(span trace.Span, kind string)
	SetLLMSpan(span trace.Span)
	SetToolSpan(span trace.Span)
	SetGeneralSpan(span trace.Span)
	SetAttribute(span trace.Span, key string, value interface{})
	SetAttributes(span trace.Span, attrs map[string]interface{})
	SetInput(span trace.Span, input interface{})
	SetOutput(span trace.Span, output interface{})
	AsyncEvaluate(ctx context.Context, scorer BaseScorer, example *Example)
	AsyncTraceEvaluate(ctx context.Context, scorer BaseScorer)
	StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span)
	EndSpan(span trace.Span)
}
