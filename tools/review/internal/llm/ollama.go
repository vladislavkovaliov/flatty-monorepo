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
	log.Printf("  ollama: http client created (base=%s, model=%s, timeout=120s)", baseURL, model)
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		http: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) Send(prompt string) (string, error) {
	log.Printf("  prompt size: %d characters", len(prompt))

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
	log.Printf("  request payload: %d bytes", len(payload))

	log.Printf("  sending POST %s/api/chat ...", c.BaseURL)
	start := time.Now()

	resp, err := c.http.Post(
		c.BaseURL+"/api/chat",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Round(time.Millisecond)
	log.Printf("  response received in %v (status=%d)", elapsed, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("ollama decode failed: %w", err)
	}

	answerSize := len(chatResp.Message.Content)
	log.Printf("  response size: %d characters", answerSize)

	return chatResp.Message.Content, nil
}

func (c *Client) Review(diff, language string) (string, error) {
	log.Printf("  building prompt...")
	prompt := buildPrompt(diff, language)
	return c.Send(prompt)
}

func buildPrompt(diff, language string) string {
	prompt := fmt.Sprintf("You are a code reviewer for %s code.\nReview the following changes.\n\n", language)

	if diff != "" {
		prompt += "## Git diff\n```diff\n" + diff + "\n```\n\n"
	}

	prompt += fmt.Sprintf(`Analyze the changes and categorize issues:

1. **Security**:
   - SQL injection, XSS, data leaks
   - Secrets/tokens in code
   - Missing input validation

2. **Logic and bugs**:
   - Race conditions, deadlocks, goroutine leaks
   - Incorrect error handling
   - Context cancellation not respected

3. **Code style** (best practices):
   - Naming, Go idioms
   - Unused code, dead code
   - Missing error handling

Output format — markdown table:

| Severity | File:Line | Issue | Fix |
|----------|-----------|-------|-----|

Severity: critical / major / minor

If no issues: "✅ No issues found."`)

	return prompt
}
