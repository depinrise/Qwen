package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	DashScopeAPIKey  string
	DashScopeBaseURL string
	AIModel          string
	HTTPPort         string
	DatabaseDSN      string
}

func Load() *Config {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		DashScopeAPIKey:  getEnv("DASHSCOPE_API_KEY", ""),
		DashScopeBaseURL: getEnv("DASHSCOPE_BASE_URL", "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"),
		AIModel:          getEnv("AI_MODEL", "qwen-mt-turbo"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		DatabaseDSN:      getEnv("DATABASE_DSN", ""),
	}

	if config.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	if config.DashScopeAPIKey == "" {
		log.Fatal("DASHSCOPE_API_KEY is required")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
