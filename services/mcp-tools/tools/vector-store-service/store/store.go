package store

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Document struct {
	ID        string
	Text      string
	Metadata  map[string]any
	Tags      []string
	Embedding map[string]float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Result struct {
	Document Document
	Score    float64
}

type MemoryStore struct {
	mu   sync.RWMutex
	docs map[string]Document
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		docs: make(map[string]Document),
	}
}

func (s *MemoryStore) Upsert(doc Document) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if existing, ok := s.docs[doc.ID]; ok {
		if doc.CreatedAt.IsZero() {
			doc.CreatedAt = existing.CreatedAt
		}
	} else if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	doc.UpdatedAt = now
	s.docs[doc.ID] = doc
}

func (s *MemoryStore) Query(queryEmbedding map[string]float64, topK int, filter []string) []Result {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if topK <= 0 {
		topK = 5
	}

	filterSet := make(map[string]struct{})
	if len(filter) > 0 {
		for _, id := range filter {
			filterSet[id] = struct{}{}
		}
	}

	results := make([]Result, 0, len(s.docs))
	for _, doc := range s.docs {
		if len(filterSet) > 0 {
			if _, ok := filterSet[doc.ID]; !ok {
				continue
			}
		}
		score := cosineSimilarity(queryEmbedding, doc.Embedding)
		if score <= 0 {
			continue
		}
		results = append(results, Result{
			Document: doc,
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Document.ID < results[j].Document.ID
		}
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

var tokenRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)

func BuildEmbedding(text string) map[string]float64 {
	text = strings.ToLower(text)
	tokens := tokenRegex.FindAllString(text, -1)
	if len(tokens) == 0 {
		return map[string]float64{}
	}

	freq := make(map[string]float64)
	for _, token := range tokens {
		if len(token) < 2 {
			continue
		}
		freq[token]++
	}

	var norm float64
	for _, count := range freq {
		norm += count * count
	}
	if norm == 0 {
		return freq
	}
	norm = math.Sqrt(norm)
	for k, count := range freq {
		freq[k] = count / norm
	}
	return freq
}

func cosineSimilarity(a, b map[string]float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	var dot float64
	for token, aval := range a {
		if bval, ok := b[token]; ok {
			dot += aval * bval
		}
	}
	if dot == 0 {
		return 0
	}

	var normA, normB float64
	for _, val := range a {
		normA += val * val
	}
	for _, val := range b {
		normB += val * val
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
