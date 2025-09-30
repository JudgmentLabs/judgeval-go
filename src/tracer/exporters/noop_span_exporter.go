package exporters

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
)

type NoOpSpanExporter struct{}

func NewNoOpSpanExporter() *NoOpSpanExporter {
	return &NoOpSpanExporter{}
}

func (n *NoOpSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	fmt.Printf("NoOpSpanExporter: Ignoring %d spans\n", len(spans))
	return nil
}

func (n *NoOpSpanExporter) Shutdown(ctx context.Context) error {
	fmt.Printf("NoOpSpanExporter: Shutdown called\n")
	return nil
}

func (n *NoOpSpanExporter) ForceFlush(ctx context.Context) error {
	fmt.Printf("NoOpSpanExporter: ForceFlush called\n")
	return nil
}
