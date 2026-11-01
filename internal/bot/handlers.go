package bot

import (
	"bytes"
	"context"
	"log"
	"strings"

	telegrambot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"generation_bot/internal/generation"
)

type Handler struct {
	client *generation.Client
	state  *StateManager
}

func NewHandler(client *generation.Client) *Handler {
	return &Handler{
		client: client,
		state:  NewStateManager(),
	}
}

func (h *Handler) HandleMessage(
	ctx context.Context,
	b *telegrambot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	message := update.Message
	chatID := message.Chat.ID
	userID := chatID
	text := strings.TrimSpace(message.Text)

	if text == "" {
		return
	}

	if strings.HasPrefix(text, "/") {
		h.handleCommand(ctx, b, chatID, userID, text)
		return
	}

	if h.state.Get(userID) == StateWaitingForPrompt {
		h.state.Set(userID, StateIdle)
		h.generateImage(ctx, b, chatID, text)
		return
	}

	sendMessage(ctx, b, chatID,
		"Use /generate to create an image.")
}

func (h *Handler) generateImage(
	ctx context.Context,
	b *telegrambot.Bot,
	chatID int64,
	prompt string,
) {
	sendMessage(ctx, b, chatID, "Generating your image... 🎨")

	imageData, err := h.client.Generate(ctx, prompt)
	if err != nil {
		log.Printf("image generation failed: %v", err)
		sendMessage(ctx, b, chatID,
			"Sorry, image generation failed. Please try again.")
		return
	}

	_, err = b.SendPhoto(ctx, &telegrambot.SendPhotoParams{
		ChatID: chatID,
		Photo: &models.InputFileUpload{
			Filename: "generated.png",
			Data:     bytes.NewReader(imageData),
		},
		Caption: "Here is your generated image! ✨",
	})

	if err != nil {
		log.Printf("failed to send image: %v", err)
		sendMessage(ctx, b, chatID,
			"Image was generated, but I couldn't send it.")
	}
}

func sendMessage(
	ctx context.Context,
	b *telegrambot.Bot,
	chatID int64,
	text string,
) {
	_, err := b.SendMessage(ctx, &telegrambot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})

	if err != nil {
		log.Printf("failed to send message: %v", err)
	}
}
