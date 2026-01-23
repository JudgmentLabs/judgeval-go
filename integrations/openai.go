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

// OpenAIMiddleware returns an OpenAI client middleware that instruments API calls with Judgment tracing.
// It can be used with github.com/openai/openai-go's option.WithMiddleware.
func OpenAIMiddleware(tracer judgeval.TracerInterface) func(req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
	return func(req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
		return openaiInstrumentedRoundTrip(tracer, req, next)
	}
}

func openaiInstrumentedRoundTrip(tracer judgeval.TracerInterface, req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
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

	spanName := openaiGetSpanName(req.URL.Path)

	ctx, span := tracer.StartSpan(ctx, spanName)
	defer span.End()

	span.SetAttributes(attribute.String(judgeval.AttributeKeysJudgmentSpanKind, "llm"))

	if requestData != nil {
		setOpenAIRequestAttributes(span, requestData)
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
		resp.Body = newOpenAIStreamingResponseBody(resp.Body, span)
	} else {
		if resp.Body != nil {
			responseBody, err := io.ReadAll(resp.Body)
			if err == nil {
				resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
				setOpenAIResponseAttributes(span, responseBody)
			}
		}
	}

	return resp, nil
}

func openaiGetSpanName(path string) string {
	switch {
	case strings.Contains(path, "/chat/completions"):
		return "openai.chat.completions"
	case strings.Contains(path, "/responses"):
		return "openai.responses"
	default:
		return "OPENAI_API_CALL"
	}
}

func setOpenAIRequestAttributes(span trace.Span, data map[string]interface{}) {
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
	if freqPenalty, ok := data["frequency_penalty"].(float64); ok {
		span.SetAttributes(attribute.Float64(judgeval.AttributeKeysGenAIRequestFrequencyPenalty, freqPenalty))
	}
	if presPenalty, ok := data["presence_penalty"].(float64); ok {
		span.SetAttributes(attribute.Float64(judgeval.AttributeKeysGenAIRequestPresencePenalty, presPenalty))
	}
	if stop, ok := data["stop"].([]interface{}); ok {
		stopStrs := make([]string, 0, len(stop))
		for _, s := range stop {
			if str, ok := s.(string); ok {
				stopStrs = append(stopStrs, str)
			}
		}
		if len(stopStrs) > 0 {
			span.SetAttributes(attribute.StringSlice(judgeval.AttributeKeysGenAIRequestStopSequences, stopStrs))
		}
	}
}

func setOpenAIResponseAttributes(span trace.Span, body []byte) {
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

	if usage, ok := data["usage"].(map[string]interface{}); ok {
		if inputTokens, ok := usage["input_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageInputTokens, int(inputTokens)))
		} else if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageInputTokens, int(promptTokens)))
		}

		if outputTokens, ok := usage["output_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageOutputTokens, int(outputTokens)))
		} else if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageOutputTokens, int(completionTokens)))
		}

		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageTotalTokens, int(totalTokens)))
		}

		if inputDetails, ok := usage["input_tokens_details"].(map[string]interface{}); ok {
			if cachedTokens, ok := inputDetails["cached_tokens"].(float64); ok {
				span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageCacheReadInputTokens, int(cachedTokens)))
			}
		} else if promptDetails, ok := usage["prompt_tokens_details"].(map[string]interface{}); ok {
			if cachedTokens, ok := promptDetails["cached_tokens"].(float64); ok {
				span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageCacheReadInputTokens, int(cachedTokens)))
			}
		}
	}

	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		finishReasons := make([]string, 0, len(choices))
		var completions []string

		for _, choice := range choices {
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				if reason, ok := choiceMap["finish_reason"].(string); ok && reason != "" {
					finishReasons = append(finishReasons, reason)
				}
				if message, ok := choiceMap["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						completions = append(completions, content)
					}
				}
				if text, ok := choiceMap["text"].(string); ok {
					completions = append(completions, text)
				}
			}
		}

		if len(finishReasons) > 0 {
			span.SetAttributes(attribute.StringSlice(judgeval.AttributeKeysGenAIResponseFinishReasons, finishReasons))
		}
		if len(completions) > 0 {
			if jsonBytes, err := json.Marshal(completions); err == nil {
				span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAICompletion, string(jsonBytes)))
			}
		}
	}

	if output, ok := data["output"].([]interface{}); ok && len(output) > 0 {
		var texts []string
		for _, item := range output {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if content, ok := itemMap["content"].([]interface{}); ok {
					for _, c := range content {
						if cMap, ok := c.(map[string]interface{}); ok {
							if text, ok := cMap["text"].(string); ok {
								texts = append(texts, text)
							}
						}
					}
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

type openaiStreamingResponseBody struct {
	body        io.ReadCloser
	span        trace.Span
	accumulated strings.Builder
}

func newOpenAIStreamingResponseBody(body io.ReadCloser, span trace.Span) *openaiStreamingResponseBody {
	return &openaiStreamingResponseBody{
		body: body,
		span: span,
	}
}

func (s *openaiStreamingResponseBody) Read(p []byte) (int, error) {
	n, err := s.body.Read(p)
	if n > 0 {
		s.accumulated.Write(p[:n])
	}
	if err == io.EOF {
		s.finalizeSpan()
	}
	return n, err
}

func (s *openaiStreamingResponseBody) Close() error {
	if s.accumulated.Len() > 0 {
		s.finalizeSpan()
	}
	return s.body.Close()
}

func (s *openaiStreamingResponseBody) finalizeSpan() {
	lines := strings.Split(s.accumulated.String(), "\n")
	var contentParts []string

	for _, line := range lines {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			continue
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Extract content from choices delta
		if choices, ok := chunk["choices"].([]interface{}); ok {
			for _, choice := range choices {
				if choiceMap, ok := choice.(map[string]interface{}); ok {
					if delta, ok := choiceMap["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							contentParts = append(contentParts, content)
						}
					}
				}
			}
		}

		if usage, ok := chunk["usage"].(map[string]interface{}); ok {
			if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
				s.span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageInputTokens, int(promptTokens)))
			}
			if completionTokens, ok := usage["completion_tokens"].(float64); ok {
				s.span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageOutputTokens, int(completionTokens)))
			}
			if totalTokens, ok := usage["total_tokens"].(float64); ok {
				s.span.SetAttributes(attribute.Int(judgeval.AttributeKeysGenAIUsageTotalTokens, int(totalTokens)))
			}
		}

		if model, ok := chunk["model"].(string); ok {
			s.span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAIResponseModel, model))
		}
	}

	if len(contentParts) > 0 {
		fullContent := strings.Join(contentParts, "")
		s.span.SetAttributes(attribute.String(judgeval.AttributeKeysGenAICompletion, fullContent))
	}
}
