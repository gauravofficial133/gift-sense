package domain

import "time"

type SatisfactionRating string

const (
	SatisfactionHelpful    SatisfactionRating = "helpful"
	SatisfactionNotHelpful SatisfactionRating = "not_helpful"
)

type PurchaseIntent string

const (
	PurchaseIntentDefinitely  PurchaseIntent = "definitely"
	PurchaseIntentMaybe       PurchaseIntent = "maybe"
	PurchaseIntentProbablyNot PurchaseIntent = "probably_not"
)

type Feedback struct {
	SessionID       string             `json:"session_id"`
	Satisfaction    SatisfactionRating `json:"satisfaction"`
	PurchaseIntent  PurchaseIntent     `json:"purchase_intent,omitempty"`
	Issues          []string           `json:"issues,omitempty"`
	FreeText        string             `json:"free_text,omitempty"`
	BudgetTier      string             `json:"budget_tier"`
	SuggestionCount int                `json:"suggestion_count"`
	Timestamp       time.Time          `json:"timestamp"`
}
