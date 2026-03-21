package migration

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type FeedbackModel struct {
	ID              uint           `gorm:"primaryKey"`
	SessionID       string         `gorm:"type:uuid;not null;index:idx_feedback_session"`
	Satisfaction    string         `gorm:"type:varchar(20);not null"`
	PurchaseIntent  string         `gorm:"type:varchar(20)"`
	Issues          pq.StringArray `gorm:"type:text[]"`
	FreeText        string         `gorm:"type:varchar(500)"`
	BudgetTier      string         `gorm:"type:varchar(20);not null"`
	SuggestionCount int            `gorm:"not null;default:0"`
	CreatedAt       time.Time      `gorm:"autoCreateTime;index:idx_feedback_created_at"`
}

func (FeedbackModel) TableName() string { return "feedbacks" }

type AnalyticsEventModel struct {
	ID        uint      `gorm:"primaryKey"`
	SessionID string    `gorm:"type:uuid;not null;index:idx_events_session"`
	EventType string    `gorm:"type:varchar(20);not null"`
	Target    string    `gorm:"type:varchar(100);not null"`
	Metadata  string    `gorm:"type:jsonb"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_events_created_at"`
}

func (AnalyticsEventModel) TableName() string { return "analytics_events" }

type RateLimitModel struct {
	IPAddress   string    `gorm:"type:varchar(45);primaryKey"`
	WindowStart time.Time `gorm:"primaryKey"`
	Count       int       `gorm:"not null;default:1"`
}

func (RateLimitModel) TableName() string { return "rate_limits" }

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&FeedbackModel{}, &AnalyticsEventModel{}, &RateLimitModel{})
}
