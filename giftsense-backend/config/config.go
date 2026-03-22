package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	OpenAIAPIKey         string
	ChatModel            string
	EmbeddingModel       string
	EmbeddingDimensions  int
	MaxTokens            int
	TopK                 int
	NumRetrievalQueries  int
	PineconeAPIKey       string
	PineconeIndexName    string
	PineconeEnvironment  string
	MaxFileSizeBytes     int64
	MaxProcessedMessages int
	ChunkWindowSize      int
	ChunkOverlapSize     int
	Port                 string
	AllowedOrigins       []string
	RateLimitPerMinute   int
	DatabaseURL          string
}

func (c *Config) HasDatabase() bool {
	return c.DatabaseURL != ""
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := loadSecrets(cfg); err != nil {
		return nil, err
	}

	loadOptionals(cfg)

	return cfg, nil
}

func loadSecrets(cfg *Config) error {
	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	if cfg.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	cfg.PineconeAPIKey = os.Getenv("PINECONE_API_KEY")
	if cfg.PineconeAPIKey == "" {
		return fmt.Errorf("PINECONE_API_KEY environment variable is required")
	}

	cfg.PineconeEnvironment = os.Getenv("PINECONE_ENVIRONMENT")
	if cfg.PineconeEnvironment == "" {
		return fmt.Errorf("PINECONE_ENVIRONMENT environment variable is required")
	}

	return nil
}

func loadOptionals(cfg *Config) {
	cfg.ChatModel = getEnvString("CHAT_MODEL", "gpt-4o")
	cfg.EmbeddingModel = getEnvString("EMBEDDING_MODEL", "text-embedding-3-small")
	cfg.EmbeddingDimensions = getEnvInt("EMBEDDING_DIMENSIONS", 1536)
	cfg.MaxTokens = getEnvInt("MAX_TOKENS", 1000)
	cfg.TopK = getEnvInt("TOP_K", 3)
	cfg.NumRetrievalQueries = getEnvInt("NUM_RETRIEVAL_QUERIES", 4)
	cfg.PineconeIndexName = getEnvString("PINECONE_INDEX_NAME", "upahaar")
	cfg.MaxFileSizeBytes = int64(getEnvInt("MAX_FILE_SIZE_BYTES", 2097152))
	cfg.MaxProcessedMessages = getEnvInt("MAX_PROCESSED_MESSAGES", 400)
	cfg.ChunkWindowSize = getEnvInt("CHUNK_WINDOW_SIZE", 8)
	cfg.ChunkOverlapSize = getEnvInt("CHUNK_OVERLAP_SIZE", 3)
	cfg.Port = getEnvString("PORT", "8080")
	cfg.AllowedOrigins = parseOrigins(getEnvString("ALLOWED_ORIGINS", "http://localhost:5173"))
	for _, o := range cfg.AllowedOrigins {
		if o == "*" {
			log.Fatal("ALLOWED_ORIGINS must not contain wildcard '*'")
		}
	}
	cfg.RateLimitPerMinute = getEnvInt("RATE_LIMIT_PER_MINUTE", 5)
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
}

func getEnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
