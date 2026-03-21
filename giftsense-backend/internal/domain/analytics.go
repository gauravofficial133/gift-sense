package domain

import "time"

type AnalyticsEvent struct {
	SessionID string            `json:"session_id"`
	EventType string            `json:"event_type"`
	Target    string            `json:"target"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}
