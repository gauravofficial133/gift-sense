package openai_test

import (
	"testing"

	openaiAdapter "github.com/giftsense/backend/internal/adapter/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmbedder_ShouldReturnError_WhenAPIKeyIsEmpty(t *testing.T) {
	_, err := openaiAdapter.NewEmbedder("", "text-embedding-3-small", 1536)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
}

func TestNewEmbedder_ShouldReturnError_WhenModelIsEmpty(t *testing.T) {
	_, err := openaiAdapter.NewEmbedder("sk-test", "", 1536)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "model")
}

func TestNewEmbedder_ShouldReturnEmbedder_WhenValidParamsProvided(t *testing.T) {
	e, err := openaiAdapter.NewEmbedder("sk-test", "text-embedding-3-small", 1536)
	require.NoError(t, err)
	assert.NotNil(t, e)
}
