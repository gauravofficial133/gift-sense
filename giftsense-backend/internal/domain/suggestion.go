package domain

type PersonalityInsight struct {
	Insight         string `json:"insight"`
	EvidenceSummary string `json:"evidence_summary"`
}

type ShoppingLinks struct {
	Amazon         string `json:"amazon"`
	Flipkart       string `json:"flipkart"`
	GoogleShopping string `json:"google_shopping"`
}

type GiftSuggestion struct {
	Name              string        `json:"name"`
	Reason            string        `json:"reason"`
	EstimatedPriceINR string        `json:"estimated_price_inr"`
	Category          string        `json:"category"`
	Links             ShoppingLinks `json:"links"`
}

type AnalysisResult struct {
	PersonalityInsights []PersonalityInsight `json:"personality_insights"`
	GiftSuggestions     []GiftSuggestion     `json:"gift_suggestions"`
	Card                *CardRender          `json:"card,omitempty"`
}
