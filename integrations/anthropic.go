package integrations

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	judgeval "github.com/JudgmentLabs/judgeval-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// AnthropicMiddleware returns an Anthropic client middleware that instruments API calls with Judgment tracing.
// It can be used with github.com/anthropics/anthropic-sdk-go's option.WithMiddleware.
func AnthropicMiddleware(tracer judgeval.JudgevalTracer) func(req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
	return func(req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
		return anthropicInstrumentedRoundTrip(tracer, req, next)
	}
}

func anthropicInstrumentedRoundTrip(tracer judgeval.JudgevalTracer, req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
	ctx := req.Context()

	var requestBody []byte
	var requestData map[string]interface{}
	isStreaming := false

	if req.Body != nil {
		var err error
		requestBody, err = io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			if err := json.Unmarshal(requestBody, &requestData); err == nil {
				if stream, ok := requestData["stream"].(bool); ok && stream {
					isStreaming = true
				}
			}
		}
	}

	spanName := anthropicGetSpanName(req.URL.Path)

	ctx, span := tracer.StartSpan(ctx, spanName)
	defer span.End()

	span.SetAttributes(attribute.String(judgeval.AttributeKeysJudgmentSpanKind, "llm"))
	span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAISystem, "anthropic"))

	if requestData != nil {
		setAnthropicRequestAttributes(span, requestData)
	}

	if len(requestBody) > 0 {
		span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAIPrompt, string(requestBody)))
	}

	req = req.WithContext(ctx)

	resp, err := next(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return resp, err
	}

	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
	}

	if isStreaming {
		resp.Body = newAnthropicStreamingResponseBody(resp.Body, span)
	} else {
		if resp.Body != nil {
			responseBody, err := io.ReadAll(resp.Body)
			if err == nil {
				resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
				setAnthropicResponseAttributes(span, responseBody)
			}
		}
	}

	return resp, nil
}

func anthropicGetSpanName(path string) string {
	if strings.Contains(path, "/messages") {
		return "anthropic.messages"
	}
	return "ANTHROPIC_API_CALL"
}

func setAnthropicRequestAttributes(span trace.Span, data map[string]interface{}) {
	if model, ok := data["model"].(string); ok {
		span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAIRequestModel, model))
	}
	if temp, ok := data["temperature"].(float64); ok {
		span.SetAttributes(attribute.Float64(judgeval.AttributeKeysGenAIRequestTemperature, temp))
	}
	if maxTokens, ok := data["max_tokens"].(float64); ok {
		span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIRequestMaxTokens, int(maxTokens)))
	}
	if topP, ok := data["top_p"].(float64); ok {
		span.SetAttributes(attribute.Float64(judgeval.AttributeKeysGenAIRequestTopP, topP))
	}
	if topK, ok := data["top_k"].(float64); ok {
		span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIRequestTopK, int(topK)))
	}
	if stopSequences, ok := data["stop_sequences"].([]interface{}); ok {
		stopStrs := make([]string, 0, len(stopSequences))
		for _, s := range stopSequences {
			if str, ok := s.(string); ok {
				stopStrs = append(stopStrs, str)
			}
		}
		if len(stopStrs) > 0 {
			span.SetAttributes(attribute.StringSlice(judgeval.AttributeKeysGenAIRequestStopSequences, stopStrs))
		}
	}
}

func setAnthropicResponseAttributes(span trace.Span, body []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return
	}

	if id, ok := data["id"].(string); ok {
		span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAIResponseID, id))
	}
	if model, ok := data["model"].(string); ok {
		span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAIResponseModel, model))
	}
	if stopReason, ok := data["stop_reason"].(string); ok {
		span.SetAttributes(attribute.StringSlice(judgeval.AttributeKeysGenAIResponseFinishReasons, []string{stopReason}))
	}

	if usage, ok := data["usage"].(map[string]interface{}); ok {
		if inputTokens, ok := usage["input_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageInputTokens, int(inputTokens)))
		}
		if outputTokens, ok := usage["output_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageOutputTokens, int(outputTokens)))
		}
		if cacheCreation, ok := usage["cache_creation_input_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageCacheCreationInputTokens, int(cacheCreation)))
		}
		if cacheRead, ok := usage["cache_read_input_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageCacheReadInputTokens, int(cacheRead)))
		}
	}

	if content, ok := data["content"].([]interface{}); ok && len(content) > 0 {
		var texts []string
		for _, item := range content {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
		if len(texts) > 0 {
			if jsonBytes, err := json.Marshal(texts); err == nil {
				span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAICompletion, string(jsonBytes)))
			}
		}
	}
}

type anthropicStreamingResponseBody struct {
	body        io.ReadCloser
	span        trace.Span
	accumulated strings.Builder
}

func newAnthropicStreamingResponseBody(body io.ReadCloser, span trace.Span) *anthropicStreamingResponseBody {
	return &anthropicStreamingResponseBody{
		body: body,
		span: span,
	}
}

func (s *anthropicStreamingResponseBody) Read(p []byte) (int, error) {
	n, err := s.body.Read(p)
	if n > 0 {
		s.accumulated.Write(p[:n])
	}
	if err == io.EOF {
		s.finalizeSpan()
	}
	return n, err
}

func (s *anthropicStreamingResponseBody) Close() error {
	if s.accumulated.Len() > 0 {
		s.finalizeSpan()
	}
	return s.body.Close()
}

func (s *anthropicStreamingResponseBody) finalizeSpan() {
	lines := strings.Split(s.accumulated.String(), "\n")
	var contentParts []string

	for _, line := range lines {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		eventType, _ := event["type"].(string)

		switch eventType {
		case "content_block_delta":
			if delta, ok := event["delta"].(map[string]interface{}); ok {
				if text, ok := delta["text"].(string); ok {
					contentParts = append(contentParts, text)
				}
			}
		case "message_delta":
			if delta, ok := event["delta"].(map[string]interface{}); ok {
				if stopReason, ok := delta["stop_reason"].(string); ok {
					s.span.SetAttributes(attribute.StringSlice(judgeval.AttributeKeysGenAIResponseFinishReasons, []string{stopReason}))
				}
			}
			if usage, ok := event["usage"].(map[string]interface{}); ok {
				if outputTokens, ok := usage["output_tokens"].(float64); ok {
					s.span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageOutputTokens, int(outputTokens)))
				}
			}
		case "message_start":
			if message, ok := event["message"].(map[string]interface{}); ok {
				if model, ok := message["model"].(string); ok {
					s.span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAIResponseModel, model))
				}
				if id, ok := message["id"].(string); ok {
					s.span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAIResponseID, id))
				}
				if usage, ok := message["usage"].(map[string]interface{}); ok {
					if inputTokens, ok := usage["input_tokens"].(float64); ok {
						s.span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageInputTokens, int(inputTokens)))
					}
				}
			}
		}
	}

	if len(contentParts) > 0 {
		fullContent := strings.Join(contentParts, "")
		s.span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAICompletion, fullContent))
	}
}
