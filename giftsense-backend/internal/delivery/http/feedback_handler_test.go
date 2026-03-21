package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
)

type inMemoryFeedbackStore struct {
	feedbacks []domain.Feedback
	events    []domain.AnalyticsEvent
}

func (s *inMemoryFeedbackStore) SaveFeedback(_ context.Context, fb domain.Feedback) error {
	s.feedbacks = append(s.feedbacks, fb)
	return nil
}

func (s *inMemoryFeedbackStore) SaveEvent(_ context.Context, evt domain.AnalyticsEvent) error {
	s.events = append(s.events, evt)
	return nil
}

func setupFeedbackRouter() (*gin.Engine, *inMemoryFeedbackStore) {
	gin.SetMode(gin.TestMode)
	store := &inMemoryFeedbackStore{}
	svc := usecase.NewFeedbackService(store)
	handler := NewFeedbackHandler(svc)

	r := gin.New()
	v1 := r.Group("/api/v1")
	v1.POST("/feedback", handler.SubmitFeedback)
	v1.POST("/events", handler.TrackEvent)

	return r, store
}

func TestSubmitFeedback_ShouldReturn201_WhenValidFeedbackProvided(t *testing.T) {
	router, store := setupFeedbackRouter()

	body := map[string]any{
		"session_id":       "550e8400-e29b-41d4-a716-446655440000",
		"satisfaction":     "helpful",
		"purchase_intent":  "definitely",
		"budget_tier":      "BUDGET",
		"suggestion_count": 5,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Feedback received", resp["message"])
	assert.Len(t, store.feedbacks, 1)
}

func TestSubmitFeedback_ShouldReturn400_WhenSatisfactionMissing(t *testing.T) {
	router, _ := setupFeedbackRouter()

	body := map[string]any{
		"session_id":  "550e8400-e29b-41d4-a716-446655440000",
		"budget_tier": "BUDGET",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitFeedback_ShouldReturn400_WhenSessionIDInvalid(t *testing.T) {
	router, _ := setupFeedbackRouter()

	body := map[string]any{
		"session_id":   "not-a-uuid",
		"satisfaction": "helpful",
		"budget_tier":  "BUDGET",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitFeedback_ShouldReturn400_WhenFreeTextTooLong(t *testing.T) {
	router, _ := setupFeedbackRouter()

	body := map[string]any{
		"session_id":   "550e8400-e29b-41d4-a716-446655440000",
		"satisfaction": "helpful",
		"free_text":    strings.Repeat("a", 501),
		"budget_tier":  "BUDGET",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTrackEvent_ShouldReturn204_WhenValidEventProvided(t *testing.T) {
	router, store := setupFeedbackRouter()

	body := map[string]any{
		"session_id": "550e8400-e29b-41d4-a716-446655440000",
		"event_type": "link_click",
		"target":     "amazon",
		"metadata":   map[string]string{"gift_name": "Pottery Kit"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Len(t, store.events, 1)
}

func TestTrackEvent_ShouldReturn400_WhenEventTypeInvalid(t *testing.T) {
	router, _ := setupFeedbackRouter()

	body := map[string]any{
		"session_id": "550e8400-e29b-41d4-a716-446655440000",
		"event_type": "page_view",
		"target":     "homepage",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
