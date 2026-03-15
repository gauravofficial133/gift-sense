package openai_test

import (
	"testing"

	openaiAdapter "github.com/giftsense/backend/internal/adapter/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLLMClient_ShouldReturnError_WhenAPIKeyIsEmpty(t *testing.T) {
	_, err := openaiAdapter.NewLLMClient("", "gpt-4o", 1000)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
}

func TestNewLLMClient_ShouldReturnError_WhenModelIsEmpty(t *testing.T) {
	_, err := openaiAdapter.NewLLMClient("sk-test", "", 1000)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "model")
}

func TestNewLLMClient_ShouldReturnClient_WhenValidParamsProvided(t *testing.T) {
	c, err := openaiAdapter.NewLLMClient("sk-test", "gpt-4o", 1000)
	require.NoError(t, err)
	assert.NotNil(t, c)
}
