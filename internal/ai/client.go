package ai

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client  *openai.Client
	model   string
	apiKey  string
	baseURL string
	params  ModelParams
}

// ModelParams controls sampling behavior
type ModelParams struct {
	Temperature float64 `json:"temperature"`
	TopK        int     `json:"top_k"`
	TopP        float64 `json:"top_p"`
}

var defaultParams = ModelParams{
	Temperature: 0.75,
	TopK:        45,
	TopP:        0.92,
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewClient(apiKey, baseURL, model string) *Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	client := openai.NewClientWithConfig(config)

	return &Client{
		client:  client,
		model:   model,
		apiKey:  apiKey,
		baseURL: baseURL,
		params:  defaultParams,
	}
}

// SetParams allows overriding default model params at runtime
func (c *Client) SetParams(p ModelParams) {
	c.params = p
}

func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	// Use streaming collection to support models that require stream=true (e.g., qwen-omni-turbo)
	// Convert our Message type to OpenAI's ChatCompletionMessage
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Always stream for compatibility and robustness
	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    openaiMessages,
		Stream:      true,
		Temperature: float32(c.params.Temperature),
		TopP:        float32(c.params.TopP),
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		// Improve error for omni models
		if strings.Contains(strings.ToLower(err.Error()), "only support with stream=true") {
			return "", fmt.Errorf("model requires streaming; retry later: %w", err)
		}
		return "", fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	var full string
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("stream recv error: %w", err)
		}
		if len(resp.Choices) > 0 {
			full += resp.Choices[0].Delta.Content
		}
	}

	if strings.TrimSpace(full) == "" {
		return "", fmt.Errorf("empty response from stream")
	}
	return full, nil
}

// ChatStream streams the AI response with callback for each chunk
func (c *Client) ChatStream(userMessage string, callback func(chunk string, isComplete bool)) {
	ctx := context.Background()

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		callback(fmt.Sprintf("Error: %v", err), true)
		return
	}
	defer stream.Close()

	var fullResponse string

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// Stream completed successfully
				callback(fullResponse, true)
				return
			}
			callback(fmt.Sprintf("Stream error: %v", err), true)
			return
		}

		if len(response.Choices) > 0 {
			chunk := response.Choices[0].Delta.Content
			if chunk != "" {
				fullResponse += chunk
				callback(chunk, false)

				// Add small delay to simulate thinking/processing
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

// ChatStreamWithThinking provides enhanced streaming with actual thinking process
func (c *Client) ChatStreamWithThinking(userMessage string, callback func(stage string, content string, isComplete bool)) {
	ctx := context.Background()

	messages := []openai.ChatCompletionMessage{{Role: "user", Content: userMessage}}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      true,
		Temperature: float32(c.params.Temperature),
		TopP:        float32(c.params.TopP),
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		callback("error", fmt.Sprintf("Error: %v", err), true)
		return
	}
	defer stream.Close()

	var fullResponse string

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// Stream completed successfully
				callback("complete", fullResponse, true)
				return
			}
			callback("error", fmt.Sprintf("Stream error: %v", err), true)
			return
		}

		if len(response.Choices) > 0 {
			chunk := response.Choices[0].Delta.Content
			if chunk != "" {
				fullResponse += chunk
				// Stream directly without complex parsing
				callback("streaming", chunk, false)
				time.Sleep(30 * time.Millisecond) // Realistic typing delay
			}
		}
	}
}
