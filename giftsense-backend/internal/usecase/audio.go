package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

const classifySystemPrompt = `You are an audio content classifier. Analyze the provided transcript and classify it.
RULES:
1. Respond ONLY with valid JSON: {"input_type": "SONG"|"CONVERSATION"|"MONOLOGUE"|"UNKNOWN", "confidence": 0.0-1.0}
2. SONG: repeated phrases/chorus, lyrical/poetic structure, rhyming, emotional metaphorical language typical of music
3. CONVERSATION: multiple distinct speaking turns or exchanges, even transcribed as one block
4. MONOLOGUE: single continuous speaker narrative (voice note, speech) without conversational turns
5. UNKNOWN: under 10 words, unintelligible noise, no discernible speech
6. When uncertain between CONVERSATION and MONOLOGUE, prefer MONOLOGUE`

const emotionSystemPrompt = `You are a music emotion analyst. Extract the emotional essence of song lyrics.
RULES:
1. Respond ONLY with valid JSON: {"emotions": [{"name": "string", "emoji": "string", "intensity": 0.0-1.0}], "lyrics_snippet": "string", "language_label": "string"}
2. emotions: max 5 items. Choose names from: Deep Love, Heartbreak, Longing, Joy, Nostalgia, Passion, Melancholy, Hope, Devotion, Playfulness, Yearning, Tenderness, Celebration, Grief, Warmth
3. intensity: 0.0 = barely present, 1.0 = overwhelmingly dominant
4. lyrics_snippet: single representative line/phrase, max 80 characters
5. language_label: human-readable (e.g. "Hindi", "English", "Tamil")
6. If no lyrics detectable: {"emotions": [], "lyrics_snippet": "", "language_label": "Unknown"}`

var sentenceSplitter = regexp.MustCompile(`[.!?]+\s+`)

// ChunkTranscript splits a prose transcript into overlapping word-window chunks.
// It lives in the usecase package so it can call the unexported enrichMetadata().
func ChunkTranscript(sessionID, transcript string, windowWords, overlapWords int) []domain.Chunk {
	sentences := sentenceSplitter.Split(strings.TrimSpace(transcript), -1)

	var allWords []string
	for _, s := range sentences {
		words := strings.Fields(s)
		allWords = append(allWords, words...)
	}

	if len(allWords) == 0 {
		return nil
	}

	step := windowWords - overlapWords
	if step <= 0 {
		step = windowWords
	}

	var chunks []domain.Chunk
	chunkIndex := 0

	for start := 0; start < len(allWords); start += step {
		end := min(start+windowWords, len(allWords))

		text := strings.Join(allWords[start:end], " ")
		chunks = append(chunks, domain.Chunk{
			ID:             fmt.Sprintf("%s_chunk_%d", sessionID, chunkIndex),
			SessionID:      sessionID,
			AnonymizedText: text,
			StartIndex:     start,
			EndIndex:       end - 1,
			Metadata:       enrichMetadata(text),
		})
		chunkIndex++

		if end == len(allWords) {
			break
		}
	}

	// Ensure minimum one chunk even for very short transcripts.
	if len(chunks) == 0 {
		text := strings.Join(allWords, " ")
		chunks = []domain.Chunk{{
			ID:             fmt.Sprintf("%s_chunk_0", sessionID),
			SessionID:      sessionID,
			AnonymizedText: text,
			StartIndex:     0,
			EndIndex:       len(allWords) - 1,
			Metadata:       enrichMetadata(text),
		}}
	}

	return chunks
}

type classifyResponse struct {
	InputType  string  `json:"input_type"`
	Confidence float64 `json:"confidence"`
}

// ClassifyTranscript calls GPT to classify the transcript as SONG/CONVERSATION/MONOLOGUE/UNKNOWN.
func (a *Analyzer) ClassifyTranscript(ctx context.Context, transcript, languageCode string) (domain.AudioInputType, float64, error) {
	userMsg := fmt.Sprintf("Transcript (language: %s):\n%s", languageCode, transcript)

	raw, err := a.llm.Complete(ctx, userMsg, port.CompletionOptions{
		JSONMode:     true,
		SystemPrompt: classifySystemPrompt,
	})
	if err != nil {
		return domain.AudioInputMonologue, 0, fmt.Errorf("classify transcript: %w", err)
	}

	var resp classifyResponse
	if err = json.Unmarshal([]byte(raw), &resp); err != nil {
		// Safe fallback on parse error.
		return domain.AudioInputMonologue, 0.5, nil
	}

	switch domain.AudioInputType(resp.InputType) {
	case domain.AudioInputConversation, domain.AudioInputMonologue, domain.AudioInputSong, domain.AudioInputUnknown:
		return domain.AudioInputType(resp.InputType), resp.Confidence, nil
	default:
		return domain.AudioInputMonologue, resp.Confidence, nil
	}
}

type emotionResponse struct {
	Emotions      []domain.EmotionSignal `json:"emotions"`
	LyricsSnippet string                 `json:"lyrics_snippet"`
	LanguageLabel string                 `json:"language_label"`
}

// ExtractSongEmotions calls GPT to extract emotional signals from song lyrics.
func (a *Analyzer) ExtractSongEmotions(ctx context.Context, transcript, languageCode string) ([]domain.EmotionSignal, string, string, error) {
	userMsg := fmt.Sprintf("Song lyrics (language: %s):\n%s", languageCode, transcript)

	raw, err := a.llm.Complete(ctx, userMsg, port.CompletionOptions{
		JSONMode:     true,
		SystemPrompt: emotionSystemPrompt,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("extract song emotions: %w", err)
	}

	var resp emotionResponse
	if err = json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, "", "", fmt.Errorf("parse emotion response: %w", err)
	}

	return resp.Emotions, resp.LyricsSnippet, resp.LanguageLabel, nil
}

// ExtractSongEmotionsFromSpotify calls GPT to extract emotional signals using
// the song's name, artist, and Spotify audio features as context.
func (a *Analyzer) ExtractSongEmotionsFromSpotify(ctx context.Context, trackName, artist string, features domain.SpotifyAudioFeatures) ([]domain.EmotionSignal, string, string, error) {
	var userMsg string
	// When audio features are all zero (e.g. the endpoint returned 403/deprecated),
	// omit the numeric profile so the LLM doesn't treat 0-valence as "maximally sad".
	if features.Valence == 0 && features.Energy == 0 && features.Danceability == 0 && features.Tempo == 0 {
		userMsg = fmt.Sprintf(
			"Song: \"%s\" by %s. Analyze the emotional essence of this song based on its title and artist.",
			trackName, artist,
		)
	} else {
		userMsg = fmt.Sprintf(
			"Song: \"%s\" by %s.\nAudio profile: valence=%.2f (0=sad, 1=happy), energy=%.2f, danceability=%.2f, tempo=%.0f BPM.\nAnalyze the emotional essence of this song.",
			trackName, artist, features.Valence, features.Energy, features.Danceability, features.Tempo,
		)
	}

	raw, err := a.llm.Complete(ctx, userMsg, port.CompletionOptions{
		JSONMode:     true,
		SystemPrompt: emotionSystemPrompt,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("extract spotify song emotions: %w", err)
	}

	var resp emotionResponse
	if err = json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, "", "", fmt.Errorf("parse spotify emotion response: %w", err)
	}

	return resp.Emotions, resp.LyricsSnippet, resp.LanguageLabel, nil
}
