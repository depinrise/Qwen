package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"Qwen/internal/ai"
	"Qwen/internal/config"
)

// ChatRequest represents a chat request from the client
type ChatRequest struct {
	Message string `json:"message"`
	Mode    string `json:"mode"` // "thinking", "regular", or "auto"
}

// ChatResponse represents the response sent back to the client
type ChatResponse struct {
	Stage    string `json:"stage"`
	Content  string `json:"content"`
	Complete bool   `json:"complete"`
	Error    string `json:"error,omitempty"`
}

// Server represents our HTTP server with AI client
type Server struct {
	aiClient *ai.Client
}

// NewServer creates a new server instance
func NewServer() *Server {
	cfg := config.Load()
	client := ai.NewClient(cfg.DashScopeAPIKey, cfg.DashScopeBaseURL, cfg.AIModel)

	return &Server{
		aiClient: client,
	}
}

// handleChat handles chat requests with thinking mode support
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set response headers for streaming
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Flush headers
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Determine which mode to use
	switch strings.ToLower(req.Mode) {
	case "thinking":
		// Use thinking mode
		s.handleThinkingChat(w, req.Message)
	case "regular":
		// Use regular chat
		s.handleRegularChat(w, req.Message)
	case "auto", "":
		// Auto-detect based on model capabilities
		if s.aiClient.IsQwenModel() {
			s.handleThinkingChat(w, req.Message)
		} else {
			s.handleRegularChat(w, req.Message)
		}
	default:
		http.Error(w, "Invalid mode", http.StatusBadRequest)
	}
}

// handleThinkingChat handles chat with thinking mode
func (s *Server) handleThinkingChat(w http.ResponseWriter, message string) {
	// Ensure thinking mode is enabled
	s.aiClient.SetThinkingMode(true)

	s.aiClient.ChatStreamWithThinking(message, func(stage string, content string, isComplete bool) {
		response := ChatResponse{
			Stage:    stage,
			Content:  content,
			Complete: isComplete,
		}

		// Send response as JSON
		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			return
		}

		// Write response with newline for streaming
		fmt.Fprintf(w, "%s\n", string(jsonData))

		// Flush to ensure streaming works
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	})
}

// handleRegularChat handles regular chat without thinking mode
func (s *Server) handleRegularChat(w http.ResponseWriter, message string) {
	// Disable thinking mode
	s.aiClient.SetThinkingMode(false)

	s.aiClient.ChatStream(message, func(chunk string, isComplete bool) {
		response := ChatResponse{
			Stage:    "streaming",
			Content:  chunk,
			Complete: isComplete,
		}

		// Send response as JSON
		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			return
		}

		// Write response with newline for streaming
		fmt.Fprintf(w, "%s\n", string(jsonData))

		// Flush to ensure streaming works
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	})
}

// handleStatus returns server status and model information
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"status":        "running",
		"model":         s.aiClient.Model,
		"thinking_mode": s.aiClient.IsQwenModel(),
		"base_url":      s.aiClient.BaseURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleCompleteThinking handles non-streaming thinking responses
func (s *Server) handleCompleteThinking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create messages for the conversation
	messages := []ai.Message{
		{Role: "user", Content: req.Message},
	}

	// Get complete thinking response
	response, err := s.aiClient.ChatWithThinking(context.Background(), messages)
	if err != nil {
		http.Error(w, fmt.Sprintf("AI error: %v", err), http.StatusInternalServerError)
		return
	}

	// Return complete response
	result := map[string]interface{}{
		"reasoning_content": response.ReasoningContent,
		"answer_content":    response.AnswerContent,
		"is_complete":       response.IsComplete,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func runServer() {
	server := NewServer()

	// Set up routes
	http.HandleFunc("/chat", server.handleChat)
	http.HandleFunc("/status", server.handleStatus)
	http.HandleFunc("/thinking", server.handleCompleteThinking)

	// Serve static files for demo
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "examples/demo.html")
		} else {
			http.NotFound(w, r)
		}
	})

	port := ":8080"
	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Printf("📱 Chat endpoint: http://localhost%s/chat\n", port)
	fmt.Printf("📊 Status endpoint: http://localhost%s/status\n", port)
	fmt.Printf("🧠 Thinking endpoint: http://localhost%s/thinking\n", port)
	fmt.Printf("🌐 Demo page: http://localhost%s\n", port)

	log.Fatal(http.ListenAndServe(port, nil))
}

// To run this server, uncomment the following lines:
// func main() {
// 	runServer()
// }
