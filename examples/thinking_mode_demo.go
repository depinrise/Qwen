package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"Qwen/internal/ai"
	"Qwen/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create AI client with thinking mode enabled by default
	client := ai.NewClient(cfg.DashScopeAPIKey, cfg.DashScopeBaseURL, cfg.AIModel)

	// Check if the model supports thinking mode
	if client.IsQwenModel() {
		fmt.Printf("✅ Using Qwen model '%s' with thinking mode enabled by default\n", cfg.AIModel)
	} else {
		fmt.Printf("⚠️  Model '%s' doesn't support thinking mode, will use regular chat\n", cfg.AIModel)
	}

	// Example 1: Basic thinking mode (enabled by default)
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("EXAMPLE 1: Basic thinking mode (enabled by default)")
	fmt.Println(strings.Repeat("=", 50))

	userMessage := "Explain how photosynthesis works in simple terms"
	fmt.Printf("User: %s\n", userMessage)

	client.ChatStreamWithThinking(userMessage, func(stage string, content string, isComplete bool) {
		switch stage {
		case "thinking":
			fmt.Printf("🤔 Thinking: %s", content)
		case "thinking_complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " THINKING COMPLETE " + strings.Repeat("=", 30))
		case "streaming":
			fmt.Printf("%s", content)
		case "complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " RESPONSE COMPLETE " + strings.Repeat("=", 30))
		case "error":
			fmt.Printf("❌ Error: %s\n", content)
		}
	})

	// Example 2: Disable thinking mode with /no_think
	fmt.Println("\n\n" + strings.Repeat("=", 50))
	fmt.Println("EXAMPLE 2: Disable thinking mode with /no_think")
	fmt.Println(strings.Repeat("=", 50))

	userMessage2 := "What is the capital of France?/no_think"
	fmt.Printf("User: %s\n", userMessage2)

	client.ChatStreamWithThinking(userMessage2, func(stage string, content string, isComplete bool) {
		switch stage {
		case "thinking":
			fmt.Printf("🤔 Thinking: %s", content)
		case "thinking_complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " THINKING COMPLETE " + strings.Repeat("=", 30))
		case "streaming":
			fmt.Printf("%s", content)
		case "complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " RESPONSE COMPLETE " + strings.Repeat("=", 30))
		case "error":
			fmt.Printf("❌ Error: %s\n", content)
		}
	})

	// Example 3: Explicitly enable thinking mode with /think
	fmt.Println("\n\n" + strings.Repeat("=", 50))
	fmt.Println("EXAMPLE 3: Explicitly enable thinking mode with /think")
	fmt.Println(strings.Repeat("=", 50))

	userMessage3 := "Solve this math problem step by step: 2x + 5 = 13/think"
	fmt.Printf("User: %s\n", userMessage3)

	client.ChatStreamWithThinking(userMessage3, func(stage string, content string, isComplete bool) {
		switch stage {
		case "thinking":
			fmt.Printf("🤔 Thinking: %s", content)
		case "thinking_complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " THINKING COMPLETE " + strings.Repeat("=", 30))
		case "streaming":
			fmt.Printf("%s", content)
		case "complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " RESPONSE COMPLETE " + strings.Repeat("=", 30))
		case "error":
			fmt.Printf("❌ Error: %s\n", content)
		}
	})

	// Example 4: Complete thinking response with both reasoning and answer
	fmt.Println("\n\n" + strings.Repeat("=", 50))
	fmt.Println("EXAMPLE 4: Complete thinking response (non-streaming)")
	fmt.Println(strings.Repeat("=", 50))

	userMessage4 := "Why is the sky blue?"
	fmt.Printf("User: %s\n", userMessage4)

	messages := []ai.Message{
		{Role: "user", Content: userMessage4},
	}

	thinkingResponse, err := client.ChatWithThinking(context.Background(), messages)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if thinkingResponse.ReasoningContent != "" {
		fmt.Println("\n🤔 THINKING PROCESS:")
		fmt.Println(strings.Repeat("=", 30))
		fmt.Println(thinkingResponse.ReasoningContent)
		fmt.Println(strings.Repeat("=", 30))
	}

	fmt.Println("\n💬 FINAL ANSWER:")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Println(thinkingResponse.AnswerContent)
	fmt.Println(strings.Repeat("=", 30))

	// Example 5: Dynamic thinking mode control
	fmt.Println("\n\n" + strings.Repeat("=", 50))
	fmt.Println("EXAMPLE 5: Dynamic thinking mode control")
	fmt.Println(strings.Repeat("=", 50))

	// Disable thinking mode globally
	client.SetThinkingMode(false)
	fmt.Println("🔴 Thinking mode disabled globally")

	userMessage5 := "What is the weather like today?"
	fmt.Printf("User: %s\n", userMessage5)

	client.ChatStreamWithThinking(userMessage5, func(stage string, content string, isComplete bool) {
		switch stage {
		case "thinking":
			fmt.Printf("🤔 Thinking: %s", content)
		case "thinking_complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " THINKING COMPLETE " + strings.Repeat("=", 30))
		case "streaming":
			fmt.Printf("%s", content)
		case "complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " RESPONSE COMPLETE " + strings.Repeat("=", 30))
		case "error":
			fmt.Printf("❌ Error: %s\n", content)
		}
	})

	// Re-enable thinking mode
	client.SetThinkingMode(true)
	fmt.Println("\n🟢 Thinking mode re-enabled globally")

	// Example 6: Tool calling support (if available)
	fmt.Println("\n\n" + strings.Repeat("=", 50))
	fmt.Println("EXAMPLE 6: Tool calling support")
	fmt.Println(strings.Repeat("=", 50))

	userMessage6 := "What time is it right now?/think"
	fmt.Printf("User: %s\n", userMessage6)

	client.ChatStreamWithThinking(userMessage6, func(stage string, content string, isComplete bool) {
		switch stage {
		case "thinking":
			fmt.Printf("🤔 Thinking: %s", content)
		case "thinking_complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " THINKING COMPLETE " + strings.Repeat("=", 30))
		case "streaming":
			fmt.Printf("%s", content)
		case "tool_call":
			fmt.Printf("🔧 Tool Call: %s\n", content)
		case "usage":
			fmt.Printf("📊 Usage: %s\n", content)
		case "complete":
			fmt.Println("\n" + strings.Repeat("=", 30) + " RESPONSE COMPLETE " + strings.Repeat("=", 30))
		case "error":
			fmt.Printf("❌ Error: %s\n", content)
		}
	})

	fmt.Println("\n🎉 Demo completed! Thinking mode is now enabled by default for Qwen models.")
	fmt.Println("Use /think to explicitly enable thinking mode")
	fmt.Println("Use /no_think to disable thinking mode for a specific query")
	fmt.Println("Use SetThinkingMode() to control thinking mode globally")
}
