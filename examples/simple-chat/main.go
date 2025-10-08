package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JudgmentLabs/judgeval-go/pkg/data"
	"github.com/JudgmentLabs/judgeval-go/pkg/scorers"
	"github.com/JudgmentLabs/judgeval-go/pkg/scorers/api_scorers"
	"github.com/JudgmentLabs/judgeval-go/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

type ChatClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
	tracer  *tracer.Tracer
}

func NewChatClient(apiKey string) *ChatClient {
	return &ChatClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1/chat/completions",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ChatClient) SetTracer(t *tracer.Tracer) {
	c.tracer = t
}

func (c *ChatClient) SendMessage(ctx context.Context, messages []ChatMessage) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Model:       "gpt-3.5-turbo",
		Messages:    messages,
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	if c.tracer != nil {
		if span := trace.SpanFromContext(ctx); span != nil {
			c.tracer.SetInput(span, reqBody)

			c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIRequestModel, reqBody.Model)
			c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIRequestTemperature, reqBody.Temperature)
			c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIRequestMaxTokens, reqBody.MaxTokens)

			if len(messages) > 0 {
				lastMessage := messages[len(messages)-1]
				if lastMessage.Role == "user" {
					c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIPrompt, lastMessage.Content)
				}
			}
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if c.tracer != nil {
		if span := trace.SpanFromContext(ctx); span != nil {
			c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIResponseModel, chatResp.Model)
			c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIUsageInputTokens, chatResp.Usage.PromptTokens)
			c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIUsageOutputTokens, chatResp.Usage.CompletionTokens)

			if len(chatResp.Choices) > 0 {
				c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAICompletion, chatResp.Choices[0].Message.Content)
				c.tracer.SetAttribute(span, tracer.AttributeKeys.GenAIResponseFinishReasons, chatResp.Choices[0].FinishReason)
			}
		}
	}

	return &chatResp, nil
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable is required")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your_api_key_here")
		os.Exit(1)
	}

	chatClient := NewChatClient(apiKey)

	if os.Getenv("JUDGMENT_API_URL") != "" && os.Getenv("JUDGMENT_API_KEY") != "" {
		t, err := tracer.NewTracer(
			tracer.WithConfiguration(tracer.NewTracerConfiguration(
				tracer.WithProjectName("default_project"),
			)),
		)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize tracer: %v\n", err)
		} else {
			chatClient.SetTracer(t)
			defer t.Shutdown(context.Background())
		}
	}

	fmt.Println("🤖 Simple Chat with OpenAI")
	fmt.Println("Type 'quit' or 'exit' to end the conversation")
	fmt.Println("Type 'clear' to clear conversation history")
	fmt.Println("----------------------------------------")

	var messages []ChatMessage
	scanner := bufio.NewScanner(os.Stdin)
	messageCount := 0

	ctx := context.Background()
	var parentSpan trace.Span
	if chatClient.tracer != nil {
		parentSpan, ctx = chatClient.tracer.Span(ctx, "chat-session")
		chatClient.tracer.SetGeneralSpan(parentSpan)
		chatClient.tracer.SetAttribute(parentSpan, "chat.session.start_time", time.Now().Unix())
		defer parentSpan.End()
	}

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		if userInput == "quit" || userInput == "exit" {
			fmt.Println("Goodbye! 👋")
			if chatClient.tracer != nil && parentSpan != nil {
				chatClient.tracer.SetAttribute(parentSpan, "chat.session.end_time", time.Now().Unix())
				chatClient.tracer.SetAttribute(parentSpan, "chat.session.message_count", messageCount)
			}
			break
		}

		if userInput == "clear" {
			messages = nil
			fmt.Println("Conversation history cleared.")
			continue
		}

		messages = append(messages, ChatMessage{
			Role:    "user",
			Content: userInput,
		})

		messageCount++
		messageCtx := ctx
		var span trace.Span
		if chatClient.tracer != nil {
			span, messageCtx = chatClient.tracer.Span(ctx, "OPENAI_API_CALL")
			chatClient.tracer.SetLLMSpan(span)
			chatClient.tracer.SetAttribute(span, "chat.message.number", messageCount)
			defer span.End()
		}

		fmt.Print("Bot: ")
		resp, err := chatClient.SendMessage(messageCtx, messages)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if len(resp.Choices) == 0 {
			fmt.Println("No response received from OpenAI")
			continue
		}

		botMessage := resp.Choices[0].Message.Content
		fmt.Println(botMessage)

		messages = append(messages, ChatMessage{
			Role:    "assistant",
			Content: botMessage,
		})

		if chatClient.tracer != nil {
			if span := trace.SpanFromContext(messageCtx); span != nil {
				chatClient.tracer.SetOutput(span, botMessage)
			}
		}

		// Async evaluation for answer relevancy
		if chatClient.tracer != nil {
			go func() {
				// Create answer relevancy scorer
				scorer := api_scorers.NewAnswerRelevancyScorer(
					scorers.WithThreshold(0.7),
					scorers.WithModel("gpt-3.5-turbo"),
				)

				// Create example for evaluation
				example := data.NewExample(
					data.WithName(fmt.Sprintf("chat-message-%d", messageCount)),
					data.WithProperty("input", userInput),
					data.WithProperty("actual_output", botMessage),
				)

				// Trigger async evaluation
				chatClient.tracer.AsyncEvaluate(messageCtx, scorer, example, "gpt-3.5-turbo")
			}()
		}

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}
