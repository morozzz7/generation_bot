package bot

import (
	"context"
	"log"
	"strings"
	telegrambot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func HandleMessage(
	ctx context.Context,
	b *telegrambot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	var response string

	switch text {
		case "/start":
			response = "Wassup! I`m generation bot 🎨\n\n" +
			"Send me description of image and I try to generate it."
		case "/help":
			response = "Available commands:\n" +
				"/start — start working\n" +
				"/help — help\n" +
				"/about - information about me"
		case "/about":
			response = "I`m diffusion model which can generate anything what you want"
		default:
			if text == "" {
				response = "Пока я умею обрабатывать только текстовые сообщения."
			} else {
				response = "Ты написал: " + text
		}
	}

	_, err := b.SendMessage(ctx, &telegrambot.SendMessageParams{
		ChatID: chatID,
		Text:   response,
	})
	if err != nil {
		log.Printf("failed to send message: %v", err)
	}
}