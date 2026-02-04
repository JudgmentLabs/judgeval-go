package judgeval

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type JudgmentSpanProcessor struct {
	delegate  sdktrace.SpanProcessor
	lifecycle []sdktrace.SpanProcessor
}

func NewJudgmentSpanProcessor(
	delegate sdktrace.SpanProcessor,
	lifecycle []sdktrace.SpanProcessor,
) sdktrace.SpanProcessor {
	return &JudgmentSpanProcessor{
		delegate:  delegate,
		lifecycle: lifecycle,
	}
}

func (p *JudgmentSpanProcessor) ForceFlush(ctx context.Context) error {
	return p.delegate.ForceFlush(ctx)
}

func (p *JudgmentSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	p.delegate.OnEnd(s)
}

func (p *JudgmentSpanProcessor) Shutdown(ctx context.Context) error {
	return p.delegate.Shutdown(ctx)
}

func (p *JudgmentSpanProcessor) OnStart(parentContext context.Context, span sdktrace.ReadWriteSpan) {
	for _, processor := range p.lifecycle {
		processor.OnStart(parentContext, span)
	}
	p.delegate.OnStart(parentContext, span)
}

type NoOpSpanProcessor struct{}

func NewNoOpSpanProcessor() sdktrace.SpanProcessor {
	return &NoOpSpanProcessor{}
}

func (p *NoOpSpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

func (p *NoOpSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
}

func (p *NoOpSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

func (p *NoOpSpanProcessor) OnStart(parentContext context.Context, span sdktrace.ReadWriteSpan) {
}
