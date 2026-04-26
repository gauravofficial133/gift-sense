package port

import (
	"context"

	"github.com/giftsense/backend/internal/domain"
)

type InteractionStore interface {
	SaveInteraction(ctx context.Context, interaction domain.CardInteraction) error
	ListInteractions(ctx context.Context, sessionID string) ([]domain.CardInteraction, error)
	ListRecent(ctx context.Context, limit int) ([]domain.CardInteraction, error)
	AggregateStats(ctx context.Context) (domain.InteractionStats, error)
}
