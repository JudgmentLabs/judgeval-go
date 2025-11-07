package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	v1 "github.com/JudgmentLabs/judgeval-go/v1"
)

func main() {
	client, err := v1.NewClient(
		v1.WithAPIKey(os.Getenv("JUDGMENT_API_KEY")),
		v1.WithOrganizationID(os.Getenv("JUDGMENT_ORG_ID")),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	tracer, err := client.Tracer.Create(ctx, v1.TracerCreateParams{
		ProjectName: "v1-example-project",
		Initialize:  v1.Bool(true),
	})
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}
	defer tracer.Shutdown(ctx)

	fmt.Println("Tracer initialized successfully")

	ctx, span := tracer.Span(ctx, "example-span")
	tracer.SetLLMSpan(span)
	tracer.SetInput(span, map[string]string{
		"prompt": "What is AI?",
	})
	tracer.SetOutput(span, map[string]string{
		"response": "Artificial Intelligence is...",
	})

	scorer := client.Scorers.BuiltIn.Faithfulness(v1.FaithfulnessScorerParams{
		Threshold: v1.Float(0.8),
	})

	example := v1.NewExample(v1.ExampleParams{
		Name: v1.String("test-example"),
		Properties: map[string]interface{}{
			"context":       "AI is a branch of computer science",
			"actual_output": "Artificial Intelligence is...",
		},
	})

	tracer.AsyncEvaluate(ctx, scorer, example, nil)

	span.End()

	flushCtx, flushCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer flushCancel()
	err = tracer.ForceFlush(flushCtx)
	if err != nil {
		log.Fatalf("Failed to force flush tracer: %v", err)
	}
	fmt.Println("Tracer force flushed successfully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	err = tracer.Shutdown(shutdownCtx)
	if err != nil {
		log.Fatalf("Failed to shutdown tracer: %v", err)
	}
	fmt.Println("Tracer shut down successfully")

}
