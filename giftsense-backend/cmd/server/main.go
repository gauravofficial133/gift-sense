package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/giftsense/backend/config"
	"github.com/giftsense/backend/internal/adapter/feedbackstore"
	"github.com/giftsense/backend/internal/adapter/linkgen"
	openaiAdapter "github.com/giftsense/backend/internal/adapter/openai"
	"github.com/giftsense/backend/internal/adapter/ratelimiter"
	sarvamAdapter "github.com/giftsense/backend/internal/adapter/sarvam"
	"github.com/giftsense/backend/internal/adapter/songcache"
	spotifyAdapter "github.com/giftsense/backend/internal/adapter/spotify"
	"github.com/giftsense/backend/internal/adapter/vectorstore"
	"github.com/giftsense/backend/internal/database"
	"github.com/giftsense/backend/internal/database/migration"
	handler "github.com/giftsense/backend/internal/delivery/http"
	"github.com/giftsense/backend/internal/port"
	"github.com/giftsense/backend/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

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

	var transcriber port.Transcriber
	if cfg.SarvamAPIKey != "" {
		transcriber = sarvamAdapter.NewTranscriber(cfg.SarvamAPIKey)
		log.Println("Sarvam transcription enabled")
	} else {
		log.Println("Sarvam transcription disabled (SARVAM_API_KEY not set)")
	}
	audioHandler := handler.NewAudioHandler(analyzer, transcriber, cfg.AudioMaxFileSizeBytes)

	var spotifyClient port.SpotifyClient
	if cfg.HasSpotify() {
		spotifyClient = spotifyAdapter.NewClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
		log.Println("Spotify integration enabled")
	} else {
		log.Println("Spotify integration disabled (SPOTIFY_CLIENT_ID/SPOTIFY_CLIENT_SECRET not set)")
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(handler.CORS(cfg.AllowedOrigins))
	router.Use(handler.RequestSizeLimiter(cfg.MaxFileSizeBytes))

	router.GET("/health", handler.Health)

	v1 := router.Group("/api/v1")

	// Audio routes use a separate 5MB size limiter.
	audioRoutes := v1.Group("/")
	audioRoutes.Use(handler.RequestSizeLimiter(cfg.AudioMaxFileSizeBytes))
	audioRoutes.POST("/analyze-audio", audioHandler.AnalyzeAudio)
	audioRoutes.POST("/analyze-from-transcript", audioHandler.AnalyzeFromTranscript)

	var songCache port.SongEmotionCache
	if cfg.HasDatabase() {
		db, dbErr := database.Connect(cfg.DatabaseURL)
		if dbErr != nil {
			log.Fatalf("database: %v", dbErr)
		}

		if migErr := migration.RunMigrations(db); migErr != nil {
			log.Fatalf("migrations: %v", migErr)
		}

		rateLimiter := ratelimiter.NewDBRateLimiter(db, cfg.RateLimitPerMinute)
		v1.POST("/analyze", handler.RateLimit(rateLimiter), analyzeHandler.Analyze)

		fbStore := feedbackstore.NewGormFeedbackStore(db)
		fbService := usecase.NewFeedbackService(fbStore)
		fbHandler := handler.NewFeedbackHandler(fbService)

		v1.POST("/feedback", fbHandler.SubmitFeedback)
		v1.POST("/events", fbHandler.TrackEvent)

		if cfg.HasSpotify() {
			songCache = songcache.NewGormSongCache(db)
			log.Println("Spotify song emotion cache enabled (DATABASE_URL + SPOTIFY configured)")
		}

		log.Println("Feedback + rate limiting enabled (DATABASE_URL configured)")
	} else {
		v1.POST("/analyze", analyzeHandler.Analyze)
		log.Println("Feedback + rate limiting disabled (DATABASE_URL not set)")
	}

	spotifyHandler := handler.NewSpotifyHandler(spotifyClient, songCache, analyzer)
	v1.GET("/spotify/search", spotifyHandler.Search)
	v1.GET("/spotify/track/:id/features", spotifyHandler.GetAudioFeatures)
	v1.POST("/spotify/analyze-song", spotifyHandler.AnalyzeSong)
	v1.POST("/analyze-from-song", spotifyHandler.AnalyzeFromSong)

	addr := ":" + cfg.Port
	log.Printf("upahaar.ai backend listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
