package port

import (
	"context"
	"github.com/giftsense/backend/internal/domain"
)

type MetadataFilter struct {
	HasPreference *bool
	HasWish       *bool
	Topics        []string
}

type VectorStore interface {
	Upsert(ctx context.Context, sessionID string, chunks []domain.Chunk, vectors [][]float32) error
	Query(ctx context.Context, sessionID string, queryVector []float32, topK int, filter MetadataFilter) ([]domain.Chunk, error)
	DeleteSession(ctx context.Context, sessionID string) error
}
