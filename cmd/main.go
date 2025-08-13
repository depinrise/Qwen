package main

import (
	"Qwen/internal/ai"
	"Qwen/internal/bot"
	"Qwen/internal/config"
	"Qwen/internal/database"
	"Qwen/internal/memory"
	"Qwen/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize AI client
	aiClient := ai.NewClient(cfg.DashScopeAPIKey, cfg.DashScopeBaseURL, cfg.AIModel)

	// Initialize database connection (optional)
	var convService *database.ConversationService
	var memoryService *memory.MemoryService
	if cfg.DatabaseDSN != "" {
		db, err := database.NewConnection(cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Warning: Failed to connect to database: %v", err)
			log.Println("Bot will continue without conversation history and memory")
		} else {
			convService = database.NewConversationService(db)
			memoryService = memory.NewMemoryService(db.GetConnection(), aiClient)
			log.Println("✅ Database connection established")
			log.Println("🧠 Memory service initialized with LLM integration")
		}
	} else {
		log.Println("🔄 No database configured - running without conversation history and memory")
	}

	// Initialize bot handler
	botHandler, err := bot.NewHandler(cfg.TelegramBotToken, aiClient, convService, memoryService)
	if err != nil {
		log.Fatal("Failed to create bot handler:", err)
	}

	// Initialize HTTP server for WebSocket
	httpServer := server.NewServer(aiClient, cfg.HTTPPort)

	// Start bot in a goroutine
	go func() {
		log.Println("Starting Telegram bot...")
		if err := botHandler.Start(); err != nil {
			log.Fatal("Failed to start bot:", err)
		}
	}()

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s...", cfg.HTTPPort)
		if err := httpServer.Start(); err != nil {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	log.Printf("Bot is running on port %s. WebSocket available at /ws. Press Ctrl+C to stop.", cfg.HTTPPort)
	<-c

	log.Println("Shutting down bot...")
	botHandler.Stop()
	log.Println("Bot stopped successfully.")
}
