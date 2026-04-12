package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/giftsense/backend/internal/delivery/dto"
	"github.com/giftsense/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ────────────────────────────────────────────────────────────────
// AudioAnalysisToDTO — unit tests
// ────────────────────────────────────────────────────────────────

func TestAudioAnalysisToDTO_ShouldMapAllFields_WhenFullAnalysisProvided(t *testing.T) {
	analysis := domain.AudioAnalysis{
		InputType:     domain.AudioInputSong,
		Transcript:    "Tujhe kitna chahne lage hum",
		LyricsSnippet: "Tujhe kitna chahne lage hum...",
		LanguageCode:  "hi-IN",
		LanguageLabel: "Hindi",
		Emotions: []domain.EmotionSignal{
			{Name: "Deep Love", Emoji: "❤️", Intensity: 0.9},
			{Name: "Nostalgia", Emoji: "🌸", Intensity: 0.7},
		},
	}

	result := dto.AudioAnalysisToDTO(analysis)

	require.NotNil(t, result)
	assert.Equal(t, "SONG", result.InputType)
	assert.Equal(t, analysis.Transcript, result.Transcript)
	assert.Equal(t, analysis.LyricsSnippet, result.LyricsSnippet)
	assert.Equal(t, analysis.LanguageCode, result.LanguageCode)
	assert.Equal(t, analysis.LanguageLabel, result.LanguageLabel)
	require.Len(t, result.Emotions, 2)
	assert.Equal(t, "Deep Love", result.Emotions[0].Name)
	assert.Equal(t, "❤️", result.Emotions[0].Emoji)
	assert.InDelta(t, 0.9, result.Emotions[0].Intensity, 0.001)
	assert.Equal(t, "Nostalgia", result.Emotions[1].Name)
}

func TestAudioAnalysisToDTO_ShouldReturnNilEmotions_WhenNoEmotionsProvided(t *testing.T) {
	analysis := domain.AudioAnalysis{
		InputType:  domain.AudioInputConversation,
		Transcript: "Hey what gift should I get her",
	}

	result := dto.AudioAnalysisToDTO(analysis)

	require.NotNil(t, result)
	assert.Nil(t, result.Emotions, "emotions should be nil when not present in domain model")
}

func TestAudioAnalysisToDTO_ShouldPreserveIntensityPrecision_WhenFractionalValue(t *testing.T) {
	analysis := domain.AudioAnalysis{
		InputType: domain.AudioInputSong,
		Emotions: []domain.EmotionSignal{
			{Name: "Joy", Emoji: "☀️", Intensity: 0.123456789},
		},
	}

	result := dto.AudioAnalysisToDTO(analysis)

	require.Len(t, result.Emotions, 1)
	assert.InDelta(t, 0.123456789, result.Emotions[0].Intensity, 1e-9)
}

func TestAudioAnalysisToDTO_ShouldSetInputType_AsStringRepresentation(t *testing.T) {
	cases := []struct {
		inputType domain.AudioInputType
		expected  string
	}{
		{domain.AudioInputConversation, "CONVERSATION"},
		{domain.AudioInputMonologue, "MONOLOGUE"},
		{domain.AudioInputSong, "SONG"},
		{domain.AudioInputUnknown, "UNKNOWN"},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			result := dto.AudioAnalysisToDTO(domain.AudioAnalysis{InputType: tc.inputType})
			assert.Equal(t, tc.expected, result.InputType)
		})
	}
}

// ────────────────────────────────────────────────────────────────
// JSON serialization — omitempty behaviour
// ────────────────────────────────────────────────────────────────

func TestAudioAnalysisDTO_ShouldOmitOptionalFields_WhenEmpty(t *testing.T) {
	analysis := domain.AudioAnalysis{
		InputType: domain.AudioInputConversation,
		// All optional fields are zero-value
	}

	result := dto.AudioAnalysisToDTO(analysis)
	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes, &m))

	assert.Contains(t, m, "input_type", "input_type must always be present")
	assert.NotContains(t, m, "transcript", "transcript should be omitted when empty")
	assert.NotContains(t, m, "emotions", "emotions should be omitted when nil")
	assert.NotContains(t, m, "lyrics_snippet", "lyrics_snippet should be omitted when empty")
	assert.NotContains(t, m, "language_code", "language_code should be omitted when empty")
	assert.NotContains(t, m, "language_label", "language_label should be omitted when empty")
}

func TestAudioAnalysisDTO_ShouldIncludeAllFields_WhenFullyPopulated(t *testing.T) {
	analysis := domain.AudioAnalysis{
		InputType:     domain.AudioInputSong,
		Transcript:    "lyrics here",
		LyricsSnippet: "snippet",
		LanguageCode:  "hi-IN",
		LanguageLabel: "Hindi",
		Emotions:      []domain.EmotionSignal{{Name: "Joy", Emoji: "☀️", Intensity: 0.8}},
	}

	result := dto.AudioAnalysisToDTO(analysis)
	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes, &m))

	for _, key := range []string{"input_type", "transcript", "emotions", "lyrics_snippet", "language_code", "language_label"} {
		assert.Contains(t, m, key, "field %q should be present", key)
	}
}

func TestEmotionSignalDTO_ShouldSerializeToJSON_WithCorrectFieldNames(t *testing.T) {
	e := dto.EmotionSignalDTO{Name: "Deep Love", Emoji: "❤️", Intensity: 0.95}
	jsonBytes, err := json.Marshal(e)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes, &m))

	assert.Equal(t, "Deep Love", m["name"])
	assert.Equal(t, "❤️", m["emoji"])
	assert.InDelta(t, 0.95, m["intensity"], 0.001)
}

