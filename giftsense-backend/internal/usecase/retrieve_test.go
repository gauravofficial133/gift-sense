package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
	"github.com/giftsense/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Stubs ---

type stubEmbedder struct {
	err error
}

func (s *stubEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if s.err != nil {
		return nil, s.err
	}
	vecs := make([][]float32, len(texts))
	for i := range texts {
		vecs[i] = []float32{float32(i), 0, 0}
	}
	return vecs, nil
}

type stubVectorStore struct {
	chunks []domain.Chunk
	err    error
}

func (s *stubVectorStore) Upsert(_ context.Context, _ string, _ []domain.Chunk, _ [][]float32) error {
	return nil
}

func (s *stubVectorStore) Query(_ context.Context, _ string, _ []float32, topK int, _ port.MetadataFilter) ([]domain.Chunk, error) {
	if s.err != nil {
		return nil, s.err
	}
	if topK > len(s.chunks) {
		return s.chunks, nil
	}
	return s.chunks[:topK], nil
}

func (s *stubVectorStore) DeleteSession(_ context.Context, _ string) error {
	return nil
}

// --- Tests ---

func TestBuildRetrievalQueries_ShouldReturnCorrectCount_WhenNumQueriesProvided(t *testing.T) {
	recipient := domain.RecipientDetails{
		Name:     "Priya",
		Relation: "mom",
		Occasion: "birthday",
		Budget:   domain.BudgetRanges[domain.BudgetTierMidRange],
	}
	queries := usecase.BuildRetrievalQueries(recipient, 4)
	assert.Len(t, queries, 4)
}

func TestBuildRetrievalQueries_ShouldIncludeRecipientContext_InEveryQuery(t *testing.T) {
	recipient := domain.RecipientDetails{
		Relation: "best friend",
		Occasion: "farewell",
		Budget:   domain.BudgetRanges[domain.BudgetTierBudget],
	}
	queries := usecase.BuildRetrievalQueries(recipient, 4)
	for _, q := range queries {
		assert.NotEmpty(t, q)
		assert.Contains(t, q, "farewell")
	}
}

func TestRetrieveAndRerank_ShouldDeduplicateChunks_WhenSameChunkReturnedByMultipleQueries(t *testing.T) {
	chunks := []domain.Chunk{
		{ID: "sess_chunk_0", SessionID: "sess", AnonymizedText: "text A"},
		{ID: "sess_chunk_1", SessionID: "sess", AnonymizedText: "text B"},
	}
	chunksByID := map[string]domain.Chunk{
		"sess_chunk_0": chunks[0],
		"sess_chunk_1": chunks[1],
	}
	store := &stubVectorStore{chunks: chunks}
	embedder := &stubEmbedder{}
	queries := []string{"query1", "query2", "query3", "query4"}

	results, err := usecase.RetrieveAndRerank(context.Background(), "sess", queries, chunksByID, embedder, store, 2)
	require.NoError(t, err)

	ids := make(map[string]bool)
	for _, r := range results {
		assert.False(t, ids[r.ID], "duplicate chunk ID found: "+r.ID)
		ids[r.ID] = true
	}
}

func TestRetrieveAndRerank_ShouldPrioritizeHasPreferenceChunks_WhenReranking(t *testing.T) {
	preferenceChunk := domain.Chunk{
		ID:             "sess_pref",
		SessionID:      "sess",
		AnonymizedText: "I love hiking",
		Metadata:       domain.ChunkMetadata{HasPreference: true},
	}
	normalChunk := domain.Chunk{
		ID:             "sess_normal",
		SessionID:      "sess",
		AnonymizedText: "General chat",
		Metadata:       domain.ChunkMetadata{HasPreference: false},
	}

	chunksByID := map[string]domain.Chunk{
		"sess_pref":   preferenceChunk,
		"sess_normal": normalChunk,
	}
	store := &stubVectorStore{chunks: []domain.Chunk{normalChunk, preferenceChunk}}
	embedder := &stubEmbedder{}
	queries := []string{"q1"}

	results, err := usecase.RetrieveAndRerank(context.Background(), "sess", queries, chunksByID, embedder, store, 2)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.True(t, results[0].Metadata.HasPreference, "first result should have HasPreference=true")
}

func TestRetrieveAndRerank_ShouldReturnError_WhenEmbedderFails(t *testing.T) {
	store := &stubVectorStore{}
	embedder := &stubEmbedder{err: errors.New("embed failed")}
	_, err := usecase.RetrieveAndRerank(context.Background(), "sess", []string{"q1"}, map[string]domain.Chunk{}, embedder, store, 3)
	require.Error(t, err)
}
