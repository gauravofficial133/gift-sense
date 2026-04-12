package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/delivery/dto"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
	"github.com/giftsense/backend/internal/usecase"
)

// SpotifyHandler handles Spotify-related endpoints.
type SpotifyHandler struct {
	client    port.SpotifyClient
	songCache port.SongEmotionCache
	analyzer  *usecase.Analyzer
}

// NewSpotifyHandler constructs the handler. client and songCache may be nil.
func NewSpotifyHandler(client port.SpotifyClient, songCache port.SongEmotionCache, analyzer *usecase.Analyzer) *SpotifyHandler {
	return &SpotifyHandler{client: client, songCache: songCache, analyzer: analyzer}
}

// Search handles GET /api/v1/spotify/search?q=...
func (h *SpotifyHandler) Search(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "spotify_unavailable",
			Message: "Spotify integration is not configured on this server.",
		})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "missing_query",
			Message: "query parameter 'q' is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tracks, err := h.client.Search(ctx, query, 5)
	if err != nil {
		log.Printf("spotify search error: %v", err)
		c.JSON(http.StatusBadGateway, dto.ErrorResponse{
			Error:   "spotify_search_failed",
			Message: "Failed to search Spotify. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SpotifySearchResponse{
		Tracks: dto.SpotifyTracksToDTO(tracks),
	})
}

// GetAudioFeatures handles GET /api/v1/spotify/track/:id/features
func (h *SpotifyHandler) GetAudioFeatures(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "spotify_unavailable",
			Message: "Spotify integration is not configured on this server.",
		})
		return
	}

	trackID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	features, err := h.client.GetAudioFeatures(ctx, trackID)
	if err != nil {
		log.Printf("spotify audio features error: %v", err)
		c.JSON(http.StatusBadGateway, dto.ErrorResponse{
			Error:   "spotify_features_failed",
			Message: "Failed to get audio features.",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SpotifyAudioFeaturesToDTO(features))
}

// AnalyzeSong handles POST /api/v1/spotify/analyze-song.
// It extracts emotions from a Spotify song (using cache or LLM) and returns them
// so the frontend can show the emotion card.
func (h *SpotifyHandler) AnalyzeSong(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "spotify_unavailable",
			Message: "Spotify integration is not configured on this server.",
		})
		return
	}

	var req dto.AnalyzeSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Check cache first.
	if h.songCache != nil {
		cached, err := h.songCache.Get(ctx, req.TrackID)
		if err != nil {
			log.Printf("song cache get error: %v", err)
		}
		if cached != nil {
			emotionDTOs := make([]dto.EmotionSignalDTO, len(cached.Emotions))
			for i, e := range cached.Emotions {
				emotionDTOs[i] = dto.EmotionSignalDTO{Name: e.Name, Emoji: e.Emoji, Intensity: e.Intensity}
			}
			c.JSON(http.StatusOK, dto.AnalyzeSongResponse{
				TrackName:     cached.TrackName,
				Artist:        cached.Artist,
				Emotions:      emotionDTOs,
				LyricsSnippet: cached.LyricsSnippet,
				LanguageLabel: cached.LanguageLabel,
				Cached:        true,
			})
			return
		}
	}

	// Fetch audio features from Spotify.
	features, err := h.client.GetAudioFeatures(ctx, req.TrackID)
	if err != nil {
		log.Printf("spotify audio features error for analyze-song: %v", err)
		// Continue with zero-valued features — the LLM can infer from name + artist.
	}

	// Call LLM to extract emotions.
	emotions, lyricsSnippet, languageLabel, err := h.analyzer.ExtractSongEmotionsFromSpotify(ctx, req.TrackName, req.Artist, features)
	if err != nil {
		log.Printf("spotify song emotion extraction error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "emotion_extraction_failed",
			Message: "Failed to analyze song emotions. Please try again.",
		})
		return
	}

	// Save to cache (fire-and-forget, don't block response).
	if h.songCache != nil {
		go func() {
			saveErr := h.songCache.Save(context.Background(), domain.CachedSongEmotion{
				TrackID:       req.TrackID,
				TrackName:     req.TrackName,
				Artist:        req.Artist,
				Emotions:      emotions,
				LyricsSnippet: lyricsSnippet,
				LanguageLabel: languageLabel,
				AudioFeatures: features,
			})
			if saveErr != nil {
				log.Printf("song cache save error: %v", saveErr)
			}
		}()
	}

	emotionDTOs := make([]dto.EmotionSignalDTO, len(emotions))
	for i, e := range emotions {
		emotionDTOs[i] = dto.EmotionSignalDTO{Name: e.Name, Emoji: e.Emoji, Intensity: e.Intensity}
	}

	c.JSON(http.StatusOK, dto.AnalyzeSongResponse{
		TrackName:     req.TrackName,
		Artist:        req.Artist,
		Emotions:      emotionDTOs,
		LyricsSnippet: lyricsSnippet,
		LanguageLabel: languageLabel,
		Cached:        false,
	})
}

// AnalyzeFromSong handles POST /api/v1/analyze-from-song.
// It generates gift suggestions from confirmed song emotions without RAG.
func (h *SpotifyHandler) AnalyzeFromSong(c *gin.Context) {
	var req dto.AnalyzeFromSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	budgetTier := domain.BudgetTier(req.BudgetTier)
	budget, ok := domain.BudgetRanges[budgetTier]
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_budget_tier",
			Message: "unknown budget tier",
		})
		return
	}

	recipient := domain.RecipientDetails{
		Name:     req.Name,
		Relation: req.Relation,
		Gender:   req.Gender,
		Occasion: req.Occasion,
		Budget:   budget,
	}

	var confirmedEmotions []domain.EmotionSignal
	for _, e := range req.ConfirmedEmotions {
		confirmedEmotions = append(confirmedEmotions, domain.EmotionSignal{
			Name:      e.Name,
			Emoji:     e.Emoji,
			Intensity: e.Intensity,
		})
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	result, err := h.analyzer.AnalyzeFromSongEmotions(ctx, recipient, req.TrackName, req.Artist, confirmedEmotions)
	if err != nil {
		h.handleSongError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AnalyzeResponse{
		Data:    result,
		Message: "Analysis complete",
	})
}

func (h *SpotifyHandler) handleSongError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAllSuggestionsFiltered):
		log.Printf("song handler: all suggestions filtered: %v", err)
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "suggestions_filtered", Message: "Could not find gift ideas within the selected budget. Try a different budget range."})
	case errors.Is(err, domain.ErrLLMResponseInvalid):
		log.Printf("song handler: LLM response invalid: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "llm_error", Message: "Analysis failed, please try again"})
	default:
		log.Printf("song handler: unexpected error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "An unexpected error occurred"})
	}
}
