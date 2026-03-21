package port

import (
	"context"

	"github.com/giftsense/backend/internal/domain"
)

type FeedbackStore interface {
	SaveFeedback(ctx context.Context, feedback domain.Feedback) error
	SaveEvent(ctx context.Context, event domain.AnalyticsEvent) error
}
