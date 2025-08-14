package ai

import (
	"testing"
)

func TestProcessThinkingPrompt(t *testing.T) {
	client := &QwenThinkingClient{
		params: defaultParams,
	}

	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedThink  bool
	}{
		{
			name:           "No suffix - use default",
			input:          "Hello world",
			expectedOutput: "Hello world",
			expectedThink:  true, // default is true
		},
		{
			name:           "Explicit think",
			input:          "Hello world/think",
			expectedOutput: "Hello world",
			expectedThink:  true,
		},
		{
			name:           "Explicit no_think",
			input:          "Hello world/no_think",
			expectedOutput: "Hello world",
			expectedThink:  false,
		},
		{
			name:           "Think with spaces",
			input:          "Hello world /think",
			expectedOutput: "Hello world",
			expectedThink:  true,
		},
		{
			name:           "No_think with spaces",
			input:          "Hello world /no_think",
			expectedOutput: "Hello world",
			expectedThink:  false,
		},
		{
			name:           "Empty string",
			input:          "",
			expectedOutput: "",
			expectedThink:  true,
		},
		{
			name:           "Only think suffix",
			input:          "/think",
			expectedOutput: "",
			expectedThink:  true,
		},
		{
			name:           "Only no_think suffix",
			input:          "/no_think",
			expectedOutput: "",
			expectedThink:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, thinking := client.processThinkingPrompt(tt.input)
			if output != tt.expectedOutput {
				t.Errorf("processThinkingPrompt(%q) output = %q, want %q", tt.input, output, tt.expectedOutput)
			}
			if thinking != tt.expectedThink {
				t.Errorf("processThinkingPrompt(%q) thinking = %t, want %t", tt.input, thinking, tt.expectedThink)
			}
		})
	}
}

func TestNewQwenThinkingClient(t *testing.T) {
	client := NewQwenThinkingClient("test-key", "https://test.com", "qwen-test")

	if client.apiKey != "test-key" {
		t.Errorf("Expected apiKey 'test-key', got '%s'", client.apiKey)
	}

	if client.baseURL != "https://test.com" {
		t.Errorf("Expected baseURL 'https://test.com', got '%s'", client.baseURL)
	}

	if client.model != "qwen-test" {
		t.Errorf("Expected model 'qwen-test', got '%s'", client.model)
	}

	if !client.params.EnableThinking {
		t.Error("Expected EnableThinking to be true by default")
	}
}

func TestSetThinkingMode(t *testing.T) {
	client := NewQwenThinkingClient("test-key", "https://test.com", "qwen-test")

	// Test default
	if !client.params.EnableThinking {
		t.Error("Expected EnableThinking to be true by default")
	}

	// Test disable
	client.SetThinkingMode(false)
	if client.params.EnableThinking {
		t.Error("Expected EnableThinking to be false after SetThinkingMode(false)")
	}

	// Test enable
	client.SetThinkingMode(true)
	if !client.params.EnableThinking {
		t.Error("Expected EnableThinking to be true after SetThinkingMode(true)")
	}
}

func TestIsQwenModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{"Qwen model", "qwen-plus-2025-04-28", true},
		{"Qwen turbo", "qwen-turbo", true},
		{"Qwen max", "qwen-max", true},
		{"Non-Qwen model", "gpt-4", false},
		{"Empty model", "", false},
		{"Case insensitive", "QWEN-TEST", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a proper client using NewClient to initialize qwenThinking
			client := NewClient("test-key", "https://test.com", tt.model)
			result := client.IsQwenModel()
			if result != tt.expected {
				t.Errorf("IsQwenModel() for '%s' = %t, want %t", tt.model, result, tt.expected)
			}
		})
	}
}

func TestThinkingResponse(t *testing.T) {
	response := &ThinkingResponse{
		ReasoningContent: "Let me think about this...",
		AnswerContent:    "The answer is 42",
		IsComplete:       true,
	}

	if response.ReasoningContent != "Let me think about this..." {
		t.Errorf("Expected ReasoningContent 'Let me think about this...', got '%s'", response.ReasoningContent)
	}

	if response.AnswerContent != "The answer is 42" {
		t.Errorf("Expected AnswerContent 'The answer is 42', got '%s'", response.AnswerContent)
	}

	if !response.IsComplete {
		t.Error("Expected IsComplete to be true")
	}
}
