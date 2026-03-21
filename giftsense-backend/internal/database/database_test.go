package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/giftsense/backend/internal/database/migration"
)

func TestConnect_ShouldReturnGormDB_WhenValidURLProvided(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := Connect(dbURL)

	require.NoError(t, err)
	assert.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
}

func TestConnect_ShouldReturnError_WhenInvalidURLProvided(t *testing.T) {
	db, err := Connect("postgres://invalid:invalid@localhost:9999/nonexistent?connect_timeout=1")

	if err != nil {
		assert.Nil(t, db)
		return
	}

	sqlDB, dbErr := db.DB()
	require.NoError(t, dbErr)
	assert.Error(t, sqlDB.Ping())
}

func TestRunMigrations_ShouldCreateTables_WhenCalledOnEmptyDatabase(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := Connect(dbURL)
	require.NoError(t, err)

	err = migration.RunMigrations(db)
	assert.NoError(t, err)

	assert.True(t, db.Migrator().HasTable(&migration.FeedbackModel{}))
	assert.True(t, db.Migrator().HasTable(&migration.AnalyticsEventModel{}))
}

func TestRunMigrations_ShouldBeIdempotent_WhenCalledMultipleTimes(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := Connect(dbURL)
	require.NoError(t, err)

	err = migration.RunMigrations(db)
	require.NoError(t, err)

	err = migration.RunMigrations(db)
	assert.NoError(t, err)
}
