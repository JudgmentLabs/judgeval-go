package tracer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JudgmentLabs/judgeval-go/src/data"
	"github.com/JudgmentLabs/judgeval-go/src/env"
	"github.com/JudgmentLabs/judgeval-go/src/internal/api"
	"github.com/JudgmentLabs/judgeval-go/src/internal/api/models"
	"github.com/JudgmentLabs/judgeval-go/src/scorers"
	"github.com/JudgmentLabs/judgeval-go/src/tracer/exporters"
	"github.com/JudgmentLabs/judgeval-go/src/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const TracerName = "judgeval"

type TracerConfiguration struct {
	APIURL           string
	APIKey           string
	OrganizationID   string
	ProjectName      string
	EnableEvaluation bool
}

type TracerConfigurationOptions func(*TracerConfiguration)

func WithAPIURL(apiURL string) TracerConfigurationOptions {
	return func(config *TracerConfiguration) {
		config.APIURL = apiURL
	}
}

func WithAPIKey(apiKey string) TracerConfigurationOptions {
	return func(config *TracerConfiguration) {
		config.APIKey = apiKey
	}
}

func WithOrganizationID(organizationID string) TracerConfigurationOptions {
	return func(config *TracerConfiguration) {
		config.OrganizationID = organizationID
	}
}

func WithProjectName(projectName string) TracerConfigurationOptions {
	return func(config *TracerConfiguration) {
		config.ProjectName = projectName
	}
}

func WithEnableEvaluation(enableEvaluation bool) TracerConfigurationOptions {
	return func(config *TracerConfiguration) {
		config.EnableEvaluation = enableEvaluation
	}
}

func NewTracerConfiguration(options ...TracerConfigurationOptions) TracerConfiguration {
	config := &TracerConfiguration{
		APIURL:           env.JudgmentAPIURL,
		APIKey:           env.JudgmentAPIKey,
		OrganizationID:   env.JudgmentOrgID,
		EnableEvaluation: true,
	}

	for _, option := range options {
		option(config)
	}

	return *config
}

type ISerializer interface {
	Serialize(obj interface{}) string
}

type BaseTracer struct {
	configuration TracerConfiguration
	apiClient     *api.Client
	serializer    ISerializer
	projectID     string
	tracer        trace.Tracer
}

func NewBaseTracer(config TracerConfiguration, serializer ISerializer, initialize bool) *BaseTracer {
	if config.APIURL == "" {
		panic("Configuration APIURL cannot be empty")
	}
	if serializer == nil {
		panic("Serializer cannot be nil")
	}

	apiClient := api.NewClient(config.APIURL, config.APIKey, config.OrganizationID)
	projectID := resolveProjectID(apiClient, config.ProjectName)

	if projectID == "" {
		utils.DefaultLogger.Error(fmt.Sprintf("Failed to resolve project %s, please create it first at https://app.judgmentlabs.ai/org/%s/projects. Skipping Judgment export.",
			config.ProjectName, config.OrganizationID))
	}

	tracer := otel.Tracer(TracerName)

	bt := &BaseTracer{
		configuration: config,
		apiClient:     apiClient,
		serializer:    serializer,
		projectID:     projectID,
		tracer:        tracer,
	}

	if initialize {
		bt.Initialize()
	}

	return bt
}

func (bt *BaseTracer) Initialize() {

}

func (bt *BaseTracer) GetSpanExporter() sdktrace.SpanExporter {
	if bt.projectID == "" {
		utils.DefaultLogger.Error("Project not resolved; cannot create exporter, returning NoOpSpanExporter")
		return exporters.NewNoOpSpanExporter()
	}
	return bt.createJudgmentSpanExporter(bt.projectID)
}

func (bt *BaseTracer) SetSpanKind(span trace.Span, kind string) {
	if span.IsRecording() && kind != "" {
		span.SetAttributes(attribute.String(AttributeKeys.JudgmentSpanKind, kind))
	}
}

