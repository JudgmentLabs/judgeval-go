package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JudgmentLabs/judgeval-go/pkg/env"
	"github.com/JudgmentLabs/judgeval-go/pkg/logger"
	"github.com/JudgmentLabs/judgeval-go/pkg/version"
	"github.com/JudgmentLabs/judgeval-go/v1/internal/api"
	"github.com/JudgmentLabs/judgeval-go/v1/internal/api/models"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const TracerName = "judgeval"

type TracerFactory struct {
	client *api.Client
}

type TracerCreateParams struct {
	ProjectName        string
	EnableEvaluation   *bool
	Serializer         SerializerFunc
	ResourceAttributes map[string]interface{}
	Initialize         *bool
}

func (f *TracerFactory) Create(ctx context.Context, params TracerCreateParams) (*Tracer, error) {
	if params.ProjectName == "" {
		return nil, fmt.Errorf("project name is required")
	}

	serializer := params.Serializer
	if serializer == nil {
		serializer = defaultJSONSerializer
	}

	projectID, err := resolveProjectID(f.client, params.ProjectName)
	if err != nil {
		logger.Error("Failed to resolve project %s: %v. Skipping Judgment export.", params.ProjectName, err)
		projectID = ""
	}

	tracer := &Tracer{
		BaseTracer: &BaseTracer{
			projectName:      params.ProjectName,
			projectID:        projectID,
			enableEvaluation: getBool(params.EnableEvaluation, true),
			apiClient:        f.client,
			serializer:       serializer,
			tracer:           otel.Tracer(TracerName),
		},
		resourceAttributes: params.ResourceAttributes,
	}

	if getBool(params.Initialize, false) {
		if err := tracer.Initialize(ctx); err != nil {
			return nil, err
		}
	}

	return tracer, nil
}

type Tracer struct {
	*BaseTracer
	tracerProvider     *sdktrace.TracerProvider
	resourceAttributes map[string]interface{}
}

func (t *Tracer) Initialize(ctx context.Context) error {
	if t.tracerProvider != nil {
		logger.Warning("Tracer already initialized")
		return nil
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceName(t.projectName),
		attribute.String("telemetry.sdk.name", TracerName),
		attribute.String("telemetry.sdk.version", version.Version),
	}

	for k, v := range t.resourceAttributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)

	spanExporter := t.getSpanExporter(ctx)

	t.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(spanExporter),
	)

	otel.SetTracerProvider(t.tracerProvider)

	logger.Info("Tracer initialized successfully")
	return nil
}

func (t *Tracer) ForceFlush(ctx context.Context) error {
	if t.tracerProvider == nil {
		logger.Warning("Tracer not initialized, skipping force flush")
		return nil
	}
	return t.tracerProvider.ForceFlush(ctx)
}

func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.tracerProvider == nil {
		logger.Warning("Tracer not initialized, skipping shutdown")
		return nil
	}

	err := t.tracerProvider.Shutdown(ctx)
	if err != nil {
		logger.Error("Failed to shutdown Tracer: %v", err)
		return err
	}

	t.tracerProvider = nil
	logger.Info("Tracer shut down successfully")
	return nil
}

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
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
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

func (b *BaseTracer) AsyncEvaluate(ctx context.Context, scorer BaseScorer, example *Example, model *string) {
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

	evaluationRun := b.createEvaluationRun(scorer, example, model, traceID, spanID)

	go func() {
		if _, err := b.apiClient.AddToRunEvalQueue(evaluationRun); err != nil {
			logger.Error("Failed to enqueue evaluation run: %v", err)
		}
	}()
}

func (b *BaseTracer) AsyncTraceEvaluate(ctx context.Context, scorer BaseScorer, model *string) {
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

	evaluationRun := b.createTraceEvaluationRun(scorer, model, traceID, spanID)

	traceEvalJSON, err := json.Marshal(evaluationRun)
	if err != nil {
		logger.Error("Failed to serialize trace evaluation: %v", err)
		return
	}

	span.SetAttributes(attribute.String(AttributeKeysPendingTraceEval, string(traceEvalJSON)))
}

func (b *BaseTracer) getSpanExporter(ctx context.Context) sdktrace.SpanExporter {
	if b.projectID != "" {
		return newJudgmentSpanExporter(ctx, b.buildEndpoint(), b.apiClient, b.projectID)
	}
	logger.Error("Project not resolved; cannot create exporter, returning NoOpSpanExporter")
	return newNoOpSpanExporter()
}

func (b *BaseTracer) buildEndpoint() string {
	baseURL := b.apiClient.GetBaseURL()
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		return baseURL + "otel/v1/traces"
	}
	return baseURL + "/otel/v1/traces"
}

func (b *BaseTracer) createEvaluationRun(scorer BaseScorer, example *Example, model *string, traceID, spanID string) *models.ExampleEvaluationRun {
	runID := "async_evaluate_" + spanID
	modelName := getString(model, env.JudgmentDefaultGPTModel)

	return &models.ExampleEvaluationRun{
		Id:              uuid.New().String(),
		ProjectName:     b.projectName,
		EvalName:        runID,
		Model:           modelName,
		TraceId:         traceID,
		TraceSpanId:     spanID,
		Examples:        []models.Example{example.toModel()},
		JudgmentScorers: []models.ScorerConfig{*scorer.GetScorerConfig()},
		CustomScorers:   []models.BaseScorer{},
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
}

func (b *BaseTracer) createTraceEvaluationRun(scorer BaseScorer, model *string, traceID, spanID string) *models.TraceEvaluationRun {
	evalName := "async_trace_evaluate_" + spanID
	modelName := getString(model, env.JudgmentDefaultGPTModel)

	return &models.TraceEvaluationRun{
		Id:              uuid.New().String(),
		ProjectName:     b.projectName,
		EvalName:        evalName,
		Model:           modelName,
		TraceAndSpanIds: [][]interface{}{{traceID, spanID}},
		JudgmentScorers: []models.ScorerConfig{*scorer.GetScorerConfig()},
		CustomScorers:   []models.BaseScorer{},
		IsOffline:       false,
		IsBucketRun:     false,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
}

func resolveProjectID(client *api.Client, projectName string) (string, error) {
	logger.Info("Resolving project ID for project: %s", projectName)

	req := &models.ResolveProjectNameRequest{
		ProjectName: projectName,
	}

	resp, err := client.ProjectsResolve(req)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project ID: %w", err)
	}

	if resp.ProjectId == "" {
		return "", fmt.Errorf("project ID not found for project: %s", projectName)
	}

	logger.Info("Resolved project ID: %s", resp.ProjectId)
	return resp.ProjectId, nil
}

func defaultJSONSerializer(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type spanWrapper struct {
	span trace.Span
	ctx  context.Context
}

func (b *BaseTracer) StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	ctx, span := b.tracer.Start(ctx, spanName)
	return ctx, span
}

func (b *BaseTracer) EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
