package ratelimiter

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rateLimitRow struct {
	IPAddress   string    `gorm:"column:ip_address;type:varchar(45);primaryKey"`
	WindowStart time.Time `gorm:"column:window_start;primaryKey"`
	Count       int       `gorm:"column:count;not null;default:1"`
}

func (rateLimitRow) TableName() string { return "rate_limits" }

type DBRateLimiter struct {
	db    *gorm.DB
	limit int
}

func NewDBRateLimiter(db *gorm.DB, limit int) *DBRateLimiter {
	return &DBRateLimiter{db: db, limit: limit}
}

func (r *DBRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	windowStart := time.Now().UTC().Truncate(time.Minute)

	row := rateLimitRow{
		IPAddress:   key,
		WindowStart: windowStart,
		Count:       1,
	}

	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ip_address"}, {Name: "window_start"}},
			DoUpdates: clause.Assignments(map[string]any{"count": gorm.Expr("rate_limits.count + 1")}),
		}).
		Create(&row)
	if result.Error != nil {
		return false, result.Error
	}

	var current rateLimitRow
	if err := r.db.WithContext(ctx).
		Where("ip_address = ? AND window_start = ?", key, windowStart).
		First(&current).Error; err != nil {
		return false, err
	}

	if current.Count > r.limit {
		return false, nil
	}

	go r.cleanup(windowStart)

	return true, nil
}

func (r *DBRateLimiter) cleanup(before time.Time) {
	cutoff := before.Add(-5 * time.Minute)
	r.db.Where("window_start < ?", cutoff).Delete(&rateLimitRow{})
}
