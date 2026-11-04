package generation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrGenerationCancelled = errors.New("generation cancelled")

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type GenerateRequest struct {
	TaskID         string `json:"task_id"`
	ModelID        string `json:"model_id"`
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}
}

func (c *Client) Generate(
	ctx context.Context,
	taskID string,
	modelID string,
	prompt string) ([]byte, error) {
	// 1. Кодируем запрос в JSON
	requestBody, err := json.Marshal(GenerateRequest{
		TaskID:  taskID,
		ModelID: modelID,
		Prompt:  prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/generate",
		bytes.NewReader(requestBody),
	)

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf(
			"API returned status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	if len(imageData) == 0 {
		return nil, fmt.Errorf("API returned empty image")
	}
	return imageData, nil
}

func (c *Client) Cancel(ctx context.Context, taskID string) error {
	endpoint := fmt.Sprintf(
		"%s/cancel/%s",
		strings.TrimRight(c.baseURL, "/"),
		taskID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		nil,
	)
	if err != nil {
		return fmt.Errorf("create cancellation request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send cancellation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return fmt.Errorf(
			"cancellation API returned %s: %s",
			resp.Status,
			strings.TrimSpace(string(body)),
		)
	}

	return nil
}
