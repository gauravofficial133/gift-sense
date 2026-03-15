// Package handler is the Vercel serverless entry point for the GiftSense backend.
// Vercel requires Go handler files in api/ with an exported http.HandlerFunc.
// The Gin engine is wired once per cold start via sync.Once and reused across
// subsequent requests within the same Lambda instance.
package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/config"
	"github.com/giftsense/backend/internal/adapter/linkgen"
	openaiAdapter "github.com/giftsense/backend/internal/adapter/openai"
	"github.com/giftsense/backend/internal/adapter/vectorstore"
	deliveryhttp "github.com/giftsense/backend/internal/delivery/http"
	"github.com/giftsense/backend/internal/usecase"
)

var (
	ginRouter *gin.Engine
	initErr   error
	once      sync.Once
)

// Handler is the single Vercel function that receives all HTTP requests.
// vercel.json rewrites route every inbound path here; Gin handles routing internally.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() { ginRouter, initErr = buildRouter() })
	if initErr != nil {
		log.Printf("startup error: %v", initErr)
		http.Error(w,
			`{"error":"startup_error","message":"service unavailable"}`,
			http.StatusInternalServerError,
		)
		return
	}
	ginRouter.ServeHTTP(w, r)
}

// buildRouter wires all dependencies and returns a fully configured Gin engine.
// Its logic is intentionally identical to cmd/server/main.go so local and
// serverless environments behave the same way.
func buildRouter() (*gin.Engine, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	embedder, err := openaiAdapter.NewEmbedder(cfg.OpenAIAPIKey, cfg.EmbeddingModel, cfg.EmbeddingDimensions)
	if err != nil {
		return nil, err
	}

	llmClient, err := openaiAdapter.NewLLMClient(cfg.OpenAIAPIKey, cfg.ChatModel, cfg.MaxTokens)
	if err != nil {
		return nil, err
	}

	store, err := vectorstore.NewPineconeStore(
		cfg.PineconeAPIKey, cfg.PineconeIndexName, cfg.PineconeEnvironment, cfg.EmbeddingDimensions,
	)
	if err != nil {
		return nil, err
	}

	analyzer := usecase.NewAnalyzer(embedder, llmClient, store, linkgen.GenerateLinks, usecase.AnalyzerConfig{
		MaxProcessedMessages: cfg.MaxProcessedMessages,
		ChunkWindowSize:      cfg.ChunkWindowSize,
		ChunkOverlapSize:     cfg.ChunkOverlapSize,
		TopK:                 cfg.TopK,
		NumRetrievalQueries:  cfg.NumRetrievalQueries,
	})

	analyzeHandler := deliveryhttp.NewAnalyzeHandler(analyzer, cfg.MaxFileSizeBytes)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(deliveryhttp.CORS(cfg.AllowedOrigins))
	r.Use(deliveryhttp.RequestSizeLimiter(cfg.MaxFileSizeBytes))

	r.GET("/health", deliveryhttp.Health)
	v1 := r.Group("/api/v1")
	v1.POST("/analyze", analyzeHandler.Analyze)

	return r, nil
}
