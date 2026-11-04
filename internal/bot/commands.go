package bot

import (
	"context"
	"log"
	"strings"

	telegrambot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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
				"/cancel — cancel current generation\n"+
				"/help — show help\n"+
				"/about — information about the bot")

	case "/generate":
		if !h.state.BeginModelSelection(userID) {
			sendMessage(ctx, b, chatID, "You already have an active generation.")
			return
		}

		keyboard := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "🎨 DreamShaper 8",
						CallbackData: "model:dreamshaper",
					},
				},
				{
					{
						Text:         "⚡ SDXL Turbo",
						CallbackData: "model:sdxl_turbo",
					},
				},
			},
		}

		_, err := b.SendMessage(ctx, &telegrambot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Choose a model for image generation:",
			ReplyMarkup: keyboard,
		})

		if err != nil {
			log.Printf("failed to send model keyboard: %v", err)
		}
	case "/about":
		sendMessage(ctx, b, chatID,
			"I`m a generation bot.\n"+
				"Now there are two available models:\n"+
				"SDXL Turbo and DreamShaper 8")
	case "/cancel":
		taskID, ok := h.state.GetTask(userID)

		if !ok {
			if h.state.Get(userID) == StateWaitingForPrompt {
				h.state.Set(userID, StateIdle)
				sendMessage(ctx, b, chatID, "Prompt cancelled.")
				return
			}

			sendMessage(ctx, b, chatID, "There is no active generation.")
			return
		}

		err := h.client.Cancel(ctx, taskID)
		if err != nil {
			log.Printf("failed to cancel task %s: %v", taskID, err)

			sendMessage(ctx, b, chatID,
				"Could not cancel generation. It may have already finished.")
			return
		}

		sendMessage(ctx, b, chatID, "Cancellation requested. ✋")

	default:
		sendMessage(ctx, b, chatID,
			"Unknown command. Use /help to see available commands.")
	}
}
