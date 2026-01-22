// Package integrations provides middleware for instrumenting LLM client libraries with Judgment tracing.
//
// OpenAI Usage:
//
//	import (
//	    "github.com/openai/openai-go"
//	    "github.com/openai/openai-go/option"
//	    "github.com/JudgmentLabs/judgeval-go/integrations"
//	)
//
//	client := openai.NewClient(
//	    option.WithAPIKey(apiKey),
//	    option.WithMiddleware(integrations.OpenAIMiddleware(tracer)),
//	)
//
// Anthropic Usage:
//
//	import (
//	    "github.com/anthropics/anthropic-sdk-go"
//	    "github.com/anthropics/anthropic-sdk-go/option"
//	    "github.com/JudgmentLabs/judgeval-go/integrations"
//	)
//
//	client := anthropic.NewClient(
//	    option.WithAPIKey(apiKey),
//	    option.WithMiddleware(integrations.AnthropicMiddleware(tracer)),
//	)
package integrations
