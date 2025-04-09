package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	api *openai.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		api: openai.NewClient(apiKey),
	}
}

// CompletionRequest contains info the LLM needs to generate something
type CompletionRequest struct {
	Prompt string
	System string // Optional: system message (e.g., "You are a helpful trainer AI")
}

// CompletionResponse is a structured response
type CompletionResponse struct {
	Text string
}

func (c *Client) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	var messages []openai.ChatCompletionMessage

	if req.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.System,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Prompt,
	})

	resp, err := c.api.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
	})

	if err != nil {
		return CompletionResponse{}, fmt.Errorf("LLM error: %w", err)
	}

	return CompletionResponse{Text: resp.Choices[0].Message.Content}, nil
}

type Interface interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}
