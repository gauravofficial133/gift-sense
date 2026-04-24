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
	"joyful":     {Background: "#FFFDE7", BackgroundSecondary: "#FFF8E1", Primary: "#D97706", Accent: "#FCD34D", Ink: "#1C1917", Muted: "#78716C", Overlay: "rgba(252,211,77,0.1)"},
	"tender":     {Background: "#FFF0F5", BackgroundSecondary: "#FFE4EF", Primary: "#DB2777", Accent: "#F9A8D4", Ink: "#1C1917", Muted: "#9D174D", Overlay: "rgba(249,168,212,0.1)"},
	"warm":       {Background: "#FFF7ED", BackgroundSecondary: "#FFF1E0", Primary: "#C2410C", Accent: "#FB923C", Ink: "#1C1917", Muted: "#7C2D12", Overlay: "rgba(251,146,60,0.1)"},
	"passionate": {Background: "#FFF1F2", BackgroundSecondary: "#FFE4E6", Primary: "#BE123C", Accent: "#FB7185", Ink: "#1C1917", Muted: "#9F1239", Overlay: "rgba(251,113,133,0.1)"},
	"reflective": {Background: "#F0F4FF", BackgroundSecondary: "#E8EEFF", Primary: "#3B4F9E", Accent: "#93A8D4", Ink: "#1C1917", Muted: "#334155", Overlay: "rgba(147,168,212,0.1)"},
	"neutral":    {Background: "#FFF9F0", BackgroundSecondary: "#FFF5E6", Primary: "#D97706", Accent: "#FDE68A", Ink: "#1C1917", Muted: "#92400E", Overlay: "rgba(253,230,138,0.1)"},
}

type PaletteEntry struct {
	Name    string
	Palette domain.CardPalette
	Emotions []domain.EmotionGroup
}

var namedPalettes = []PaletteEntry{
	{Name: "sunrise_warmth", Emotions: []domain.EmotionGroup{"warm"}, Palette: domain.CardPalette{
		Background: "#FFF5E6", BackgroundSecondary: "#FFEDD5", Primary: "#D4451A", Accent: "#FFB347", Ink: "#5C3D2E", Muted: "#78716C", Overlay: "rgba(255,180,71,0.1)",
	}},
	{Name: "soft_rose_gold", Emotions: []domain.EmotionGroup{"tender", "passionate"}, Palette: domain.CardPalette{
		Background: "#FFF0F0", BackgroundSecondary: "#FFE8E8", Primary: "#8B3A62", Accent: "#E8A0BF", Ink: "#5C4A4A", Muted: "#9C8B8B", Overlay: "rgba(232,160,191,0.1)",
	}},
	{Name: "ocean_calm", Emotions: []domain.EmotionGroup{"reflective"}, Palette: domain.CardPalette{
		Background: "#EBF5FB", BackgroundSecondary: "#D6EAF8", Primary: "#1A5276", Accent: "#5DADE2", Ink: "#2C3E50", Muted: "#5D6D7E", Overlay: "rgba(93,173,226,0.1)",
	}},
	{Name: "electric_joy", Emotions: []domain.EmotionGroup{"joyful"}, Palette: domain.CardPalette{
		Background: "#FFFDE7", BackgroundSecondary: "#FFF9C4", Primary: "#E65100", Accent: "#FFD600", Ink: "#3E2723", Muted: "#795548", Overlay: "rgba(255,214,0,0.15)",
	}},
	{Name: "midnight_elegant", Emotions: []domain.EmotionGroup{"reflective", "neutral"}, Palette: domain.CardPalette{
		Background: "#1A1A2E", BackgroundSecondary: "#162447", Primary: "#E8D5B7", Accent: "#B8860B", Ink: "#F5F0E8", Muted: "#C4B59D", Overlay: "rgba(184,134,11,0.1)",
	}},
	{Name: "forest_peace", Emotions: []domain.EmotionGroup{"warm", "reflective"}, Palette: domain.CardPalette{
		Background: "#F1F8E9", BackgroundSecondary: "#DCEDC8", Primary: "#2E7D32", Accent: "#81C784", Ink: "#3E4A3E", Muted: "#689F38", Overlay: "rgba(129,199,132,0.1)",
	}},
	{Name: "lavender_dream", Emotions: []domain.EmotionGroup{"tender"}, Palette: domain.CardPalette{
		Background: "#F3E5F5", BackgroundSecondary: "#E1BEE7", Primary: "#6A1B9A", Accent: "#CE93D8", Ink: "#4A3A5C", Muted: "#8E6DAF", Overlay: "rgba(206,147,216,0.1)",
	}},
	{Name: "golden_celebration", Emotions: []domain.EmotionGroup{"joyful", "warm"}, Palette: domain.CardPalette{
		Background: "#FFFAF0", BackgroundSecondary: "#FFF3E0", Primary: "#BF360C", Accent: "#FFB300", Ink: "#4E342E", Muted: "#8D6E63", Overlay: "rgba(255,179,0,0.15)",
	}},
}

func GetNamedPalette(name string) (domain.CardPalette, bool) {
	for _, p := range namedPalettes {
		if p.Name == name {
			return p.Palette, true
		}
	}
	return domain.CardPalette{}, false
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

func ListPaletteNames() []string {
	names := make([]string, len(namedPalettes))
	for i, p := range namedPalettes {
		names[i] = p.Name
	}
	return names
}
