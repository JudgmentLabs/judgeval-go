package judgeval

import (
	"context"

	"go.opentelemetry.io/otel/trace/noop"
)

// NoOpTracer is a no-op implementation of JudgevalTracer.
type NoOpTracer struct {
	*BaseTracer
}

var _ JudgevalTracer = (*NoOpTracer)(nil)

func NewNoOpTracer() *NoOpTracer {
	noopTracer := noop.NewTracerProvider().Tracer("noop")
	return &NoOpTracer{
		BaseTracer: &BaseTracer{
			tracer: noopTracer,
		},
	}
}

func (n *NoOpTracer) Initialize(ctx context.Context) error {
	return nil
}

func (n *NoOpTracer) ForceFlush(ctx context.Context) error {
	return nil
}

func (n *NoOpTracer) Shutdown(ctx context.Context) error {
	return nil
}
