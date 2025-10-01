package exporters

import (
	"context"
	"net/http"
	"time"

	"github.com/JudgmentLabs/judgeval-go/pkg/logger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

type JudgmentSpanExporter struct {
	delegate  trace.SpanExporter
	endpoint  string
	apiKey    string
	orgID     string
	projectID string
}

type JudgmentSpanExporterBuilder struct {
	endpoint  string
	apiKey    string
	orgID     string
	projectID string
}

func NewJudgmentSpanExporterBuilder() *JudgmentSpanExporterBuilder {
	return &JudgmentSpanExporterBuilder{}
}

func (b *JudgmentSpanExporterBuilder) Endpoint(endpoint string) *JudgmentSpanExporterBuilder {
	b.endpoint = endpoint
	return b
}

func (b *JudgmentSpanExporterBuilder) APIKey(apiKey string) *JudgmentSpanExporterBuilder {
	b.apiKey = apiKey
	return b
}

func (b *JudgmentSpanExporterBuilder) OrganizationID(orgID string) *JudgmentSpanExporterBuilder {
	b.orgID = orgID
	return b
}

func (b *JudgmentSpanExporterBuilder) ProjectID(projectID string) *JudgmentSpanExporterBuilder {
	b.projectID = projectID
	return b
}

func (b *JudgmentSpanExporterBuilder) Build() *JudgmentSpanExporter {

	if b.endpoint == "" {
		panic("endpoint is required")
	}
	if b.apiKey == "" {
		panic("apiKey is required")
	}
	if b.orgID == "" {
		panic("organizationId is required")
	}
	if b.projectID == "" {
		panic("projectId is required")
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	delegate, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpointURL(b.endpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization":     "Bearer " + b.apiKey,
			"X-Organization-Id": b.orgID,
			"X-Project-Id":      b.projectID,
		}),
		otlptracehttp.WithHTTPClient(client),
	)
	if err != nil {
		panic("Failed to create OTLP HTTP exporter: " + err.Error())
	}

	return &JudgmentSpanExporter{
		delegate:  delegate,
		endpoint:  b.endpoint,
		apiKey:    b.apiKey,
		orgID:     b.orgID,
		projectID: b.projectID,
	}
}

func (j *JudgmentSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	logger.Info("JudgmentSpanExporter: Exported %d spans", len(spans))
	return j.delegate.ExportSpans(ctx, spans)
}

func (j *JudgmentSpanExporter) Shutdown(ctx context.Context) error {
	logger.Info("JudgmentSpanExporter: Shutting down exporter for project %s", j.projectID)
	return j.delegate.Shutdown(ctx)
}

func (j *JudgmentSpanExporter) ForceFlush(ctx context.Context) error {
	logger.Info("JudgmentSpanExporter: Force flushing spans for project %s", j.projectID)
	return nil
}
