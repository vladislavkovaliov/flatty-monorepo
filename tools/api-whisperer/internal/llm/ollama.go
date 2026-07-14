package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ollama/ollama/api"
)

type Client struct {
	client *api.Client
}

func New(baseURL string) *Client {
	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		u, _ = url.Parse("http://localhost:11434")
	}
	return &Client{
		client: api.NewClient(u, http.DefaultClient),
	}
}

func (c *Client) Embed(model, text string) ([]float32, error) {
	resp, err := c.client.Embed(context.Background(), &api.EmbedRequest{
		Model: model,
		Input: text,
	})
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("ollama embed: empty response")
	}
	return resp.Embeddings[0], nil
}

func (c *Client) Generate(model, prompt string) (string, error) {
	stream := false
	var result strings.Builder

	err := c.client.Generate(context.Background(), &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: &stream,
	}, func(r api.GenerateResponse) error {
		result.WriteString(r.Response)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("ollama generate: %w", err)
	}

	return result.String(), nil
}
