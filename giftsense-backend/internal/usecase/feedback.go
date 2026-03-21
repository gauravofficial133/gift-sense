package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type FeedbackService struct {
	store port.FeedbackStore
}

func NewFeedbackService(store port.FeedbackStore) *FeedbackService {
	return &FeedbackService{store: store}
}

func (s *FeedbackService) SubmitFeedback(ctx context.Context, fb domain.Feedback) error {
	if err := validateSatisfaction(fb.Satisfaction); err != nil {
		return err
	}

	if len(fb.FreeText) > 500 {
		return fmt.Errorf("free text exceeds maximum length of 500 characters")
	}

	fb.Timestamp = time.Now()

	return s.store.SaveFeedback(ctx, fb)
}

func (s *FeedbackService) TrackEvent(ctx context.Context, evt domain.AnalyticsEvent) error {
	if evt.EventType != "link_click" {
		return fmt.Errorf("invalid event type: %s", evt.EventType)
	}

	evt.Timestamp = time.Now()

	return s.store.SaveEvent(ctx, evt)
}

func validateSatisfaction(rating domain.SatisfactionRating) error {
	switch rating {
	case domain.SatisfactionHelpful, domain.SatisfactionNotHelpful:
		return nil
	default:
		return fmt.Errorf("invalid satisfaction rating: %s", rating)
	}
}
