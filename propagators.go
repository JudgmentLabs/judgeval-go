package judgeval

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type customerIDContextKey struct{}
type sessionIDContextKey struct{}

func contextWithCustomerID(ctx context.Context, customerID string) context.Context {
	return context.WithValue(ctx, customerIDContextKey{}, customerID)
}

func contextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDContextKey{}, sessionID)
}

func customerIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(customerIDContextKey{}).(string)
	return value, ok
}

func sessionIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(sessionIDContextKey{}).(string)
	return value, ok
}

type CustomerIDProcessor struct{}

func (p *CustomerIDProcessor) ForceFlush(ctx context.Context) error { return nil }
func (p *CustomerIDProcessor) OnEnd(s sdktrace.ReadOnlySpan)        {}
func (p *CustomerIDProcessor) Shutdown(ctx context.Context) error   { return nil }
func (p *CustomerIDProcessor) OnStart(parentContext context.Context, span sdktrace.ReadWriteSpan) {
	if customerID, ok := customerIDFromContext(parentContext); ok {
		span.SetAttributes(attribute.String(AttributeKeysJudgmentCustomerID, customerID))
	}
}

type SessionIDProcessor struct{}

func (p *SessionIDProcessor) ForceFlush(ctx context.Context) error { return nil }
func (p *SessionIDProcessor) OnEnd(s sdktrace.ReadOnlySpan)        {}
func (p *SessionIDProcessor) Shutdown(ctx context.Context) error   { return nil }
func (p *SessionIDProcessor) OnStart(parentContext context.Context, span sdktrace.ReadWriteSpan) {
	if sessionID, ok := sessionIDFromContext(parentContext); ok {
		span.SetAttributes(attribute.String(AttributeKeysJudgmentSessionID, sessionID))
	}
}

func lifecycleSpanProcessors() []sdktrace.SpanProcessor {
	return []sdktrace.SpanProcessor{
		&CustomerIDProcessor{},
		&SessionIDProcessor{},
	}
}
