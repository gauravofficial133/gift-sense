package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
)

type InteractionHandler struct {
	service *usecase.InteractionService
}

func NewInteractionHandler(service *usecase.InteractionService) *InteractionHandler {
	return &InteractionHandler{service: service}
}

type logInteractionRequest struct {
	SessionID  string            `json:"session_id" binding:"required"`
	CardIndex  int               `json:"card_index"`
	EventType  string            `json:"event_type" binding:"required"`
	DurationMs int64             `json:"duration_ms"`
	Changes    map[string]string `json:"changes"`
}

func (h *InteractionHandler) LogInteraction(c *gin.Context) {
	var req logInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	interaction := domain.CardInteraction{
		SessionID:  req.SessionID,
		CardIndex:  req.CardIndex,
		EventType:  req.EventType,
		DurationMs: req.DurationMs,
		Changes:    req.Changes,
	}

	if err := h.service.LogInteraction(c.Request.Context(), interaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged"})
}
