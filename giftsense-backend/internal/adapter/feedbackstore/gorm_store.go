package feedbackstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/giftsense/backend/internal/database/migration"
	"github.com/giftsense/backend/internal/domain"
)

type GormFeedbackStore struct {
	db *gorm.DB
}

func NewGormFeedbackStore(db *gorm.DB) *GormFeedbackStore {
	return &GormFeedbackStore{db: db}
}

func (s *GormFeedbackStore) SaveFeedback(ctx context.Context, fb domain.Feedback) error {
	model := migration.FeedbackModel{
		SessionID:       fb.SessionID,
		Satisfaction:    string(fb.Satisfaction),
		PurchaseIntent:  string(fb.PurchaseIntent),
		Issues:          pq.StringArray(fb.Issues),
		FreeText:        fb.FreeText,
		BudgetTier:      fb.BudgetTier,
		SuggestionCount: fb.SuggestionCount,
		CreatedAt:       fb.Timestamp,
	}

	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("saving feedback: %w", err)
	}

	return nil
}

func (s *GormFeedbackStore) SaveEvent(ctx context.Context, evt domain.AnalyticsEvent) error {
	metadataJSON := "{}"
	if len(evt.Metadata) > 0 {
		data, err := json.Marshal(evt.Metadata)
		if err != nil {
			return fmt.Errorf("marshaling event metadata: %w", err)
		}
		metadataJSON = string(data)
	}

	model := migration.AnalyticsEventModel{
		SessionID: evt.SessionID,
		EventType: evt.EventType,
		Target:    evt.Target,
		Metadata:  metadataJSON,
		CreatedAt: evt.Timestamp,
	}

	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("saving analytics event: %w", err)
	}

	return nil
}
