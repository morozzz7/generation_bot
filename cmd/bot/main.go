package main

import (
	"context"
	telegrambot "github.com/go-telegram/bot"
	"log"
	"os"
	"os/signal"
	"syscall"
	"generation_bot/internal/bot"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found; using environment variables")
	}
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	opts := []telegrambot.Option{
		telegrambot.WithDefaultHandler(bot.HandleMessage),
	}

	b, err := telegrambot.New(token, opts...)
	if err != nil {
		log.Fatalf("failed to initialize bot: %v", err)
	}

	log.Println("Bot started")
	b.Start(ctx)
}
