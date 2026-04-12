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

// AudioHandler handles the audio transcription and analysis endpoints.
type AudioHandler struct {
	analyzer              *usecase.Analyzer
	transcriber           port.Transcriber
	audioMaxFileSizeBytes int64
}

// NewAudioHandler constructs the handler. transcriber may be nil if SARVAM_API_KEY is absent.
func NewAudioHandler(analyzer *usecase.Analyzer, transcriber port.Transcriber, audioMaxFileSizeBytes int64) *AudioHandler {
	return &AudioHandler{
		analyzer:              analyzer,
		transcriber:           transcriber,
		audioMaxFileSizeBytes: audioMaxFileSizeBytes,
	}
}

// AnalyzeAudio handles POST /api/v1/analyze-audio.
// It transcribes the audio, classifies it, and routes to the appropriate path.
func (h *AudioHandler) AnalyzeAudio(c *gin.Context) {
	if h.transcriber == nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "transcription_unavailable",
			Message: "Audio transcription is not configured on this server. Please use text upload instead.",
		})
		return
	}

	var req dto.AnalyzeAudioRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	fh, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "missing_file",
			Message: "audio file is required",
		})
		return
	}

	audioData, err := ValidateAudioFile(fh, h.audioMaxFileSizeBytes)
	if err != nil {
		h.handleAudioError(c, err)
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

	// 6-minute context: Sarvam batch jobs routinely take 3-4 min (upload to Azure + process + poll).
	// Remaining time covers GPT classify + RAG pipeline if audio is a conversation.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 360*time.Second)
	defer cancel()

	transcribeResult, err := h.transcriber.Transcribe(ctx, port.TranscribeRequest{
		Data:     audioData,
		Filename: fh.Filename,
	})
	if err != nil {
		h.handleAudioError(c, err)
		return
	}

	inputType, _, err := h.analyzer.ClassifyTranscript(ctx, transcribeResult.Transcript, transcribeResult.LanguageCode)
	if err != nil {
		h.handleAudioError(c, err)
		return
	}

	recipient := domain.RecipientDetails{
		Name:     req.Name,
		Relation: req.Relation,
		Gender:   req.Gender,
		Occasion: req.Occasion,
		Budget:   budget,
	}

	switch inputType {
	case domain.AudioInputSong:
		emotions, lyricsSnippet, languageLabel, err := h.analyzer.ExtractSongEmotions(ctx, transcribeResult.Transcript, transcribeResult.LanguageCode)
		if err != nil {
			h.handleAudioError(c, err)
			return
		}
		analysis := domain.AudioAnalysis{
			InputType:     domain.AudioInputSong,
			Transcript:    transcribeResult.Transcript,
			Emotions:      emotions,
			LyricsSnippet: lyricsSnippet,
			LanguageCode:  transcribeResult.LanguageCode,
			LanguageLabel: languageLabel,
		}
		c.JSON(http.StatusOK, gin.H{
			"audio_analysis": dto.AudioAnalysisToDTO(analysis),
			"message":        "Song detected — confirm emotions to continue",
		})

	case domain.AudioInputUnknown:
		analysis := domain.AudioAnalysis{
			InputType:    domain.AudioInputUnknown,
			Transcript:   transcribeResult.Transcript,
			LanguageCode: transcribeResult.LanguageCode,
		}
		c.JSON(http.StatusOK, gin.H{
			"audio_analysis": dto.AudioAnalysisToDTO(analysis),
			"message":        "Audio unclear — please confirm transcript",
		})

	default:
		// CONVERSATION or MONOLOGUE: run the full RAG pipeline immediately.
		analyzeCtx, analyzeCancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer analyzeCancel()

		result, err := h.analyzer.AnalyzeFromTranscript(analyzeCtx, req.SessionID, transcribeResult.Transcript, recipient, nil)
		if err != nil {
			h.handleAudioError(c, err)
			return
		}
		analysis := domain.AudioAnalysis{
			InputType:    inputType,
			LanguageCode: transcribeResult.LanguageCode,
		}
		c.JSON(http.StatusOK, dto.AnalyzeResponse{
			Data:          result,
			Message:       "Analysis complete",
			AudioAnalysis: dto.AudioAnalysisToDTO(analysis),
		})
	}
}

// AnalyzeFromTranscript handles POST /api/v1/analyze-from-transcript.
// Used as the second step in the song and unknown audio flows.
func (h *AudioHandler) AnalyzeFromTranscript(c *gin.Context) {
	var req dto.AnalyzeFromTranscriptRequest
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := h.analyzer.AnalyzeFromTranscript(ctx, req.SessionID, req.Transcript, recipient, confirmedEmotions)
	if err != nil {
		h.handleAudioError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AnalyzeResponse{
		Data:    result,
		Message: "Analysis complete",
	})
}

func (h *AudioHandler) handleAudioError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAudioTooLong):
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "audio_too_long", Message: err.Error()})
	case errors.Is(err, domain.ErrAudioFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, dto.ErrorResponse{Error: "audio_too_large", Message: err.Error()})
	case errors.Is(err, domain.ErrAudioInvalidFormat):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_audio_format", Message: err.Error()})
	case errors.Is(err, domain.ErrTranscriberUnavailable):
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: "transcription_unavailable", Message: err.Error()})
	case errors.Is(err, domain.ErrTranscriptionFailed):
		log.Printf("audio handler: transcription failed: %v", err)
		c.JSON(http.StatusBadGateway, dto.ErrorResponse{Error: "transcription_failed", Message: "Transcription failed — please try again"})
	case errors.Is(err, domain.ErrConversationTooShort):
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "transcript_too_short", Message: "Recording too short. Try a longer voice note."})
	case errors.Is(err, domain.ErrRetrievalFailed):
		log.Printf("audio handler: retrieval failed: %v", err)
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "retrieval_failed", Message: "Not enough context in the recording to generate suggestions. Try a longer or clearer clip."})
	case errors.Is(err, domain.ErrAllSuggestionsFiltered):
		log.Printf("audio handler: all suggestions filtered: %v", err)
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "suggestions_filtered", Message: "Could not find gift ideas within the selected budget. Try a different budget range."})
	case errors.Is(err, domain.ErrLLMResponseInvalid):
		log.Printf("audio handler: LLM response invalid: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "llm_error", Message: "Analysis failed, please try again"})
	default:
		log.Printf("audio handler: unexpected error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "An unexpected error occurred"})
	}
}
