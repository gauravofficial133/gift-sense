package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/delivery/dto"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
)

type FeedbackHandler struct {
	service *usecase.FeedbackService
}

func NewFeedbackHandler(service *usecase.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{service: service}
}

func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	var req dto.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	fb := domain.Feedback{
		SessionID:       req.SessionID,
		Satisfaction:    domain.SatisfactionRating(req.Satisfaction),
		PurchaseIntent:  domain.PurchaseIntent(req.PurchaseIntent),
		Issues:          req.Issues,
		FreeText:        req.FreeText,
		BudgetTier:      req.BudgetTier,
		SuggestionCount: req.SuggestionCount,
	}

	if err := h.service.SubmitFeedback(c.Request.Context(), fb); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to save feedback",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.FeedbackResponse{
		Message: "Feedback received",
	})
}

func (h *FeedbackHandler) TrackEvent(c *gin.Context) {
	var req dto.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	evt := domain.AnalyticsEvent{
		SessionID: req.SessionID,
		EventType: req.EventType,
		Target:    req.Target,
		Metadata:  req.Metadata,
	}

	if err := h.service.TrackEvent(c.Request.Context(), evt); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "metadata") || strings.Contains(errMsg, "invalid event") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "validation_error",
				Message: errMsg,
			})
			return
		}
		log.Printf("analytics event error: %v", err)
	}

	c.Status(http.StatusNoContent)
}
