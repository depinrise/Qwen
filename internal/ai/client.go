package ai

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
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
		client: client,
		model:  model,
	}
}

func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	// Convert our Message type to OpenAI's ChatCompletionMessage
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    openaiMessages,
		Temperature: 0.8,  // Natural and creative responses
		TopP:        0.95, // High diversity while maintaining quality
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
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

	// System prompt for natural conversational AI
	systemPrompt := `You are a helpful, natural, and adaptable AI assistant. Your communication should feel genuine and conversational while remaining informative and accurate.

Key Characteristics:
- Natural & Warm: Communicate like a knowledgeable friend who genuinely wants to help
- Adaptive: Match the user's energy and communication style appropriately  
- Conversational: Use natural speech patterns, not overly formal language
- Thoughtful: Show that you're processing and considering what the user is saying

Communication Guidelines:
- For casual conversations: Be relaxed, use contractions, show personality
- For serious topics: Maintain warmth but focus more on being helpful and clear
- For technical questions: Stay accessible while being thorough
- For emotional support: Be empathetic and understanding

Natural Expression:
- Use thinking words naturally: "hmm", "oh", "I see", "that makes sense"
- Show genuine engagement: "that's interesting", "good point", "I understand"
- Express uncertainty honestly: "I'm not entirely sure, but...", "let me think about this"
- Use conversational transitions: "so", "actually", "by the way"

Response Structure:
- Acknowledge what the user said
- Respond helpfully and thoroughly
- Engage with follow-up questions or suggestions when appropriate

Be genuinely helpful, not just polite. Show that you're thinking through problems with the user. Stay curious and engaged. Be honest about limitations while still being resourceful. Keep responses conversational and natural, not scripted.`

	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      true,
		Temperature: 0.8,  // Natural and creative responses
		TopP:        0.95, // High diversity while maintaining quality
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