// ────────────────────────────────────────────────────────────────
// AnalyzeFromTranscriptRequest — JSON round-trip
// ────────────────────────────────────────────────────────────────

func TestAnalyzeFromTranscriptRequest_ShouldDeserialize_WhenValidJSON(t *testing.T) {
	raw := `{
		"session_id": "550e8400-e29b-41d4-a716-446655440000",
		"transcript": "Hey I was thinking about getting her a pottery kit",
		"name": "Priya",
		"relation": "sister",
		"gender": "female",
		"occasion": "birthday",
		"budget_tier": "MID_RANGE",
		"confirmed_emotions": [
			{"name": "Joy", "emoji": "☀️", "intensity": 0.8},
			{"name": "Warmth", "emoji": "🤗", "intensity": 0.6}
		]
	}`

	var req dto.AnalyzeFromTranscriptRequest
	err := json.Unmarshal([]byte(raw), &req)

	require.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.SessionID)
	assert.Equal(t, "Priya", req.Name)
	assert.Equal(t, "MID_RANGE", req.BudgetTier)
	require.Len(t, req.ConfirmedEmotions, 2)
	assert.Equal(t, "Joy", req.ConfirmedEmotions[0].Name)
	assert.InDelta(t, 0.8, req.ConfirmedEmotions[0].Intensity, 0.001)
}

func TestAnalyzeFromTranscriptRequest_ShouldDeserialize_WhenNoEmotionsProvided(t *testing.T) {
	raw := `{
		"session_id": "550e8400-e29b-41d4-a716-446655440000",
		"transcript": "She loves cooking and food",
		"name": "Meera",
		"occasion": "anniversary",
		"budget_tier": "PREMIUM"
	}`

	var req dto.AnalyzeFromTranscriptRequest
	err := json.Unmarshal([]byte(raw), &req)

	require.NoError(t, err)
	assert.Empty(t, req.ConfirmedEmotions)
}

// ────────────────────────────────────────────────────────────────
// Integration test — AudioAnalysisToDTO round-trip through JSON
// ────────────────────────────────────────────────────────────────

// TestAudioAnalysisToDTO_Integration_ShouldRoundTripThroughJSON verifies that a
// domain.AudioAnalysis → DTO → JSON → map chain produces exactly the expected
// wire format consumed by the frontend.
func TestAudioAnalysisToDTO_Integration_ShouldRoundTripThroughJSON(t *testing.T) {
	analysis := domain.AudioAnalysis{
		InputType:     domain.AudioInputSong,
		Transcript:    "Tujhe kitna chahne lage hum",
		LyricsSnippet: "Tujhe kitna chahne lage hum...",
		LanguageCode:  "hi-IN",
		LanguageLabel: "Hindi",
		Emotions: []domain.EmotionSignal{
			{Name: "Deep Love", Emoji: "❤️", Intensity: 0.9},
			{Name: "Longing", Emoji: "💫", Intensity: 0.75},
			{Name: "Nostalgia", Emoji: "🌸", Intensity: 0.6},
		},
	}

	// Convert to DTO then marshal to JSON
	dtoObj := dto.AudioAnalysisToDTO(analysis)
	jsonBytes, err := json.Marshal(dtoObj)
	require.NoError(t, err)

	// Unmarshal back into a fresh DTO
	var recovered dto.AudioAnalysisDTO
	require.NoError(t, json.Unmarshal(jsonBytes, &recovered))

	assert.Equal(t, "SONG", recovered.InputType)
	assert.Equal(t, analysis.Transcript, recovered.Transcript)
	assert.Equal(t, analysis.LyricsSnippet, recovered.LyricsSnippet)
	assert.Equal(t, analysis.LanguageCode, recovered.LanguageCode)
	assert.Equal(t, analysis.LanguageLabel, recovered.LanguageLabel)
	require.Len(t, recovered.Emotions, 3)

	for i, expected := range analysis.Emotions {
		assert.Equal(t, expected.Name, recovered.Emotions[i].Name)
		assert.Equal(t, expected.Emoji, recovered.Emotions[i].Emoji)
		assert.InDelta(t, expected.Intensity, recovered.Emotions[i].Intensity, 0.001)
	}
}

// TestAudioAnalysisToDTO_Integration_ShouldProduceValidWireFormat verifies the
// exact JSON field names and structure the frontend expects for a SONG response.
func TestAudioAnalysisToDTO_Integration_ShouldProduceValidWireFormat(t *testing.T) {
	analysis := domain.AudioAnalysis{
		InputType:     domain.AudioInputSong,
		Transcript:    "some lyrics",
		LyricsSnippet: "snippet here",
		LanguageCode:  "hi-IN",
		LanguageLabel: "Hindi",
		Emotions:      []domain.EmotionSignal{{Name: "Joy", Emoji: "☀️", Intensity: 0.8}},
	}

	jsonBytes, err := json.Marshal(dto.AudioAnalysisToDTO(analysis))
	require.NoError(t, err)

	// Decode to generic map to assert exact wire keys
	var wire map[string]any
	require.NoError(t, json.Unmarshal(jsonBytes, &wire))

	assert.Equal(t, "SONG", wire["input_type"])
	assert.Equal(t, "some lyrics", wire["transcript"])
	assert.Equal(t, "snippet here", wire["lyrics_snippet"])
	assert.Equal(t, "hi-IN", wire["language_code"])
	assert.Equal(t, "Hindi", wire["language_label"])

	emotions, ok := wire["emotions"].([]any)
	require.True(t, ok)
	require.Len(t, emotions, 1)

	e := emotions[0].(map[string]any)
	assert.Equal(t, "Joy", e["name"])
	assert.Equal(t, "☀️", e["emoji"])
	assert.InDelta(t, 0.8, e["intensity"], 0.001)
}
