package interactionstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/giftsense/backend/internal/domain"
)

type InteractionModel struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	SessionID  string    `gorm:"index"`
	CardIndex  int
	EventType  string    `gorm:"index"`
	DurationMs int64
	ChangesJSON string
	CreatedAt  time.Time `gorm:"index"`
}

func (InteractionModel) TableName() string { return "card_interactions" }

type SQLiteStore struct {
	db *gorm.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}
	if err := db.AutoMigrate(&InteractionModel{}); err != nil {
		return nil, fmt.Errorf("migrating sqlite: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) SaveInteraction(ctx context.Context, interaction domain.CardInteraction) error {
	changesJSON := "{}"
	if len(interaction.Changes) > 0 {
		data, _ := json.Marshal(interaction.Changes)
		changesJSON = string(data)
	}
	model := InteractionModel{
		SessionID:   interaction.SessionID,
		CardIndex:   interaction.CardIndex,
		EventType:   interaction.EventType,
		DurationMs:  interaction.DurationMs,
		ChangesJSON: changesJSON,
		CreatedAt:   interaction.Timestamp,
	}
	return s.db.WithContext(ctx).Create(&model).Error
}

func (s *SQLiteStore) ListInteractions(ctx context.Context, sessionID string) ([]domain.CardInteraction, error) {
	var models []InteractionModel
	err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("listing interactions: %w", err)
	}
	return toInteractions(models), nil
}

func (s *SQLiteStore) ListRecent(ctx context.Context, limit int) ([]domain.CardInteraction, error) {
	var models []InteractionModel
	err := s.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("listing recent interactions: %w", err)
	}
	return toInteractions(models), nil
}

func (s *SQLiteStore) AggregateStats(ctx context.Context) (domain.InteractionStats, error) {
	var stats domain.InteractionStats
	stats.TemplatePopularity = make(map[string]int)
	stats.PalettePopularity = make(map[string]int)
	stats.ModelPopularity = make(map[string]int)

	var views, downloads, edits int64
	s.db.WithContext(ctx).Model(&InteractionModel{}).Where("event_type = ?", "view").Count(&views)
	s.db.WithContext(ctx).Model(&InteractionModel{}).Where("event_type = ?", "download").Count(&downloads)
	s.db.WithContext(ctx).Model(&InteractionModel{}).Where("event_type = ?", "edit").Count(&edits)
	stats.TotalViews = int(views)
	stats.TotalDownloads = int(downloads)
	stats.TotalEdits = int(edits)

	var avgDur struct{ Avg float64 }
	s.db.WithContext(ctx).Model(&InteractionModel{}).Where("event_type = ? AND duration_ms > 0", "view").Select("COALESCE(AVG(duration_ms), 0) as avg").Scan(&avgDur)
	stats.AvgViewDurationMs = int64(avgDur.Avg)

	recent, _ := s.ListRecent(ctx, 20)
	stats.RecentInteractions = recent

	return stats, nil
}

func toInteractions(models []InteractionModel) []domain.CardInteraction {
	result := make([]domain.CardInteraction, len(models))
	for i, m := range models {
		changes := make(map[string]string)
		_ = json.Unmarshal([]byte(m.ChangesJSON), &changes)
		result[i] = domain.CardInteraction{
			SessionID:  m.SessionID,
			CardIndex:  m.CardIndex,
			EventType:  m.EventType,
			DurationMs: m.DurationMs,
			Changes:    changes,
			Timestamp:  m.CreatedAt,
		}
	}
	return result
}
