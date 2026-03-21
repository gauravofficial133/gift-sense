package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/delivery/dto"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
)

// AnalyzeHandler holds dependencies for the analyze endpoint.
type AnalyzeHandler struct {
	analyzer       *usecase.Analyzer
	maxFileSizeB   int64
}

// NewAnalyzeHandler constructs the handler.
func NewAnalyzeHandler(analyzer *usecase.Analyzer, maxFileSizeB int64) *AnalyzeHandler {
	return &AnalyzeHandler{analyzer: analyzer, maxFileSizeB: maxFileSizeB}
}

// Analyze handles POST /api/v1/analyze.
func (h *AnalyzeHandler) Analyze(c *gin.Context) {
	var req dto.AnalyzeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	fh, err := c.FormFile("conversation")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "missing_file",
			Message: "conversation file is required",
		})
		return
	}

	text, err := ValidateConversationFile(fh, h.maxFileSizeB)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	budgetTier := domain.BudgetTier(req.BudgetTier)
	budget, ok := domain.BudgetRanges[budgetTier]
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_budget_tier",
			Message: "unknown budget tier",
		})
		return
	}

	recipient := domain.RecipientDetails{
		Name:     req.Name,
		Relation: req.Relation,
		Gender:   req.Gender,
		Occasion: req.Occasion,
		Budget:   budget,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	result, err := h.analyzer.Analyze(ctx, req.SessionID, text, recipient)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AnalyzeResponse{
		Data:    result,
		Message: "Analysis complete",
	})
}

// Health handles GET /health.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AnalyzeHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, dto.ErrorResponse{Error: "file_too_large", Message: err.Error()})
	case errors.Is(err, domain.ErrInvalidFileType):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_file_type", Message: err.Error()})
	case errors.Is(err, domain.ErrConversationTooShort):
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "conversation_too_short", Message: err.Error()})
	case errors.Is(err, domain.ErrLLMResponseInvalid):
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "llm_error", Message: "Analysis failed, please try again"})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "An unexpected error occurred"})
	}
}
