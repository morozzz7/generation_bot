package bot

import (
	"bytes"
	"context"
	"log"
	"strings"

	telegrambot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"

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

	if message.From == nil {
		return
	}

	userID := message.From.ID
	text := strings.TrimSpace(message.Text)

	if text == "" {
		return
	}

	if strings.HasPrefix(text, "/") {
		h.handleCommand(ctx, b, chatID, userID, text)
		return
	}

	taskID := uuid.NewString()

	modelID, ok := h.state.StartTask(userID, taskID)
	if !ok {
		switch h.state.Get(userID) {
		case StateWaitingForModel:
			sendMessage(ctx, b, chatID, "Choose a model using the buttons above.")
		case StateGenerating:
			sendMessage(ctx, b, chatID, "You already have an active generation.")
		default:
			sendMessage(ctx, b, chatID, "Use /generate to create an image.")
		}
		return
	}

	go h.generateImage(
		context.Background(),
		b,
		chatID,
		userID,
		taskID,
		modelID,
		text,
	)
}

func (h *Handler) HandleCallbackQuery(
	ctx context.Context,
	b *telegrambot.Bot,
	update *models.Update,
) {
	callback := update.CallbackQuery
	if callback == nil {
		return
	}

	_, err := b.AnswerCallbackQuery(
		ctx,
		&telegrambot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
		},
	)
	if err != nil {
		log.Printf("failed to answer callback query: %v", err)
	}

	chatID := callback.Message.Message.Chat.ID
	userID := callback.From.ID

	var modelID string

	switch callback.Data {
	case "model:dreamshaper":
		modelID = "dreamshaper"
	case "model:sdxl_turbo":
		modelID = "sdxl_turbo"
	default:
		sendMessage(ctx, b, chatID, "Unknown model.")
		return
	}

	if !h.state.SelectModel(userID, modelID) {
		sendMessage(ctx, b, chatID, "This model selection has expired. Use /generate again.")
		return
	}

	modelName := "DreamShaper 8"
	if modelID == "sdxl_turbo" {
		modelName = "SDXL Turbo"
	}

	sendMessage(
		ctx,
		b,
		chatID,
		"Selected: "+modelName+"\n\nNow send me your image prompt. 📝",
	)
}

func (h *Handler) generateImage(
	ctx context.Context,
	b *telegrambot.Bot,
	chatID int64,
	userID int64,
	taskID string,
	modelID string,
	prompt string,
) {
	defer h.state.ClearTask(userID, taskID)

	modelName := "DreamShaper 8"
	if modelID == "sdxl_turbo" {
		modelName = "SDXL Turbo"
	}

	sendMessage(ctx, b, chatID, "Generating with "+modelName+"... 🎨")

	imageData, err := h.client.Generate(ctx, taskID, modelID, prompt)

	if err != nil {
		log.Printf("generation failed (task %s): %v", taskID, err)

		if strings.Contains(err.Error(), "Generation cancelled") {
			sendMessage(ctx, b, chatID, "Generation cancelled. ✋")
			return
		}

		sendMessage(ctx, b, chatID, "Sorry, image generation failed.")
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
		sendMessage(ctx, b, chatID, "Image was generated, but I couldn't send it.")
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
