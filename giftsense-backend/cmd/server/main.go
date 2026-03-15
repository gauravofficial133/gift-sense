package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/config"
	"github.com/giftsense/backend/internal/adapter/linkgen"
	openaiAdapter "github.com/giftsense/backend/internal/adapter/openai"
	"github.com/giftsense/backend/internal/adapter/vectorstore"
	handler "github.com/giftsense/backend/internal/delivery/http"
	"github.com/giftsense/backend/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	embedder, err := openaiAdapter.NewEmbedder(cfg.OpenAIAPIKey, cfg.EmbeddingModel, cfg.EmbeddingDimensions)
	if err != nil {
		log.Fatalf("embedder: %v", err)
	}

	llmClient, err := openaiAdapter.NewLLMClient(cfg.OpenAIAPIKey, cfg.ChatModel, cfg.MaxTokens)
	if err != nil {
		log.Fatalf("llm client: %v", err)
	}

	store, err := vectorstore.NewPineconeStore(cfg.PineconeAPIKey, cfg.PineconeIndexName, cfg.PineconeEnvironment, cfg.EmbeddingDimensions)
	if err != nil {
		log.Fatalf("vector store: %v", err)
	}

	analyzer := usecase.NewAnalyzer(embedder, llmClient, store, linkgen.GenerateLinks, usecase.AnalyzerConfig{
		MaxProcessedMessages: cfg.MaxProcessedMessages,
		ChunkWindowSize:      cfg.ChunkWindowSize,
		ChunkOverlapSize:     cfg.ChunkOverlapSize,
		TopK:                 cfg.TopK,
		NumRetrievalQueries:  cfg.NumRetrievalQueries,
	})

	analyzeHandler := handler.NewAnalyzeHandler(analyzer, cfg.MaxFileSizeBytes)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(handler.CORS(cfg.AllowedOrigins))
	router.Use(handler.RequestSizeLimiter(cfg.MaxFileSizeBytes))

	router.GET("/health", handler.Health)

	v1 := router.Group("/api/v1")
	v1.POST("/analyze", analyzeHandler.Analyze)

	addr := ":" + cfg.Port
	log.Printf("GiftSense backend listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
