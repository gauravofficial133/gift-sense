package dto

import "github.com/giftsense/backend/internal/domain"

// AudioAnalysisDTO is the wire representation of an AudioAnalysis result.
type AudioAnalysisDTO struct {
	InputType     string             `json:"input_type"`
	Transcript    string             `json:"transcript,omitempty"`
	Emotions      []EmotionSignalDTO `json:"emotions,omitempty"`
	LyricsSnippet string             `json:"lyrics_snippet,omitempty"`
	LanguageCode  string             `json:"language_code,omitempty"`
	LanguageLabel string             `json:"language_label,omitempty"`
}

// AudioAnalysisToDTO converts a domain.AudioAnalysis to its DTO representation.
func AudioAnalysisToDTO(a domain.AudioAnalysis) *AudioAnalysisDTO {
	dto := &AudioAnalysisDTO{
		InputType:     string(a.InputType),
		Transcript:    a.Transcript,
		LyricsSnippet: a.LyricsSnippet,
		LanguageCode:  a.LanguageCode,
		LanguageLabel: a.LanguageLabel,
	}
	if len(a.Emotions) > 0 {
		dto.Emotions = make([]EmotionSignalDTO, len(a.Emotions))
		for i, e := range a.Emotions {
			dto.Emotions[i] = EmotionSignalDTO{
				Name:      e.Name,
				Emoji:     e.Emoji,
				Intensity: e.Intensity,
			}
		}
	}
	return dto
}
