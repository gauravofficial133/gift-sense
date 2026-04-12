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

type SongEmotionModel struct {
	ID             uint    `gorm:"primaryKey"`
	SpotifyTrackID string  `gorm:"type:varchar(50);uniqueIndex;not null"`
	TrackName      string  `gorm:"type:varchar(255);not null"`
	Artist         string  `gorm:"type:varchar(255);not null"`
	Emotions       string  `gorm:"type:jsonb;not null"`
	LyricsSnippet  string  `gorm:"type:varchar(200)"`
	LanguageLabel  string  `gorm:"type:varchar(50)"`
	Valence        float64 `gorm:"type:decimal(4,3)"`
	Energy         float64 `gorm:"type:decimal(4,3)"`
	Danceability   float64 `gorm:"type:decimal(4,3)"`
	Tempo          float64 `gorm:"type:decimal(6,2)"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (SongEmotionModel) TableName() string { return "song_emotions" }

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&FeedbackModel{}, &AnalyticsEventModel{}, &RateLimitModel{}, &SongEmotionModel{})
}
