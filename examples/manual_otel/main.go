package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JudgmentLabs/judgeval-go/pkg/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func processData(ctx context.Context, data string) string {
	_, span := otel.Tracer("data-processor").Start(ctx, "process_data")
	defer span.End()

	span.SetAttributes(
		attribute.String("data.input", data),
		attribute.Int("data.length", len(data)),
	)

	time.Sleep(100 * time.Millisecond)

	result := fmt.Sprintf("Processed: %s", data)
	span.SetAttributes(attribute.String("data.output", result))

	return result
}

func initOtel(judgmentTracer *tracer.Tracer) (func(), error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", "manual-otel-example"),
			attribute.String("service.version", "1.0.0"),
			attribute.String("telemetry.sdk.name", "opentelemetry"),
			attribute.String("telemetry.sdk.version", "1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var spanExporter sdktrace.SpanExporter
	if judgmentTracer != nil {
		spanExporter = judgmentTracer.GetSpanExporter()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(spanExporter)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		time.Sleep(10 * time.Second)
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}, nil
}

func main() {
	fmt.Println("Manual OpenTelemetry Instrumentation Example")
	fmt.Println("===========================================")

	judgmentTracer, _ := tracer.NewTracer(
		tracer.WithConfiguration(tracer.NewTracerConfiguration(
			tracer.WithProjectName("manual-otel-example"),
		)),
		tracer.WithInitialize(false),
	)

	cleanup, _ := initOtel(judgmentTracer)
	defer cleanup()
	defer judgmentTracer.Shutdown(context.Background())

	ctx := context.Background()

	ctx, rootSpan := otel.Tracer("main").Start(ctx, "manual_instrumentation_example")
	rootSpan.SetAttributes(attribute.String("example.type", "manual_otel"))

	result1 := processData(ctx, "Hello, World!")
	fmt.Printf("Result 1: %s\n", result1)

	result2 := processData(ctx, "Manual instrumentation")
	fmt.Printf("Result 2: %s\n", result2)

	judgmentSpan, _ := judgmentTracer.Span(ctx, "judgment_metrics")
	judgmentTracer.SetGeneralSpan(judgmentSpan)
	judgmentTracer.SetAttribute(judgmentSpan, "service.type", "manual_otel")

	requestData := map[string]interface{}{
		"endpoint":  "/api/process",
		"method":    "POST",
		"timestamp": time.Now().Unix(),
	}
	judgmentTracer.SetInput(judgmentSpan, requestData)

	time.Sleep(50 * time.Millisecond)

	responseData := map[string]interface{}{
		"status_code":      200,
		"response_time_ms": 50,
		"processed_items":  2,
	}
	judgmentTracer.SetOutput(judgmentSpan, responseData)
	judgmentSpan.End()

	rootSpan.End()

	fmt.Println("\nExample completed!")
}