func (bt *BaseTracer) SetAttribute(span trace.Span, key string, value interface{}) {
	if span.IsRecording() {
		span.SetAttributes(attribute.String(key, bt.serializer.Serialize(value)))
	}
}

func (bt *BaseTracer) AsyncEvaluate(scorer scorers.BaseScorer, example *data.Example, model string) {
	if !bt.configuration.EnableEvaluation {
		return
	}

	span := trace.SpanFromContext(context.Background())
	utils.DefaultLogger.Info(fmt.Sprintf("DEBUG: AsyncEvaluate - Span found: %v, IsRecording: %v", span != nil, span.IsRecording()))

	if !span.IsRecording() {
		utils.DefaultLogger.Info("DEBUG: Span not recording, returning")
		return
	}

	spanContext := span.SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	utils.DefaultLogger.Info(fmt.Sprintf("asyncEvaluate: project=%s, traceId=%s, spanId=%s, scorer=%s",
		bt.configuration.ProjectName, traceID, spanID, scorer.GetName()))

	evaluationRun := bt.createEvaluationRun(scorer, example, model, traceID, spanID)
	bt.enqueueEvaluation(evaluationRun)
}

func (bt *BaseTracer) AsyncEvaluateWithContext(ctx context.Context, scorer scorers.BaseScorer, example *data.Example, model string) {
	if !bt.configuration.EnableEvaluation {
		return
	}

	span := trace.SpanFromContext(ctx)
	utils.DefaultLogger.Info(fmt.Sprintf("DEBUG: AsyncEvaluateWithContext - Span found: %v, IsRecording: %v", span != nil, span.IsRecording()))

	if !span.IsRecording() {
		utils.DefaultLogger.Info("DEBUG: Span not recording, returning")
		return
	}

	spanContext := span.SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	utils.DefaultLogger.Info(fmt.Sprintf("asyncEvaluate: project=%s, traceId=%s, spanId=%s, scorer=%s",
		bt.configuration.ProjectName, traceID, spanID, scorer.GetName()))

	evaluationRun := bt.createEvaluationRun(scorer, example, model, traceID, spanID)
	bt.enqueueEvaluation(evaluationRun)
}

func (bt *BaseTracer) AsyncEvaluateWithDefaultModel(scorer scorers.BaseScorer, example *data.Example) {
	bt.AsyncEvaluate(scorer, example, "")
}

func (bt *BaseTracer) AsyncTraceEvaluate(scorer scorers.BaseScorer, model string) {
	if !bt.configuration.EnableEvaluation {
		return
	}

	span := trace.SpanFromContext(context.Background())
	if !span.IsRecording() {
		return
	}

	spanContext := span.SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	utils.DefaultLogger.Info(fmt.Sprintf("asyncTraceEvaluate: project=%s, traceId=%s, spanId=%s, scorer=%s",
		bt.configuration.ProjectName, traceID, spanID, scorer.GetName()))

	evaluationRun := bt.createTraceEvaluationRun(scorer, model, traceID, spanID)

	traceEvalJSON := bt.serializer.Serialize(evaluationRun)
	span.SetAttributes(attribute.String(AttributeKeys.PendingTraceEval, traceEvalJSON))
}

func (bt *BaseTracer) AsyncTraceEvaluateWithDefaultModel(scorer scorers.BaseScorer) {
	bt.AsyncTraceEvaluate(scorer, "")
}

func (bt *BaseTracer) SetAttributes(span trace.Span, attributes map[string]interface{}) {
	if attributes == nil {
		return
	}

	if span.IsRecording() {
		attrs := make([]attribute.KeyValue, 0, len(attributes))
		for key, value := range attributes {
			attrs = append(attrs, attribute.String(key, bt.serializer.Serialize(value)))
		}
		span.SetAttributes(attrs...)
	}
}

func (bt *BaseTracer) SetLLMSpan(span trace.Span) {
	bt.SetSpanKind(span, "llm")
}

func (bt *BaseTracer) SetToolSpan(span trace.Span) {
	bt.SetSpanKind(span, "tool")
}

