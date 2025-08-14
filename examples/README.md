# QwenAI Thinking Mode Examples

This directory contains examples demonstrating how to use the QwenAI thinking mode functionality.

## 🚀 Quick Start

### 1. Thinking Mode Demo
Run the comprehensive demo showing all thinking mode features:

```bash
cd examples
go run thinking_mode_demo.go
```

This demo showcases:
- Basic thinking mode (enabled by default)
- Disabling thinking with `/no_think`
- Explicitly enabling with `/think`
- Complete thinking responses
- Dynamic thinking mode control
- Tool calling support

### 2. Server Integration
Run the HTTP server with thinking mode support:

```bash
cd examples
go run server_integration.go
```

Then open your browser to `http://localhost:8080` to see the interactive demo.

## 📁 Files

- **`thinking_mode_demo.go`** - Command-line demo of all thinking mode features
- **`server_integration.go`** - HTTP server with thinking mode endpoints
- **`demo.html`** - Interactive web interface for testing

## 🔧 API Endpoints

When running the server, you can use these endpoints:

- **`POST /chat`** - Stream chat with thinking mode
- **`GET /status`** - Server and model status
- **`POST /thinking`** - Complete thinking response (non-streaming)

## 💡 Usage Examples

### Basic Thinking Mode
```go
client := ai.NewClient(apiKey, baseURL, "qwen-plus-2025-04-28")

client.ChatStreamWithThinking("Explain quantum computing", func(stage, content string, isComplete bool) {
    switch stage {
    case "thinking":
        fmt.Printf("🤔 Thinking: %s", content)
    case "streaming":
        fmt.Printf("%s", content)
    }
})
```

### Prompt Control
```
What is the capital of France?/think    // Force thinking mode
What is 2+2?/no_think                  // Disable thinking mode
```

### Global Control
```go
client.SetThinkingMode(false)  // Disable globally
client.SetThinkingMode(true)   // Re-enable globally
```

## 🎯 Features Demonstrated

- **Real-time streaming** of thinking process
- **Dynamic mode switching** with prompt controls
- **Tool calling support** for enhanced functionality
- **Usage statistics** and token counting
- **Error handling** and graceful fallbacks
- **Web interface** for easy testing

## 🔍 Testing

1. **Run the demo**: `go run thinking_mode_demo.go`
2. **Start the server**: `go run server_integration.go`
3. **Open browser**: Navigate to `http://localhost:8080`
4. **Test different modes**: Try various prompts and modes
5. **Check status**: Use the status button to verify server state

## 🚨 Requirements

- Go 1.21 or later
- Valid DashScope API key
- Qwen model that supports thinking mode
- Environment variables configured (see main README)

## 📚 Next Steps

After exploring these examples:
1. Integrate thinking mode into your own applications
2. Customize the callback stages for your needs
3. Add additional tool calling capabilities
4. Implement your own UI/UX around the thinking process

Happy coding! 🎉
