package judgeval

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JudgmentLabs/judgeval-go/internal/api"
	"github.com/JudgmentLabs/judgeval-go/internal/api/models"
	"github.com/JudgmentLabs/judgeval-go/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const TracerName = "judgeval"

type TracerFactory struct {
	client      *api.Client
	projectName string
	projectID   string
}

type TracerCreateParams struct {
	EnableEvaluation   *bool
	Serializer         SerializerFunc
	ResourceAttributes map[string]any
	FilterTracer       FilterTracerFunc
	Initialize         *bool
}

func (f *TracerFactory) Create(ctx context.Context, params TracerCreateParams) (*Tracer, error) {
	serializer := params.Serializer
	if serializer == nil {
		serializer = defaultJSONSerializer
	}

	tracer := &Tracer{
		BaseTracer: &BaseTracer{
			projectName:      f.projectName,
			projectID:        f.projectID,
			enableEvaluation: getBool(params.EnableEvaluation, true),
			apiClient:        f.client,
			serializer:       serializer,
		},
		resourceAttributes: params.ResourceAttributes,
		filterTracer:       params.FilterTracer,
	}

	if getBool(params.Initialize, true) {
		if err := tracer.Initialize(ctx); err != nil {
			return nil, err
		}
	}

	return tracer, nil
}

var _ JudgevalTracerLike = (*Tracer)(nil)

type Tracer struct {
	*BaseTracer
	tracerProvider     *JudgmentTracerProvider
	resourceAttributes map[string]any
	filterTracer       FilterTracerFunc
}

func (t *Tracer) Initialize(ctx context.Context) error {
	if t.tracerProvider != nil {
		logger.Warning("Tracer already initialized")
		return nil
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceName(t.projectName),
		attribute.String("telemetry.sdk.name", TracerName),
		attribute.String("telemetry.sdk.version", Version),
	}

	for k, v := range t.resourceAttributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int8:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case int16:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case int32:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case uint:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case uint8:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case uint16:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case uint32:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case uint64:
			attrs = append(attrs, attribute.Int64(k, int64(val)))
		case float32:
			attrs = append(attrs, attribute.Float64(k, float64(val)))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		case []string:
			attrs = append(attrs, attribute.StringSlice(k, val))
		default:
			serialized, err := t.serializer(val)
			if err == nil {
				attrs = append(attrs, attribute.String(k, serialized))
			}
		}
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)

	t.tracerProvider = NewJudgmentTracerProvider(
		&JudgmentTracerProviderConfig{
			FilterTracer: t.filterTracer,
		},
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(t.getSpanProcessor(ctx)),
	)

	otel.SetTracerProvider(t.tracerProvider)
	t.tracer = otel.Tracer(TracerName)

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

func resolveProjectID(client *api.Client, projectName string) (string, error) {
	logger.Info("Resolving project ID for project: %s", projectName)

	req := &models.ResolveProjectRequest{
		ProjectName: projectName,
	}

	resp, err := client.PostProjectsResolve(req)
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
