package main

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "github.com/JudgmentLabs/judgeval-go"
	"github.com/openai/openai-go"
	oaioption "github.com/openai/openai-go/option"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	client, err := v1.NewJudgeval(
		"simple_chat",
		v1.WithAPIKey(os.Getenv("JUDGMENT_API_KEY")),
		v1.WithOrganizationID(os.Getenv("JUDGMENT_ORG_ID")),
	)
	if err != nil {
		fmt.Printf("Error: Failed to create Judgment client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	tracer, err := client.Tracer.Create(ctx, v1.TracerCreateParams{})
	if err != nil {
		fmt.Printf("Error: Failed to create tracer: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		tracer.Shutdown(shutdownCtx)
	}()

	rootCtx, rootSpan := tracer.Span(ctx, "chat-session")
	defer rootSpan.End()

	rootCtx = tracer.SetCustomerID(rootCtx, "customer-123")
	rootCtx = tracer.SetSessionID(rootCtx, "session-456")

	userInput := "What is the capital of France?"
	output := callLLM(rootCtx, tracer, apiKey, userInput)

	fmt.Printf("Question: %s\n", userInput)
	fmt.Printf("Answer: %s\n", output)

	evalCtx, evalSpan := tracer.Span(rootCtx, "evaluation")
	defer evalSpan.End()

	scorer := client.Scorers.BuiltIn.AnswerCorrectness(v1.AnswerCorrectnessScorerParams{
		Threshold: v1.Float(0.7),
	})

	example := v1.NewExample(v1.ExampleParams{
		"input":           userInput,
		"actual_output":   output,
		"expected_output": "Paris",
	})

	tracer.AsyncEvaluate(evalCtx, scorer, example)

	fmt.Println("\nEvaluation submitted successfully")
}

func callLLM(ctx context.Context, tracer *v1.Tracer, apiKey string, userInput string) string {
	llmCtx, llmSpan := tracer.Span(ctx, "llm-call")
	defer llmSpan.End()

	tracer.SetLLMSpan(llmSpan)
	tracer.SetInput(llmSpan, userInput)

	openaiClient := openai.NewClient(oaioption.WithAPIKey(apiKey))

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a helpful assistant."),
		openai.UserMessage(userInput),
	}

	response, err := openaiClient.Chat.Completions.New(llmCtx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModelGPT4o,
		Messages: messages,
	})
	if err != nil {
		fmt.Printf("Error: Chat completion failed: %v\n", err)
		os.Exit(1)
	}

	output := response.Choices[0].Message.Content
	tracer.SetOutput(llmSpan, output)

	return output
}
