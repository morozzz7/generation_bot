package bot

import (
	"context"
	"strings"

	telegrambot "github.com/go-telegram/bot"
)

func (h *Handler) handleCommand(
	ctx context.Context,
	b *telegrambot.Bot,
	chatID int64,
	userID int64,
	text string,
) {
	command := strings.Fields(text)[0]
	command = strings.Split(command, "@")[0]

	switch command {
	case "/start":
		sendMessage(ctx, b, chatID,
			"Welcome to Generation Bot! 🎨\n\n"+
				"Use /generate to create an image.")

	case "/help":
		sendMessage(ctx, b, chatID,
			"Available commands:\n"+
				"/start — start working\n"+
				"/generate — generate an image\n"+
				"/cancel — cancel current operation\n"+
				"/help — show help\n"+
				"/about — information about the bot")

	case "/generate":
		h.state.Set(userID, StateWaitingForPrompt)
		sendMessage(ctx, b, chatID,
			"Send me a prompt for image generation. 📝")

	default:
		sendMessage(ctx, b, chatID,
			"Unknown command. Use /help to see available commands.")
	}
}
