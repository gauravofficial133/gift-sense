package config_test

import (
	"os"
	"testing"

	"github.com/giftsense/backend/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequiredEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv("OPENAI_API_KEY", "test-openai-key")
	t.Setenv("PINECONE_API_KEY", "test-pinecone-key")
	t.Setenv("PINECONE_ENVIRONMENT", "us-east-1")
}

func TestLoad_ShouldReturnConfig_WhenAllRequiredEnvVarsAreSet(t *testing.T) {
	setRequiredEnvVars(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "test-openai-key", cfg.OpenAIAPIKey)
	assert.Equal(t, "test-pinecone-key", cfg.PineconeAPIKey)
	assert.Equal(t, "us-east-1", cfg.PineconeEnvironment)
}

func TestLoad_ShouldApplyDefaults_WhenOptionalEnvVarsAreAbsent(t *testing.T) {
	setRequiredEnvVars(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o", cfg.ChatModel)
	assert.Equal(t, "text-embedding-3-small", cfg.EmbeddingModel)
	assert.Equal(t, 1536, cfg.EmbeddingDimensions)
	assert.Equal(t, 1000, cfg.MaxTokens)
	assert.Equal(t, 3, cfg.TopK)
	assert.Equal(t, 4, cfg.NumRetrievalQueries)
	assert.Equal(t, "giftsense", cfg.PineconeIndexName)
	assert.Equal(t, int64(2097152), cfg.MaxFileSizeBytes)
	assert.Equal(t, 400, cfg.MaxProcessedMessages)
	assert.Equal(t, 8, cfg.ChunkWindowSize)
	assert.Equal(t, 3, cfg.ChunkOverlapSize)
	assert.Equal(t, "8080", cfg.Port)
}

func TestLoad_ShouldReturnError_WhenOpenAIAPIKeyIsMissing(t *testing.T) {
	t.Setenv("PINECONE_API_KEY", "test-pinecone-key")
	t.Setenv("PINECONE_ENVIRONMENT", "us-east-1")
	os.Unsetenv("OPENAI_API_KEY")
	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY")
}

func TestLoad_ShouldReturnError_WhenPineconeAPIKeyIsMissing(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-openai-key")
	t.Setenv("PINECONE_ENVIRONMENT", "us-east-1")
	os.Unsetenv("PINECONE_API_KEY")
	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PINECONE_API_KEY")
}

func TestLoad_ShouldReturnError_WhenPineconeEnvironmentIsMissing(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-openai-key")
	t.Setenv("PINECONE_API_KEY", "test-pinecone-key")
	os.Unsetenv("PINECONE_ENVIRONMENT")
	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PINECONE_ENVIRONMENT")
}

func TestLoad_ShouldParseAllowedOrigins_WhenCommaSeparated(t *testing.T) {
	setRequiredEnvVars(t)
	t.Setenv("ALLOWED_ORIGINS", "https://example.com,https://other.com")
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"https://example.com", "https://other.com"}, cfg.AllowedOrigins)
}
