package port

import (
	"context"

	"github.com/giftsense/backend/internal/domain"
)

type SongEmotionCache interface {
	Get(ctx context.Context, trackID string) (*domain.CachedSongEmotion, error)
	Save(ctx context.Context, cached domain.CachedSongEmotion) error
}