func (bt *BaseTracer) SetGeneralSpan(span trace.Span) {
	bt.SetSpanKind(span, "span")
}

func (bt *BaseTracer) SetInput(span trace.Span, input interface{}) {
	bt.SetAttribute(span, AttributeKeys.JudgmentInput, input)
}

func (bt *BaseTracer) SetOutput(span trace.Span, output interface{}) {
	bt.SetAttribute(span, AttributeKeys.JudgmentOutput, output)
}

func (bt *BaseTracer) GetTracer() trace.Tracer {
	return bt.tracer
}

func (bt *BaseTracer) Span(ctx context.Context, spanName string) (trace.Span, context.Context) {
	spanCtx, span := bt.tracer.Start(ctx, spanName)

	spanCtx = trace.ContextWithSpan(spanCtx, span)
	return span, spanCtx
}

func resolveProjectID(apiClient *api.Client, projectName string) string {
	request := &models.ResolveProjectNameRequest{
		ProjectName: projectName,
	}

	response, err := apiClient.ProjectsResolve(request)
	if err != nil {
		return ""
	}

	if response.ProjectId != "" {
		return response.ProjectId
	}
	return ""
}

func (bt *BaseTracer) createJudgmentSpanExporter(projectID string) sdktrace.SpanExporter {

	endpoint := bt.configuration.APIURL
	if endpoint[len(endpoint)-1] != '/' {
		endpoint += "/"
	}
	endpoint += "otel/v1/traces"

	return exporters.NewJudgmentSpanExporterBuilder().
		Endpoint(endpoint).
		APIKey(bt.configuration.APIKey).
		OrganizationID(bt.configuration.OrganizationID).
		ProjectID(projectID).
		Build()
}

func (bt *BaseTracer) createEvaluationRun(scorer scorers.BaseScorer, example *data.Example, model, traceID, spanID string) *data.ExampleEvaluationRun {
	runID := fmt.Sprintf("async_evaluate_%s", spanID)
	if spanID == "" {
		runID = fmt.Sprintf("async_evaluate_%d", time.Now().UnixMilli())
	}

	modelName := model
	if modelName == "" {
		modelName = env.JudgmentDefaultGPTModel
	}

	return data.NewExampleEvaluationRun(
		data.WithProjectName(bt.configuration.ProjectName),
		data.WithEvalName(runID),
		data.WithExamples([]*data.Example{example}),
		data.WithScorers([]models.ScorerConfig{scorer.GetScorerConfig()}),
		data.WithModel(modelName),
		data.WithOrganizationId(bt.configuration.OrganizationID),
		data.WithTraceId(traceID),
		data.WithTraceSpanId(spanID),
	)
}

func (bt *BaseTracer) createTraceEvaluationRun(scorer scorers.BaseScorer, model, traceID, spanID string) *data.TraceEvaluationRun {
	evalName := fmt.Sprintf("async_trace_evaluate_%s", spanID)
	if spanID == "" {
		evalName = fmt.Sprintf("async_trace_evaluate_%d", time.Now().UnixMilli())
	}

	modelName := model
	if modelName == "" {
		modelName = env.JudgmentDefaultGPTModel
	}

	return data.NewTraceEvaluationRunWithOptions(data.TraceEvaluationRunOptions{
		ProjectName: bt.configuration.ProjectName,
		EvalName:    evalName,
		Scorer:      scorer.GetScorerConfig(),
		Model:       modelName,
		TraceId:     traceID,
		SpanId:      spanID,
	})
}

func (bt *BaseTracer) enqueueEvaluation(evaluationRun *data.ExampleEvaluationRun) {
	_, err := bt.apiClient.AddToRunEvalQueue(evaluationRun.ExampleEvaluationRun)
	if err != nil {
		utils.DefaultLogger.Error(fmt.Sprintf("Failed to enqueue evaluation run: %v", err))
	}
}

type JSONSerializer struct{}

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

func (js *JSONSerializer) Serialize(obj interface{}) string {
	if obj == nil {
		return "null"
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("%+v", obj)
	}

	return string(jsonBytes)
}
