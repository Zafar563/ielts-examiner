package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken     string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	LLMProvider  string // "gemini" or "openai"
	LLMAPIKey    string
	LLMModel     string
	LLMBaseURL   string // optional custom base url
}

// Load reads config from environment variables and optional .env file
func Load() (*Config, error) {
	// Attempt to load .env, ignore if missing
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded, using environment variables")
	}

	cfg := &Config{
		BotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "ielts_bot"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		LLMProvider: getEnv("LLM_PROVIDER", "gemini"),
		LLMAPIKey:   getEnv("LLM_API_KEY", ""),
		LLMModel:    getEnv("LLM_MODEL", ""),
		LLMBaseURL:  getEnv("LLM_BASE_URL", ""),
	}

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if cfg.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY is required")
	}

	// Default models if not specified
	if cfg.LLMModel == "" {
		if cfg.LLMProvider == "openai" {
			cfg.LLMModel = "gpt-4o-mini"
		} else {
			cfg.LLMModel = "gemini-3.6-flash"
		}
	}

	return cfg, nil
}

// DatabaseDSN generates PostgreSQL connection string
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}
