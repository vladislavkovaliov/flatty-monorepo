package rag

import "unicode/utf8"

type TextSplitter struct {
	ChunkSize    int
	ChunkOverlap int
}

func NewSplitter(chunkSize, chunkOverlap int) *TextSplitter {
	return &TextSplitter{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

func (ts *TextSplitter) SplitText(text string) []string {
	if utf8.RuneCountInString(text) <= ts.ChunkSize {
		return []string{text}
	}

	var chunks []string
	runes := []rune(text)
	start := 0

	for start < len(runes) {
		end := start + ts.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunks = append(chunks, string(runes[start:end]))

		nextStart := end - ts.ChunkOverlap
		if nextStart <= start {
			nextStart = end
		}
		start = nextStart

		if start >= len(runes) {
			break
		}
	}

	return chunks
}
