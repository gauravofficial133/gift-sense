package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/config"
	"github.com/giftsense/backend/internal/adapter/feedbackstore"
	"github.com/giftsense/backend/internal/adapter/linkgen"
	openaiAdapter "github.com/giftsense/backend/internal/adapter/openai"
	"github.com/giftsense/backend/internal/adapter/ratelimiter"
	"github.com/giftsense/backend/internal/adapter/vectorstore"
	"github.com/giftsense/backend/internal/database"
	"github.com/giftsense/backend/internal/database/migration"
	deliveryhttp "github.com/giftsense/backend/internal/delivery/http"
	"github.com/giftsense/backend/internal/usecase"
)

var (
	ginRouter *gin.Engine
	initErr   error
	once      sync.Once
)

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

	if cfg.HasDatabase() {
		db, dbErr := database.Connect(cfg.DatabaseURL)
		if dbErr != nil {
			log.Printf("database connection failed, feedback/rate-limit disabled: %v", dbErr)
			v1.POST("/analyze", analyzeHandler.Analyze)
			return r, nil
		}

		if migErr := migration.RunMigrations(db); migErr != nil {
			log.Printf("migrations failed, feedback/rate-limit disabled: %v", migErr)
			v1.POST("/analyze", analyzeHandler.Analyze)
			return r, nil
		}

		rateLimiter := ratelimiter.NewDBRateLimiter(db, cfg.RateLimitPerMinute)
		v1.POST("/analyze", deliveryhttp.RateLimit(rateLimiter), analyzeHandler.Analyze)

		fbStore := feedbackstore.NewGormFeedbackStore(db)
		fbService := usecase.NewFeedbackService(fbStore)
		fbHandler := deliveryhttp.NewFeedbackHandler(fbService)

		v1.POST("/feedback", fbHandler.SubmitFeedback)
		v1.POST("/events", fbHandler.TrackEvent)

		log.Println("Feedback + rate limiting enabled (DATABASE_URL configured)")
	} else {
		v1.POST("/analyze", analyzeHandler.Analyze)
		log.Println("Feedback + rate limiting disabled (DATABASE_URL not set)")
	}

	return r, nil
}
