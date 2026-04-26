package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type InteractionService struct {
	store port.InteractionStore
}

func NewInteractionService(store port.InteractionStore) *InteractionService {
	return &InteractionService{store: store}
}

func (s *InteractionService) LogInteraction(ctx context.Context, interaction domain.CardInteraction) error {
	if interaction.SessionID == "" {
		return fmt.Errorf("interaction: session_id required")
	}
	if interaction.EventType == "" {
		return fmt.Errorf("interaction: event_type required")
	}
	validTypes := map[string]bool{
		"view": true, "download": true, "edit": true,
		"palette_change": true, "text_change": true,
	}
	if !validTypes[interaction.EventType] {
		return fmt.Errorf("interaction: invalid event_type %q", interaction.EventType)
	}
	if interaction.Timestamp.IsZero() {
		interaction.Timestamp = time.Now()
	}
	return s.store.SaveInteraction(ctx, interaction)
}

func (s *InteractionService) GetStats(ctx context.Context) (domain.InteractionStats, error) {
	return s.store.AggregateStats(ctx)
}

func (s *InteractionService) GetRecent(ctx context.Context, limit int) ([]domain.CardInteraction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.store.ListRecent(ctx, limit)
}
