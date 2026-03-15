package dto

// AnalyzeRequest holds the parsed multipart form fields for POST /api/v1/analyze.
// The conversation file is handled separately by the handler.
type AnalyzeRequest struct {
	SessionID  string `form:"session_id"  binding:"required,uuid"`
	Name       string `form:"name"        binding:"required"`
	Relation   string `form:"relation"`
	Gender     string `form:"gender"`
	Occasion   string `form:"occasion"    binding:"required"`
	BudgetTier string `form:"budget_tier" binding:"required,oneof=BUDGET MID_RANGE PREMIUM LUXURY"`
}
