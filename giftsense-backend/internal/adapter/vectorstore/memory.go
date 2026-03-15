package vectorstore

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type storedEntry struct {
	chunk  domain.Chunk
	vector []float32
}

type MemoryStore struct {
	mu       sync.Mutex
	sessions map[string][]storedEntry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string][]storedEntry)}
}

func (m *MemoryStore) Upsert(_ context.Context, sessionID string, chunks []domain.Chunk, vectors [][]float32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entries := make([]storedEntry, len(chunks))
	for i, c := range chunks {
		entries[i] = storedEntry{chunk: c, vector: vectors[i]}
	}
	m.sessions[sessionID] = append(m.sessions[sessionID], entries...)
	return nil
}

func (m *MemoryStore) Query(_ context.Context, sessionID string, queryVector []float32, topK int, filter port.MetadataFilter) ([]domain.Chunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries := m.sessions[sessionID]
	type scored struct {
		chunk domain.Chunk
		score float64
	}

	var candidates []scored
	for _, e := range entries {
		if !matchesFilter(e.chunk, filter) {
			continue
		}
		candidates = append(candidates, scored{chunk: e.chunk, score: cosineSimilarity(queryVector, e.vector)})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if topK > len(candidates) {
		topK = len(candidates)
	}
	result := make([]domain.Chunk, topK)
	for i := 0; i < topK; i++ {
		result[i] = candidates[i].chunk
	}
	return result, nil
}

func (m *MemoryStore) DeleteSession(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func matchesFilter(chunk domain.Chunk, filter port.MetadataFilter) bool {
	if filter.HasPreference != nil && chunk.Metadata.HasPreference != *filter.HasPreference {
		return false
	}
	if filter.HasWish != nil && chunk.Metadata.HasWish != *filter.HasWish {
		return false
	}
	return true
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
