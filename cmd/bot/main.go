package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"ielts_bot/internal/bot"
	"ielts_bot/internal/config"
	"ielts_bot/internal/database"
	"ielts_bot/internal/llm"
)

func main() {
	log.Println("Starting IELTS Writing Examiner Bot...")

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// 2. Connect to PostgreSQL
	log.Println("Connecting to PostgreSQL database...")
	db, err := database.NewPostgresDB(cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established successfully.")

	// 3. Run database migrations
	log.Println("Applying database migrations...")
	if err := database.RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("Database migration error: %v", err)
	}

	// 4. Initialize repository & LLM client
	repo := database.NewRepository(db)
	llmClient := llm.NewClient(cfg)

	// 5. Initialize Telegram Bot
	telegramBot, err := bot.New(cfg, repo, llmClient)
	if err != nil {
		log.Fatalf("Telegram Bot initialization error: %v", err)
	}

	// 6. Graceful shutdown handler
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopChan
		log.Println("Shutting down bot gracefully...")
		os.Exit(0)
	}()

	// 7. Start bot
	log.Println("IELTS Writing Bot is now running!")
	if err := telegramBot.Start(); err != nil {
		log.Fatalf("Bot runtime error: %v", err)
	}
}
