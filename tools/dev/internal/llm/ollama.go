package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const defaultModel = "deepseek-coder:6.7b"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Message Message `json:"message"`
}

type Client struct {
	BaseURL string
	Model   string
	http    *http.Client
}

func NewClient(baseURL, model string) *Client {
	if model == "" {
		model = defaultModel
	}
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		http: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *Client) Send(prompt string) (string, error) {
	body := ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Post(
		c.BaseURL+"/api/chat",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("ollama decode failed: %w", err)
	}

	return chatResp.Message.Content, nil
}
