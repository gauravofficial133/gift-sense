package usecase

import (
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

var occasionTemplates = map[domain.OccasionKey]domain.OccasionTemplate{
	"birthday":    {Key: "birthday", Greeting: "Happy Birthday, %s!", Motif: "birthday-confetti"},
	"mothers_day": {Key: "mothers_day", Greeting: "Happy Mother's Day, %s!", Motif: "mothers-floral"},
	"anniversary": {Key: "anniversary", Greeting: "Happy Anniversary, %s!", Motif: "anniversary-rings"},
	"friendship":  {Key: "friendship", Greeting: "For You, %s!", Motif: "friendship-stars"},
	"default":     {Key: "default", Greeting: "For %s,", Motif: "sunburst"},
}

var emotionGroupPalettes = map[domain.EmotionGroup]domain.CardPalette{
	"joyful":     {Background: "#FFFDE7", Primary: "#D97706", Accent: "#FCD34D", Ink: "#1C1917", Muted: "#78716C"},
	"tender":     {Background: "#FFF0F5", Primary: "#DB2777", Accent: "#F9A8D4", Ink: "#1C1917", Muted: "#9D174D"},
	"warm":       {Background: "#FFF7ED", Primary: "#C2410C", Accent: "#FB923C", Ink: "#1C1917", Muted: "#7C2D12"},
	"passionate": {Background: "#FFF1F2", Primary: "#BE123C", Accent: "#FB7185", Ink: "#1C1917", Muted: "#9F1239"},
	"reflective": {Background: "#F0F4FF", Primary: "#3B4F9E", Accent: "#93A8D4", Ink: "#1C1917", Muted: "#334155"},
	"neutral":    {Background: "#FFF9F0", Primary: "#D97706", Accent: "#FDE68A", Ink: "#1C1917", Muted: "#92400E"},
}

var emotionToGroup = map[string]domain.EmotionGroup{
	"Joy": "joyful", "Playfulness": "joyful", "Celebration": "joyful",
	"Tenderness": "tender", "Deep Love": "tender", "Devotion": "tender",
	"Warmth": "warm", "Hope": "warm", "Nostalgia": "warm",
	"Passion": "passionate", "Longing": "passionate", "Yearning": "passionate",
	"Melancholy": "reflective", "Grief": "reflective", "Heartbreak": "reflective",
}

func DetectOccasion(occasionStr string) domain.OccasionKey {
	s := strings.ToLower(strings.TrimSpace(occasionStr))
	switch {
	case strings.Contains(s, "birthday"):
		return "birthday"
	case strings.Contains(s, "mother"):
		return "mothers_day"
	case strings.Contains(s, "anniversary"):
		return "anniversary"
	case strings.Contains(s, "friendship") || strings.Contains(s, "thank"):
		return "friendship"
	default:
		return "default"
	}
}

func DetectEmotionGroup(emotions []domain.EmotionSignal) domain.EmotionGroup {
	if len(emotions) == 0 {
		return "neutral"
	}
	strongest := emotions[0]
	for _, e := range emotions[1:] {
		if e.Intensity > strongest.Intensity {
			strongest = e
		}
	}
	if group, ok := emotionToGroup[strongest.Name]; ok {
		return group
	}
	return "neutral"
}

func GetOccasionTemplate(key domain.OccasionKey) domain.OccasionTemplate {
	if t, ok := occasionTemplates[key]; ok {
		return t
	}
	return occasionTemplates["default"]
}

func GetEmotionPalette(group domain.EmotionGroup) domain.CardPalette {
	if p, ok := emotionGroupPalettes[group]; ok {
		return p
	}
	return emotionGroupPalettes["neutral"]
}
