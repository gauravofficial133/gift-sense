package songcache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/giftsense/backend/internal/domain"
)

// SongEmotionModel is the GORM model for cached Spotify song emotion analyses.
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

// GormSongCache implements port.SongEmotionCache using GORM/PostgreSQL.
type GormSongCache struct {
	db *gorm.DB
}

func NewGormSongCache(db *gorm.DB) *GormSongCache {
	return &GormSongCache{db: db}
}

func (s *GormSongCache) Get(_ context.Context, trackID string) (*domain.CachedSongEmotion, error) {
	var model SongEmotionModel
	result := s.db.Where("spotify_track_id = ?", trackID).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("song cache get: %w", result.Error)
	}

	var emotions []domain.EmotionSignal
	if err := json.Unmarshal([]byte(model.Emotions), &emotions); err != nil {
		return nil, fmt.Errorf("song cache unmarshal emotions: %w", err)
	}

	return &domain.CachedSongEmotion{
		TrackID:       model.SpotifyTrackID,
		TrackName:     model.TrackName,
		Artist:        model.Artist,
		Emotions:      emotions,
		LyricsSnippet: model.LyricsSnippet,
		LanguageLabel: model.LanguageLabel,
		AudioFeatures: domain.SpotifyAudioFeatures{
			Valence:      model.Valence,
			Energy:       model.Energy,
			Danceability: model.Danceability,
			Tempo:        model.Tempo,
		},
	}, nil
}

func (s *GormSongCache) Save(_ context.Context, cached domain.CachedSongEmotion) error {
	emotionsJSON, err := json.Marshal(cached.Emotions)
	if err != nil {
		return fmt.Errorf("song cache marshal emotions: %w", err)
	}

	model := SongEmotionModel{
		SpotifyTrackID: cached.TrackID,
		TrackName:      cached.TrackName,
		Artist:         cached.Artist,
		Emotions:       string(emotionsJSON),
		LyricsSnippet:  cached.LyricsSnippet,
		LanguageLabel:  cached.LanguageLabel,
		Valence:        cached.AudioFeatures.Valence,
		Energy:         cached.AudioFeatures.Energy,
		Danceability:   cached.AudioFeatures.Danceability,
		Tempo:          cached.AudioFeatures.Tempo,
	}

	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "spotify_track_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"track_name", "artist", "emotions", "lyrics_snippet", "language_label", "valence", "energy", "danceability", "tempo"}),
	}).Create(&model)

	if result.Error != nil {
		return fmt.Errorf("song cache save: %w", result.Error)
	}
	return nil
}
