package judgeval

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JudgmentLabs/judgeval-go/internal/api"
	"github.com/JudgmentLabs/judgeval-go/internal/api/models"
	"github.com/JudgmentLabs/judgeval-go/logger"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type BaseTracer struct {
	projectName      string
	projectID        string
	enableEvaluation bool
	apiClient        *api.Client
	serializer       SerializerFunc
	tracer           trace.Tracer
}

func (b *BaseTracer) GetTracer() trace.Tracer {
	return b.tracer
}

func (b *BaseTracer) Span(ctx context.Context, spanName string) (context.Context, trace.Span) {
	ctx, span := b.tracer.Start(ctx, spanName)
	return ctx, span
}

func (b *BaseTracer) SetSpanKind(span trace.Span, kind string) {
	if kind != "" {
		span.SetAttributes(attribute.String(AttributeKeysJudgmentSpanKind, kind))
	}
}

func (b *BaseTracer) SetLLMSpan(span trace.Span) {
	b.SetSpanKind(span, "llm")
}

func (b *BaseTracer) SetToolSpan(span trace.Span) {
	b.SetSpanKind(span, "tool")
}

func (b *BaseTracer) SetGeneralSpan(span trace.Span) {
	b.SetSpanKind(span, "span")
}

func (b *BaseTracer) SetAttribute(span trace.Span, key string, value interface{}) {
	if key == "" {
		return
	}

	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int8:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case int16:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case int32:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case uint:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case uint8:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case uint16:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case uint32:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case uint64:
		span.SetAttributes(attribute.Int64(key, int64(v)))
	case float32:
		span.SetAttributes(attribute.Float64(key, float64(v)))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	case []string:
		span.SetAttributes(attribute.StringSlice(key, v))
	default:
		serialized, err := b.serializer(v)
		if err == nil {
			span.SetAttributes(attribute.String(key, serialized))
		}
	}
}

func (b *BaseTracer) SetAttributes(span trace.Span, attrs map[string]interface{}) {
	for k, v := range attrs {
		b.SetAttribute(span, k, v)
	}
}

func (b *BaseTracer) SetInput(span trace.Span, input interface{}) {
	b.SetAttribute(span, AttributeKeysJudgmentInput, input)
}

func (b *BaseTracer) SetOutput(span trace.Span, output interface{}) {
	b.SetAttribute(span, AttributeKeysJudgmentOutput, output)
}

func (b *BaseTracer) AsyncEvaluate(ctx context.Context, scorer BaseScorer, example *Example) {
	if !b.enableEvaluation {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsSampled() {
		return
	}

	spanContext := span.SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	logger.Info("asyncEvaluate: project=%s, traceId=%s, spanId=%s, scorer=%s",
		b.projectName, traceID, spanID, scorer.GetName())

	evaluationRun := b.createEvaluationRun(scorer, example, traceID, spanID)

	go func() {
		if _, err := b.apiClient.AddToRunEvalQueue(evaluationRun); err != nil {
			logger.Error("Failed to enqueue evaluation run: %v", err)
		}
	}()
}

func (b *BaseTracer) AsyncTraceEvaluate(ctx context.Context, scorer BaseScorer) {
	if !b.enableEvaluation {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsSampled() {
		return
	}

	spanContext := span.SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	logger.Info("asyncTraceEvaluate: project=%s, traceId=%s, spanId=%s, scorer=%s",
		b.projectName, traceID, spanID, scorer.GetName())

	evaluationRun := b.createTraceEvaluationRun(scorer, traceID, spanID)

	traceEvalJSON, err := json.Marshal(evaluationRun)
	if err != nil {
		logger.Error("Failed to serialize trace evaluation: %v", err)
		return
	}

	span.SetAttributes(attribute.String(AttributeKeysPendingTraceEval, string(traceEvalJSON)))
}

func (b *BaseTracer) getSpanExporter(ctx context.Context) sdktrace.SpanExporter {
	if b.projectID != "" {
		return NewJudgmentSpanExporter(ctx, b.buildEndpoint(), b.apiClient, b.projectID)
	}
	logger.Error("Project not resolved; cannot create exporter, returning NoOpSpanExporter")
	return NewNoOpSpanExporter()
}

func (b *BaseTracer) getSpanProcessor(ctx context.Context) sdktrace.SpanProcessor {
	if b.projectID != "" {
		exporter := b.getSpanExporter(ctx)
		batchProcessor := sdktrace.NewBatchSpanProcessor(exporter)
		return NewJudgmentSpanProcessor(batchProcessor)
	}
	logger.Error("Project not resolved; cannot create processor, returning NoOpSpanProcessor")
	return NewNoOpSpanProcessor()
}

func (b *BaseTracer) buildEndpoint() string {
	baseURL := b.apiClient.GetBaseURL()
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		return baseURL + "otel/v1/traces"
	}
	return baseURL + "/otel/v1/traces"
}

func (b *BaseTracer) createEvaluationRun(scorer BaseScorer, example *Example, traceID, spanID string) *models.ExampleEvaluationRun {
	runID := "async_evaluate_" + spanID

	return &models.ExampleEvaluationRun{
		Id:              uuid.New().String(),
		ProjectName:     b.projectName,
		EvalName:        runID,
		TraceId:         traceID,
		TraceSpanId:     spanID,
		Examples:        []models.Example{example.toModel()},
		JudgmentScorers: []models.ScorerConfig{*scorer.GetScorerConfig()},
		CustomScorers:   []models.BaseScorer{},
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
}

func (b *BaseTracer) createTraceEvaluationRun(scorer BaseScorer, traceID, spanID string) *models.TraceEvaluationRun {
	evalName := "async_trace_evaluate_" + spanID

	return &models.TraceEvaluationRun{
		Id:              uuid.New().String(),
		ProjectName:     b.projectName,
		EvalName:        evalName,
		TraceAndSpanIds: [][]any{{traceID, spanID}},
		JudgmentScorers: []models.ScorerConfig{*scorer.GetScorerConfig()},
		CustomScorers:   []models.BaseScorer{},
		IsOffline:       false,
		IsBucketRun:     false,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
}

func (b *BaseTracer) StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	ctx, span := b.tracer.Start(ctx, spanName)
	return ctx, span
}

func (b *BaseTracer) EndSpan(span trace.Span) {
	span.End()
}
