package dto

import "github.com/giftsense/backend/internal/domain"

// AnalyzeResponse is the successful 200 body for POST /api/v1/analyze.
type AnalyzeResponse struct {
	Data    domain.AnalysisResult `json:"data"`
	Message string                `json:"message"`
}

// ErrorResponse is returned for all 4xx/5xx responses.
type ErrorResponse struct {
	Error   string         `json:"error"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}
