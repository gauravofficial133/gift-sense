package dto

type FeedbackRequest struct {
	SessionID       string   `json:"session_id" binding:"required,uuid"`
	Satisfaction    string   `json:"satisfaction" binding:"required,oneof=helpful not_helpful"`
	PurchaseIntent  string   `json:"purchase_intent,omitempty" binding:"omitempty,oneof=definitely maybe probably_not"`
	Issues          []string `json:"issues,omitempty"`
	FreeText        string   `json:"free_text,omitempty" binding:"max=500"`
	BudgetTier      string   `json:"budget_tier" binding:"required"`
	SuggestionCount int      `json:"suggestion_count" binding:"min=0"`
}

type EventRequest struct {
	SessionID string            `json:"session_id" binding:"required,uuid"`
	EventType string            `json:"event_type" binding:"required,oneof=link_click"`
	Target    string            `json:"target" binding:"required"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type FeedbackResponse struct {
	Message string `json:"message"`
}
