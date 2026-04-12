package domain

import "errors"

// AudioInputType classifies what kind of audio was uploaded.
type AudioInputType string

const (
	AudioInputConversation AudioInputType = "CONVERSATION"
	AudioInputMonologue    AudioInputType = "MONOLOGUE"
	AudioInputSong         AudioInputType = "SONG"
	AudioInputUnknown      AudioInputType = "UNKNOWN"
)

// EmotionSignal is a single detected emotion from song lyrics.
type EmotionSignal struct {
	Name      string  `json:"name"`
	Emoji     string  `json:"emoji"`
	Intensity float64 `json:"intensity"` // 0.0–1.0
}

// AudioAnalysis holds the complete result of audio processing before gift analysis.
type AudioAnalysis struct {
	InputType     AudioInputType  `json:"input_type"`
	Transcript    string          `json:"transcript"`
	Emotions      []EmotionSignal `json:"emotions,omitempty"`
	LyricsSnippet string          `json:"lyrics_snippet,omitempty"`
	LanguageCode  string          `json:"language_code,omitempty"`
	LanguageLabel string          `json:"language_label,omitempty"`
}

var (
	ErrAudioFileTooLarge      = errors.New("audio file exceeds 5 MB — please upload a shorter clip")
	ErrAudioTooLong           = errors.New("recording exceeds 1-minute limit — please trim your audio and try again")
	ErrAudioInvalidFormat     = errors.New("unsupported audio format — accepted: .mp3 .wav .ogg .opus .m4a")
	ErrTranscriptionFailed    = errors.New("transcription failed — please try again")
	ErrTranscriberUnavailable = errors.New("transcription service is not configured")
)
