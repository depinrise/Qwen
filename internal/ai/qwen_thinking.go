package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// QwenThinkingClient handles Qwen-specific thinking mode functionality
type QwenThinkingClient struct {
	apiKey  string
	baseURL string
	model   string
	params  ModelParams
}

// QwenThinkingRequest represents the request structure for Qwen thinking mode
type QwenThinkingRequest struct {
	Model     string                 `json:"model"`
	Messages  []QwenMessage          `json:"messages"`
	Stream    bool                   `json:"stream"`
	ExtraBody map[string]interface{} `json:"-"`
}

// QwenMessage represents a message in the Qwen API format
type QwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// QwenThinkingStreamResponse represents the streaming response from Qwen thinking mode
type QwenThinkingStreamResponse struct {
	Choices []QwenStreamChoice `json:"choices"`
	Usage   *QwenUsage         `json:"usage,omitempty"`
}

// QwenStreamChoice represents a choice in the streaming response
type QwenStreamChoice struct {
	Delta QwenStreamDelta `json:"delta"`
}

// QwenStreamDelta represents the delta content in streaming responses
type QwenStreamDelta struct {
	Content          string         `json:"content,omitempty"`
	ReasoningContent string         `json:"reasoning_content,omitempty"`
	ToolCalls        []QwenToolCall `json:"tool_calls,omitempty"`
}

// QwenToolCall represents tool calling information
type QwenToolCall struct {
	ID       string       `json:"id,omitempty"`
	Function QwenFunction `json:"function,omitempty"`
	Index    int          `json:"index,omitempty"`
}

// QwenFunction represents function information in tool calls
type QwenFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// QwenUsage represents token usage information
type QwenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewQwenThinkingClient creates a new client specifically for Qwen thinking mode
func NewQwenThinkingClient(apiKey, baseURL, model string) *QwenThinkingClient {
	return &QwenThinkingClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		params:  defaultParams,
	}
}

// ChatWithThinkingStream streams the thinking process and final response
func (q *QwenThinkingClient) ChatWithThinkingStream(ctx context.Context, messages []QwenMessage, callback func(stage string, content string, isComplete bool)) error {
	// Process thinking prompt controls for the last user message
	var thinkingEnabled bool
	if len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		if lastMessage.Role == "user" {
			_, thinkingEnabled = q.processThinkingPrompt(lastMessage.Content)
		}
	}

	// Build request
	reqBody := QwenThinkingRequest{
		Model:    q.model,
		Messages: messages,
		Stream:   true,
	}

	// Add thinking mode support
	if thinkingEnabled {
		reqBody.ExtraBody = map[string]interface{}{
			"enable_thinking": true,
		}
	}

	// Marshal request body
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", q.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+q.apiKey)

	// Execute request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Process streaming response
	reader := bufio.NewReader(resp.Body)
	var reasoningContent string
	var answerContent string
	isAnswering := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove "data: " prefix if present
		line = strings.TrimPrefix(line, "data: ")

		// Skip "[DONE]" message
		if line == "[DONE]" {
			break
		}

		// Parse JSON response
		var streamResp QwenThinkingStreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			continue // Skip malformed JSON
		}

		if len(streamResp.Choices) > 0 {
			delta := streamResp.Choices[0].Delta

			// Handle reasoning content (thinking process)
			if delta.ReasoningContent != "" {
				reasoningContent += delta.ReasoningContent
				callback("thinking", delta.ReasoningContent, false)
			}

			// Handle regular content (final response)
			if delta.Content != "" {
				if !isAnswering {
					callback("thinking_complete", "", false)
					isAnswering = true
				}
				answerContent += delta.Content
				callback("streaming", delta.Content, false)
			}

			// Handle tool calls if present
			if len(delta.ToolCalls) > 0 {
				for _, toolCall := range delta.ToolCalls {
					toolInfo := fmt.Sprintf("Tool: %s", toolCall.Function.Name)
					if toolCall.Function.Arguments != "" {
						toolInfo += fmt.Sprintf(" - Args: %s", toolCall.Function.Arguments)
					}
					callback("tool_call", toolInfo, false)
				}
			}
		}

		// Handle usage information
		if streamResp.Usage != nil {
			usageInfo := fmt.Sprintf("Tokens: %d prompt, %d completion, %d total",
				streamResp.Usage.PromptTokens,
				streamResp.Usage.CompletionTokens,
				streamResp.Usage.TotalTokens)
			callback("usage", usageInfo, false)
		}
	}

	// Final callback with complete response
	callback("complete", answerContent, true)
	return nil
}

// ChatWithThinking provides a complete thinking response with both reasoning and answer
func (q *QwenThinkingClient) ChatWithThinking(ctx context.Context, messages []QwenMessage) (*ThinkingResponse, error) {
	response := &ThinkingResponse{}

	err := q.ChatWithThinkingStream(ctx, messages, func(stage string, content string, isComplete bool) {
		switch stage {
		case "thinking":
			response.ReasoningContent += content
		case "streaming":
			response.AnswerContent += content
		case "complete":
			response.IsComplete = true
		}
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// processThinkingPrompt checks for /think or /no_think in the prompt and adjusts thinking mode
func (q *QwenThinkingClient) processThinkingPrompt(content string) (string, bool) {
	content = strings.TrimSpace(content)

	// Check for /no_think to disable thinking mode
	if strings.HasSuffix(content, "/no_think") {
		content = strings.TrimSuffix(content, "/no_think")
		return strings.TrimSpace(content), false
	}

	// Check for /think to explicitly enable thinking mode
	if strings.HasSuffix(content, "/think") {
		content = strings.TrimSuffix(content, "/think")
		return strings.TrimSpace(content), true
	}

	// Default to current thinking mode setting
	return content, q.params.EnableThinking
}

// SetParams allows overriding default model params at runtime
func (q *QwenThinkingClient) SetParams(p ModelParams) {
	q.params = p
}

// SetThinkingMode enables or disables thinking mode
func (q *QwenThinkingClient) SetThinkingMode(enabled bool) {
	q.params.EnableThinking = enabled
}
