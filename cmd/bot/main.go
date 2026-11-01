package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	telegrambot "github.com/go-telegram/bot"
	"github.com/joho/godotenv"

	"generation_bot/internal/bot"
	"generation_bot/internal/generation"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found; using environment variables")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://127.0.0.1:8000"
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	generationClient := generation.NewClient(apiBaseURL)
	handler := bot.NewHandler(generationClient)

	b, err := telegrambot.New(
		token,
		telegrambot.WithDefaultHandler(handler.HandleMessage),
	)
	if err != nil {
		log.Fatalf("failed to initialize bot: %v", err)
	}

	log.Println("Bot started")
	b.Start(ctx)
}
