# QwenAI Thinking Mode Implementation

This project now includes full support for Qwen's thinking mode, which provides enhanced reasoning capabilities by showing the AI's thought process before delivering the final answer.

## Features

### 🧠 Thinking Mode by Default
- **Automatically enabled** for all Qwen models
- Shows the AI's reasoning process in real-time
- Provides more transparent and explainable AI responses

### 🎛️ Dynamic Control
- **Global control**: Enable/disable thinking mode for all requests
- **Per-query control**: Use `/think` or `/no_think` in your prompts
- **Automatic detection**: Works with any Qwen model that supports thinking

### 📡 Streaming Support
- Real-time streaming of thinking process
- Seamless transition from thinking to final response
- Support for tool calling and usage information

## Quick Start

### 1. Basic Usage

```go
package main

import (
    "Qwen/internal/ai"
    "Qwen/internal/config"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Create client with thinking mode enabled by default
    client := ai.NewClient(cfg.DashScopeAPIKey, cfg.DashScopeBaseURL, cfg.AIModel)
    
    // Check if thinking mode is supported
    if client.IsQwenModel() {
        fmt.Println("✅ Thinking mode enabled!")
    }
}
```

### 2. Streaming with Thinking Process

```go
client.ChatStreamWithThinking("Explain quantum computing", func(stage string, content string, isComplete bool) {
    switch stage {
    case "thinking":
        fmt.Printf("🤔 Thinking: %s", content)
    case "thinking_complete":
        fmt.Println("\n=== THINKING COMPLETE ===")
    case "streaming":
        fmt.Printf("%s", content)
    case "complete":
        fmt.Println("\n=== RESPONSE COMPLETE ===")
    }
})
```

### 3. Complete Thinking Response

```go
messages := []ai.Message{
    {Role: "user", Content: "Why is the sky blue?"},
}

response, err := client.ChatWithThinking(context.Background(), messages)
if err != nil {
    log.Fatal(err)
}

fmt.Println("🤔 Thinking Process:")
fmt.Println(response.ReasoningContent)

fmt.Println("\n💬 Final Answer:")
fmt.Println(response.AnswerContent)
```

## Advanced Features

### Prompt-Based Control

#### Enable Thinking Mode
```
What is the capital of France?/think
```
The `/think` suffix explicitly enables thinking mode for this query.

#### Disable Thinking Mode
```
What is 2+2?/no_think
```
The `/no_think` suffix disables thinking mode for this specific query.

### Global Control

```go
// Disable thinking mode globally
client.SetThinkingMode(false)

// Re-enable thinking mode globally
client.SetThinkingMode(true)
```

### Tool Calling Support

When using thinking mode, the AI can also make tool calls:

```go
client.ChatStreamWithThinking("What time is it now?", func(stage string, content string, isComplete bool) {
    switch stage {
    case "thinking":
        fmt.Printf("🤔 Thinking: %s", content)
    case "tool_call":
        fmt.Printf("🔧 Tool Call: %s\n", content)
    case "streaming":
        fmt.Printf("%s", content)
    }
})
```

## Supported Models

Thinking mode is automatically enabled for models containing "qwen" in their name:

- `qwen-plus-2025-04-28`
- `qwen-max`
- `qwen-turbo`
- `qwen-plus`
- And other Qwen variants

For non-Qwen models, the client gracefully falls back to regular chat functionality.

## Configuration

### Environment Variables

```bash
# Required
DASHSCOPE_API_KEY=your_api_key_here

# Optional (defaults shown)
DASHSCOPE_BASE_URL=https://dashscope-intl.aliyuncs.com/compatible-mode/v1
AI_MODEL=qwen-plus-2025-04-28
```

### Model Parameters

```go
// Customize thinking mode parameters
params := ai.ModelParams{
    Temperature:     0.7,
    TopK:           40,
    TopP:           0.9,
    EnableThinking: true,  // Default: true
}

client.SetParams(params)
```

## API Reference

### Client Methods

#### `NewClient(apiKey, baseURL, model) *Client`
Creates a new AI client with thinking mode enabled by default for Qwen models.

#### `SetThinkingMode(enabled bool)`
Globally enables or disables thinking mode.

#### `IsQwenModel() bool`
Returns true if the current model supports thinking mode.

#### `ChatStreamWithThinking(message string, callback func(stage, content string, isComplete bool))`
Streams the thinking process and final response with stage-based callbacks.

#### `ChatWithThinking(ctx context.Context, messages []Message) (*ThinkingResponse, error)`
Returns a complete response with both reasoning and answer content.

### Callback Stages

The streaming callback receives different stages:

- **`"thinking"`**: AI's reasoning process
- **`"thinking_complete"`**: Thinking phase finished
- **`"streaming"`**: Final response content
- **`"tool_call"`**: Tool calling information
- **`"usage"`**: Token usage statistics
- **`"complete"`**: Response finished
- **`"error"`**: Error occurred

### Response Types

#### `ThinkingResponse`
```go
type ThinkingResponse struct {
    ReasoningContent string `json:"reasoning_content"`  // AI's thinking process
    AnswerContent    string `json:"answer_content"`     // Final response
    IsComplete       bool   `json:"is_complete"`        // Response status
}
```

## Examples

### Running the Demo

```bash
cd examples
go run thinking_mode_demo.go
```

The demo showcases:
1. Basic thinking mode (enabled by default)
2. Disabling thinking with `/no_think`
3. Explicitly enabling with `/think`
4. Complete thinking responses
5. Dynamic thinking mode control
6. Tool calling support

### Integration with WebSocket

```go
// In your WebSocket handler
func handleChatMessage(client *ai.Client, message string, ws *websocket.Conn) {
    client.ChatStreamWithThinking(message, func(stage string, content string, isComplete bool) {
        response := map[string]interface{}{
            "stage":   stage,
            "content": content,
            "complete": isComplete,
        }
        
        ws.WriteJSON(response)
    })
}
```

## Benefits

### 🎯 Enhanced Transparency
- See exactly how the AI arrives at its conclusions
- Understand the reasoning behind complex responses
- Build trust in AI-generated content

### 🧩 Better Problem Solving
- Step-by-step reasoning for complex problems
- Multiple approaches considered before final answer
- More thorough and thoughtful responses

### 🔧 Educational Value
- Learn from the AI's reasoning process
- Understand different problem-solving strategies
- Improve critical thinking skills

## Troubleshooting

### Common Issues

#### Thinking Mode Not Working
- Ensure you're using a Qwen model
- Check that `EnableThinking` is set to `true`
- Verify your API key has access to thinking-enabled models

#### No Reasoning Content
- Some simple queries may not generate reasoning
- Try adding `/think` to force thinking mode
- Check if the model supports thinking mode

#### Streaming Issues
- Ensure your callback handles all stages properly
- Check for network connectivity issues
- Verify API rate limits

### Debug Mode

Enable debug logging to see detailed API interactions:

```go
// Add logging to your callback
client.ChatStreamWithThinking("Test message", func(stage string, content string, isComplete bool) {
    fmt.Printf("DEBUG: Stage=%s, Content='%s', Complete=%t\n", stage, content, isComplete)
    // ... your normal handling
})
```

## Contributing

To extend the thinking mode functionality:

1. **Add new stages**: Extend the callback stage types
2. **Support new models**: Add model detection logic
3. **Enhance tool calling**: Implement additional tool integrations
4. **Improve error handling**: Add more robust error recovery

## License

This project is licensed under the same terms as the main QwenAI project.

---

**Note**: Thinking mode requires Qwen models released in April 2025 or later. For older models, the client will gracefully fall back to regular chat functionality.
