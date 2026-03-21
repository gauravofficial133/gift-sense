package ratelimiter

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/giftsense/backend/internal/database"
)

func TestDBRateLimiter_ShouldAllow_WhenUnderLimit(t *testing.T) {
	db := setupTestDB(t)
	limiter := NewDBRateLimiter(db, 5)

	allowed, err := limiter.Allow(context.Background(), "127.0.0.1")

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestDBRateLimiter_ShouldDeny_WhenOverLimit(t *testing.T) {
	db := setupTestDB(t)
	limiter := NewDBRateLimiter(db, 2)

	for range 2 {
		allowed, err := limiter.Allow(context.Background(), "127.0.0.1")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	allowed, err := limiter.Allow(context.Background(), "127.0.0.1")
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestDBRateLimiter_ShouldTrackSeparateKeys(t *testing.T) {
	db := setupTestDB(t)
	limiter := NewDBRateLimiter(db, 1)

	allowed, err := limiter.Allow(context.Background(), "127.0.0.1")
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = limiter.Allow(context.Background(), "192.168.1.1")
	require.NoError(t, err)
	assert.True(t, allowed)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping rate limiter integration test")
	}
	db, err := database.Connect(dbURL)
	require.NoError(t, err)

	err = db.AutoMigrate(&rateLimitRow{})
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Exec("DELETE FROM rate_limits")
	})

	return db
}
