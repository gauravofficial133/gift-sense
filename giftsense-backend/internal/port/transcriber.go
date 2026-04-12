package port

import "context"

// TranscribeRequest holds the audio bytes and optional language hint.
type TranscribeRequest struct {
	Data         []byte
	Filename     string
	LanguageCode string // optional BCP-47 hint, e.g. "hi-IN"
}

// TranscribeResult is returned by a successful transcription.
type TranscribeResult struct {
	Transcript   string
	LanguageCode string
}

// Transcriber converts audio bytes to text.
type Transcriber interface {
	Transcribe(ctx context.Context, req TranscribeRequest) (TranscribeResult, error)
}
