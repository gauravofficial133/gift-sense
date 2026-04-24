package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/giftsense/backend/assets/webfonts"
	"github.com/giftsense/backend/config"
	"github.com/giftsense/backend/internal/adapter/cardrender"
	"github.com/giftsense/backend/internal/adapter/feedbackstore"
	"github.com/giftsense/backend/internal/adapter/linkgen"
	"github.com/giftsense/backend/internal/adapter/templatestore"
	anthropicAdapter "github.com/giftsense/backend/internal/adapter/anthropic"
	imagegenAdapter "github.com/giftsense/backend/internal/adapter/imagegen"
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

	var anthropicLLM port.LLMClient
	if cfg.HasAnthropic() {
		anthropicLLM, err = anthropicAdapter.NewLLMClient(cfg.AnthropicAPIKey, cfg.AnthropicModel, cfg.MaxTokens)
		if err != nil {
			log.Fatalf("anthropic llm client: %v", err)
		}
		log.Println("Anthropic LLM enabled")
	} else {
		log.Println("Anthropic LLM disabled (ANTHROPIC_API_KEY not set)")
	}

	store, err := vectorstore.NewPineconeStore(cfg.PineconeAPIKey, cfg.PineconeIndexName, cfg.PineconeEnvironment, cfg.EmbeddingDimensions)
	if err != nil {
		log.Fatalf("vector store: %v", err)
	}

	chromePool, err := cardrender.NewChromePool()
	if err != nil {
		log.Printf("WARNING: Chrome rendering disabled: %v", err)
	}
	var renderer *cardrender.Renderer
	if chromePool != nil {
		defer chromePool.Close()
		renderer = cardrender.NewRenderer(chromePool)
	}

	engine := cardrender.NewTemplateEngine()
	engine.RegisterFont("Great Vibes", "400", "normal", webfonts.GreatVibesRegular)
	engine.RegisterFont("Cormorant Garamond", "400", "normal", webfonts.CormorantGaramondRegular)
	engine.RegisterFont("Cormorant Garamond", "700", "normal", webfonts.CormorantGaramondBold)
	engine.RegisterFont("Abril Fatface", "400", "normal", webfonts.AbrilFatfaceRegular)
	engine.RegisterFont("Quicksand", "400", "normal", webfonts.QuicksandRegular)
	engine.RegisterFont("Source Serif 4", "400", "normal", webfonts.SourceSerif4Regular)
	engine.RegisterFont("Dancing Script", "400", "normal", webfonts.DancingScriptRegular)
	engine.RegisterFont("Playfair Display", "700", "normal", webfonts.PlayfairDisplayBold)

	log.Println("Card rendering engine initialized (fonts loaded)")

	var imageGen port.ImageGenerator
	if cfg.OpenAIImageModel != "" {
		ig, igErr := imagegenAdapter.NewDallEGenerator(cfg.OpenAIAPIKey, cfg.OpenAIImageModel)
		if igErr != nil {
			log.Printf("WARNING: Image generation disabled: %v", igErr)
		} else {
			imageGen = ig
			log.Printf("DALL-E image generation enabled (model: %s)", cfg.OpenAIImageModel)
		}
	} else {
		log.Println("Image generation disabled (OPENAI_IMAGE_MODEL not set)")
	}

	var assetLib *usecase.AssetLibrary
	assetLib, assetErr := usecase.NewAssetLibrary("assets/generated", imageGen)
	if assetErr != nil {
		log.Printf("WARNING: Asset library disabled: %v", assetErr)
		assetLib = nil
	}

	tplStore, err := templatestore.NewFSStore("assets/templates")
	if err != nil {
		log.Printf("WARNING: Template store disabled: %v", err)
	}
	tplManager := usecase.NewTemplateManager(tplStore)
	htmlCompiler := usecase.NewHTMLCompiler(engine.Fonts(), assetLib)

	cardGen := usecase.NewCardGenerator(llmClient, anthropicLLM, renderer, engine, imageGen, assetLib)
	if tplStore != nil {
		cardGen.SetTemplateStore(tplStore, htmlCompiler)
	}

	analyzer := usecase.NewAnalyzer(embedder, llmClient, store, linkgen.GenerateLinks, usecase.AnalyzerConfig{
		MaxProcessedMessages: cfg.MaxProcessedMessages,
		ChunkWindowSize:      cfg.ChunkWindowSize,
		ChunkOverlapSize:     cfg.ChunkOverlapSize,
		TopK:                 cfg.TopK,
		NumRetrievalQueries:  cfg.NumRetrievalQueries,
	}, renderer, engine, anthropicLLM, imageGen, assetLib)

	if tplStore != nil {
		analyzer.SetTemplateStore(tplStore, htmlCompiler)
	}

	analyzeHandler := handler.NewAnalyzeHandler(analyzer, cfg.MaxFileSizeBytes)
	cardHandler := handler.NewCardHandler(cardGen)

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
	router.Use(gin.Logger())
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
	v1.POST("/cards/download", cardHandler.Download)
	v1.POST("/cards/re-render", cardHandler.ReRender)
	v1.GET("/cards/palettes", cardHandler.ListPalettes)

	adminTemplateHandler := handler.NewAdminTemplateHandler(tplManager, htmlCompiler, renderer)
	admin := v1.Group("/admin")
	admin.GET("/templates", adminTemplateHandler.List)
	admin.GET("/templates/:id", adminTemplateHandler.Get)
	admin.POST("/templates", adminTemplateHandler.Create)
	admin.PUT("/templates/:id", adminTemplateHandler.Update)
	admin.DELETE("/templates/:id", adminTemplateHandler.Delete)
	admin.POST("/templates/:id/preview", adminTemplateHandler.Preview)
	admin.POST("/templates/:id/duplicate", adminTemplateHandler.Duplicate)

	assetPlanner := usecase.NewAssetPromptPlanner(llmClient)
	adminAssetHandler := handler.NewAdminAssetHandler(assetLib, assetPlanner, imageGen)
	admin.GET("/assets", adminAssetHandler.List)
	admin.POST("/assets/plan-prompt", adminAssetHandler.PlanPrompt)
	admin.POST("/assets/generate", adminAssetHandler.Generate)
	admin.POST("/assets/upload", adminAssetHandler.Upload)

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
