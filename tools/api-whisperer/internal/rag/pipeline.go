package rag

import (
	"fmt"
	"strings"

	"github.com/vladislavkovaliov/api-whisperer/internal/llm"
)

type Pipeline struct {
	client     *llm.Client
	embedModel string
	chunks     []string
	store      *VectorStore
}

func NewPipeline(client *llm.Client, text, embedModel string, splitter *TextSplitter) *Pipeline {
	p := &Pipeline{
		client:     client,
		embedModel: embedModel,
		store:      NewVectorStore(),
	}

	p.chunks = splitter.SplitText(text)
	for _, chunk := range p.chunks {
		emb, err := client.Embed(embedModel, chunk)
		if err != nil {
			continue
		}
		p.store.Add(chunk, emb)
	}

	return p
}

func (p *Pipeline) Search(query string, topK int) (string, error) {
	queryEmb, err := p.client.Embed(p.embedModel, query)
	if err != nil {
		return "", fmt.Errorf("embed query: %w", err)
	}

	topChunks := p.store.Search(queryEmb, topK)
	var b strings.Builder
	for i, c := range topChunks {
		b.WriteString(fmt.Sprintf("Chunk %d:\n%s\n\n", i+1, c))
	}

	return b.String(), nil
}
