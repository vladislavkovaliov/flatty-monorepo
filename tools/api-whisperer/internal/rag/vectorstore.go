package rag

import (
	"math"
	"sort"
)

type VectorStore struct {
	Texts      []string
	Embeddings [][]float32
}

func NewVectorStore() *VectorStore {
	return &VectorStore{}
}

func (vs *VectorStore) Add(text string, embedding []float32) {
	vs.Texts = append(vs.Texts, text)
	vs.Embeddings = append(vs.Embeddings, embedding)
}

func (vs *VectorStore) Search(query []float32, k int) []string {
	type scoredIndex struct {
		index int
		score float32
	}

	scores := make([]scoredIndex, len(vs.Embeddings))
	for i, emb := range vs.Embeddings {
		scores[i] = scoredIndex{index: i, score: cosineSimilarity(query, emb)}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	if k > len(scores) {
		k = len(scores)
	}

	result := make([]string, k)
	for i := 0; i < k; i++ {
		result[i] = vs.Texts[scores[i].index]
	}
	return result
}

func cosineSimilarity(a, b []float32) float32 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	var dotProduct, normA, normB float64
	for i := range n {
		va := float64(a[i])
		vb := float64(b[i])
		dotProduct += va * vb
		normA += va * va
		normB += vb * vb
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}
