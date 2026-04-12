package port

import (
	"context"

	"github.com/giftsense/backend/internal/domain"
)

type SpotifyClient interface {
	Search(ctx context.Context, query string, limit int) ([]domain.SpotifyTrack, error)
	GetAudioFeatures(ctx context.Context, trackID string) (domain.SpotifyAudioFeatures, error)
}
