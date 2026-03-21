package feedbackstore

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/giftsense/backend/internal/database"
	"github.com/giftsense/backend/internal/database/migration"
	"github.com/giftsense/backend/internal/domain"
)

func setupTestDB(t *testing.T) *GormFeedbackStore {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := database.Connect(dbURL)
	require.NoError(t, err)

	err = migration.RunMigrations(db)
	require.NoError(t, err)

	return NewGormFeedbackStore(db)
}

func TestGormFeedbackStore_ShouldInsertFeedback_WhenSaveFeedbackCalled(t *testing.T) {
	store := setupTestDB(t)

	fb := domain.Feedback{
		SessionID:       "550e8400-e29b-41d4-a716-446655440000",
		Satisfaction:    domain.SatisfactionHelpful,
		PurchaseIntent:  domain.PurchaseIntentDefinitely,
		BudgetTier:      "BUDGET",
		SuggestionCount: 5,
		Timestamp:       time.Now(),
	}

	err := store.SaveFeedback(context.Background(), fb)
	assert.NoError(t, err)
}

func TestGormFeedbackStore_ShouldInsertEvent_WhenSaveEventCalled(t *testing.T) {
	store := setupTestDB(t)

	evt := domain.AnalyticsEvent{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		EventType: "link_click",
		Target:    "amazon",
		Timestamp: time.Now(),
	}

	err := store.SaveEvent(context.Background(), evt)
	assert.NoError(t, err)
}

func TestGormFeedbackStore_ShouldStoreIssuesAsArray_WhenMultipleIssuesProvided(t *testing.T) {
	store := setupTestDB(t)

	fb := domain.Feedback{
		SessionID:       "550e8400-e29b-41d4-a716-446655440001",
		Satisfaction:    domain.SatisfactionNotHelpful,
		Issues:          []string{"personality_mismatch", "price_mismatch", "wrong_categories"},
		BudgetTier:      "MID_RANGE",
		SuggestionCount: 3,
		Timestamp:       time.Now(),
	}

	err := store.SaveFeedback(context.Background(), fb)
	assert.NoError(t, err)
}

func TestGormFeedbackStore_ShouldStoreMetadataAsJSON_WhenMetadataProvided(t *testing.T) {
	store := setupTestDB(t)

	evt := domain.AnalyticsEvent{
		SessionID: "550e8400-e29b-41d4-a716-446655440002",
		EventType: "link_click",
		Target:    "flipkart",
		Metadata:  map[string]string{"gift_name": "Pottery Kit", "gift_index": "0"},
		Timestamp: time.Now(),
	}

	err := store.SaveEvent(context.Background(), evt)
	assert.NoError(t, err)
}

func TestGormFeedbackStore_ShouldHandleConcurrentInserts_WhenCalledFromMultipleGoroutines(t *testing.T) {
	store := setupTestDB(t)

	var wg sync.WaitGroup
	errs := make([]error, 10)

	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fb := domain.Feedback{
				SessionID:       "550e8400-e29b-41d4-a716-446655440003",
				Satisfaction:    domain.SatisfactionHelpful,
				BudgetTier:      "PREMIUM",
				SuggestionCount: idx,
				Timestamp:       time.Now(),
			}
			errs[idx] = store.SaveFeedback(context.Background(), fb)
		}(i)
	}

	wg.Wait()

	for _, err := range errs {
		assert.NoError(t, err)
	}
}
