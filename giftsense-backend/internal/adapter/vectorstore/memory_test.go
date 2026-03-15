package vectorstore_test

import (
	"context"
	"testing"

	"github.com/giftsense/backend/internal/adapter/vectorstore"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestChunks(sessionID string, n int) []domain.Chunk {
	chunks := make([]domain.Chunk, n)
	for i := 0; i < n; i++ {
		chunks[i] = domain.Chunk{
			ID:             sessionID + "_chunk_" + string(rune('0'+i)),
			SessionID:      sessionID,
			AnonymizedText: "test text " + string(rune('a'+i)),
			Metadata: domain.ChunkMetadata{
				HasPreference: i%2 == 0,
				HasWish:       i%3 == 0,
				Topics:        []string{"cooking"},
			},
		}
	}
	return chunks
}

func makeTestVectors(n, dims int) [][]float32 {
	vecs := make([][]float32, n)
	for i := 0; i < n; i++ {
		v := make([]float32, dims)
		v[i%dims] = 1.0
		vecs[i] = v
	}
	return vecs
}

func TestMemoryStore_ShouldReturnTopK_WhenChunksAreUpserted(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	ctx := context.Background()

	chunks := makeTestChunks("sess1", 5)
	vectors := makeTestVectors(5, 5)

	err := store.Upsert(ctx, "sess1", chunks, vectors)
	require.NoError(t, err)

	queryVec := []float32{1, 0, 0, 0, 0}
	results, err := store.Query(ctx, "sess1", queryVec, 2, port.MetadataFilter{})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(results), 2)
}

func TestMemoryStore_ShouldIsolateSessionsByID(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	ctx := context.Background()

	chunks1 := makeTestChunks("session-A", 3)
	chunks2 := makeTestChunks("session-B", 3)
	vecs := makeTestVectors(3, 3)

	require.NoError(t, store.Upsert(ctx, "session-A", chunks1, vecs))
	require.NoError(t, store.Upsert(ctx, "session-B", chunks2, vecs))

	queryVec := []float32{1, 0, 0}
	results, err := store.Query(ctx, "session-A", queryVec, 10, port.MetadataFilter{})
	require.NoError(t, err)
	for _, r := range results {
		assert.Equal(t, "session-A", r.SessionID)
	}
}

func TestMemoryStore_ShouldReturnEmpty_AfterDeleteSession(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	ctx := context.Background()

	chunks := makeTestChunks("sess-del", 3)
	vecs := makeTestVectors(3, 3)
	require.NoError(t, store.Upsert(ctx, "sess-del", chunks, vecs))
	require.NoError(t, store.DeleteSession(ctx, "sess-del"))

	queryVec := []float32{1, 0, 0}
	results, err := store.Query(ctx, "sess-del", queryVec, 10, port.MetadataFilter{})
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestMemoryStore_ShouldFilterByHasPreference_WhenMetadataFilterApplied(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	ctx := context.Background()

	chunks := makeTestChunks("sess-filter", 4)
	vecs := makeTestVectors(4, 4)
	require.NoError(t, store.Upsert(ctx, "sess-filter", chunks, vecs))

	trueVal := true
	filter := port.MetadataFilter{HasPreference: &trueVal}
	queryVec := []float32{1, 0, 0, 0}
	results, err := store.Query(ctx, "sess-filter", queryVec, 10, filter)
	require.NoError(t, err)
	for _, r := range results {
		assert.True(t, r.Metadata.HasPreference)
	}
}
