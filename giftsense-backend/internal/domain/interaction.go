package domain

import "time"

type CardInteraction struct {
	SessionID  string            `json:"session_id"`
	CardIndex  int               `json:"card_index"`
	EventType  string            `json:"event_type"`
	DurationMs int64             `json:"duration_ms,omitempty"`
	Changes    map[string]string `json:"changes,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

type InteractionStats struct {
	TotalViews           int                `json:"total_views"`
	TotalDownloads       int                `json:"total_downloads"`
	TotalEdits           int                `json:"total_edits"`
	TemplatePopularity   map[string]int     `json:"template_popularity"`
	PalettePopularity    map[string]int     `json:"palette_popularity"`
	ModelPopularity      map[string]int     `json:"model_popularity"`
	AvgViewDurationMs    int64              `json:"avg_view_duration_ms"`
	RecentInteractions   []CardInteraction  `json:"recent_interactions,omitempty"`
}

type DashboardOverview struct {
	TotalSessions        int                `json:"total_sessions"`
	TotalCardsGenerated  int                `json:"total_cards_generated"`
	TotalDownloads       int                `json:"total_downloads"`
	AvgValidationScore   float64            `json:"avg_validation_score"`
	AvgScoringTotal      float64            `json:"avg_scoring_total"`
	TemplatePopularity   map[string]int     `json:"template_popularity"`
	PalettePopularity    map[string]int     `json:"palette_popularity"`
	FamilyUsage          map[string]int     `json:"family_usage"`
	RecentValidations    []ValidationResult `json:"recent_validations,omitempty"`
	ScoringDistribution  []float64          `json:"scoring_distribution,omitempty"`
}
