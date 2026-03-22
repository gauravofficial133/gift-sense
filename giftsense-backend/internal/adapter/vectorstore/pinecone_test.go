package vectorstore_test

import (
	"testing"

	"github.com/giftsense/backend/internal/adapter/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPineconeStore_ShouldReturnError_WhenAPIKeyIsEmpty(t *testing.T) {
	_, err := vectorstore.NewPineconeStore("", "upahaar", "us-east-1", 1536)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
}

func TestNewPineconeStore_ShouldReturnError_WhenIndexNameIsEmpty(t *testing.T) {
	_, err := vectorstore.NewPineconeStore("test-key", "", "us-east-1", 1536)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "index name")
}

func TestNewPineconeStore_ShouldReturnError_WhenEnvironmentIsEmpty(t *testing.T) {
	_, err := vectorstore.NewPineconeStore("test-key", "upahaar", "", 1536)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "environment")
}
