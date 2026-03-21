package usecase

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/giftsense/backend/internal/domain"
)

type mockFeedbackStore struct {
	savedFeedback *domain.Feedback
	savedEvent    *domain.AnalyticsEvent
	err           error
}

func (m *mockFeedbackStore) SaveFeedback(_ context.Context, fb domain.Feedback) error {
	m.savedFeedback = &fb
	return m.err
}

func (m *mockFeedbackStore) SaveEvent(_ context.Context, evt domain.AnalyticsEvent) error {
	m.savedEvent = &evt
	return m.err
}

func TestSubmitFeedback_ShouldSaveFeedback_WhenSatisfactionIsHelpful(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	fb := domain.Feedback{
		SessionID:    "test-session",
		Satisfaction: domain.SatisfactionHelpful,
		BudgetTier:   "BUDGET",
	}

	err := svc.SubmitFeedback(context.Background(), fb)

	require.NoError(t, err)
	assert.NotNil(t, store.savedFeedback)
	assert.Equal(t, domain.SatisfactionHelpful, store.savedFeedback.Satisfaction)
}

func TestSubmitFeedback_ShouldSaveFeedback_WhenSatisfactionIsNotHelpful(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	fb := domain.Feedback{
		SessionID:    "test-session",
		Satisfaction: domain.SatisfactionNotHelpful,
		BudgetTier:   "BUDGET",
	}

	err := svc.SubmitFeedback(context.Background(), fb)

	require.NoError(t, err)
	assert.NotNil(t, store.savedFeedback)
	assert.Equal(t, domain.SatisfactionNotHelpful, store.savedFeedback.Satisfaction)
}

func TestSubmitFeedback_ShouldReturnError_WhenSatisfactionIsInvalid(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	fb := domain.Feedback{
		SessionID:    "test-session",
		Satisfaction: "invalid_rating",
		BudgetTier:   "BUDGET",
	}

	err := svc.SubmitFeedback(context.Background(), fb)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid satisfaction rating")
	assert.Nil(t, store.savedFeedback)
}

func TestSubmitFeedback_ShouldReturnError_WhenFreeTextExceedsMaxLength(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	fb := domain.Feedback{
		SessionID:    "test-session",
		Satisfaction: domain.SatisfactionHelpful,
		FreeText:     strings.Repeat("a", 501),
		BudgetTier:   "BUDGET",
	}

	err := svc.SubmitFeedback(context.Background(), fb)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "free text exceeds maximum length")
	assert.Nil(t, store.savedFeedback)
}

func TestSubmitFeedback_ShouldSetTimestamp_WhenCalled(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	fb := domain.Feedback{
		SessionID:    "test-session",
		Satisfaction: domain.SatisfactionHelpful,
		BudgetTier:   "BUDGET",
	}

	err := svc.SubmitFeedback(context.Background(), fb)

	require.NoError(t, err)
	assert.False(t, store.savedFeedback.Timestamp.IsZero())
}

func TestTrackEvent_ShouldSaveEvent_WhenEventTypeIsLinkClick(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	evt := domain.AnalyticsEvent{
		SessionID: "test-session",
		EventType: "link_click",
		Target:    "amazon",
		Metadata:  map[string]string{"gift_name": "Pottery Kit"},
	}

	err := svc.TrackEvent(context.Background(), evt)

	require.NoError(t, err)
	assert.NotNil(t, store.savedEvent)
	assert.Equal(t, "link_click", store.savedEvent.EventType)
	assert.False(t, store.savedEvent.Timestamp.IsZero())
}

func TestTrackEvent_ShouldReturnError_WhenEventTypeIsInvalid(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	evt := domain.AnalyticsEvent{
		SessionID: "test-session",
		EventType: "page_view",
		Target:    "homepage",
	}

	err := svc.TrackEvent(context.Background(), evt)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event type")
	assert.Nil(t, store.savedEvent)
}

func TestTrackEvent_ShouldReturnError_WhenMetadataExceedsMaxKeys(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	metadata := make(map[string]string)
	for i := range 11 {
		metadata[fmt.Sprintf("key_%d", i)] = "value"
	}

	evt := domain.AnalyticsEvent{
		SessionID: "test-session",
		EventType: "link_click",
		Target:    "amazon",
		Metadata:  metadata,
	}

	err := svc.TrackEvent(context.Background(), evt)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata exceeds maximum of 10 keys")
	assert.Nil(t, store.savedEvent)
}

func TestTrackEvent_ShouldReturnError_WhenMetadataKeyTooLong(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	evt := domain.AnalyticsEvent{
		SessionID: "test-session",
		EventType: "link_click",
		Target:    "amazon",
		Metadata:  map[string]string{strings.Repeat("k", 51): "value"},
	}

	err := svc.TrackEvent(context.Background(), evt)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata key exceeds maximum length of 50 characters")
	assert.Nil(t, store.savedEvent)
}

func TestTrackEvent_ShouldReturnError_WhenMetadataValueTooLong(t *testing.T) {
	store := &mockFeedbackStore{}
	svc := NewFeedbackService(store)

	evt := domain.AnalyticsEvent{
		SessionID: "test-session",
		EventType: "link_click",
		Target:    "amazon",
		Metadata:  map[string]string{"gift_name": strings.Repeat("v", 257)},
	}

	err := svc.TrackEvent(context.Background(), evt)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata value exceeds maximum length of 256 characters")
	assert.Nil(t, store.savedEvent)
}
