package usecase

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

var occasionSampleContent = map[domain.OccasionKey]domain.CardContent{
	"birthday": {
		Headline:  "Happy Birthday, Alex!",
		Body:      "Another year of your incredible journey. You bring warmth and laughter to everyone around you.",
		Closing:   "With love",
		Signature: "Always yours",
		Recipient: "Alex",
	},
	"mothers_day": {
		Headline:  "Happy Mother's Day!",
		Body:      "Thank you for filling our home with warmth, wisdom, and unconditional love every single day.",
		Closing:   "With all my love",
		Signature: "Your grateful child",
		Recipient: "Mom",
	},
	"anniversary": {
		Headline:  "Happy Anniversary!",
		Body:      "Every moment with you is a chapter in the most beautiful story ever written. Here's to forever.",
		Closing:   "Eternally yours",
		Signature: "With all my heart",
		Recipient: "Love",
	},
	"friendship": {
		Headline:  "For You, Dear Friend!",
		Body:      "Some people make the world brighter just by being in it. You are one of those rare, wonderful people.",
		Closing:   "Always grateful",
		Signature: "Your friend",
		Recipient: "Friend",
	},
	"default": {
		Headline:  "Thinking of You",
		Body:      "Just a note to say how much you mean to everyone around you. You make the ordinary extraordinary.",
		Closing:   "With warmth",
		Signature: "Yours truly",
		Recipient: "You",
	},
}

var memorySampleContent = domain.CardContent{
	Headline:  "Our Favorite Moments",
	Body:      "\"Remember that rainy afternoon we spent laughing?\" ... \"You always know exactly what to say\" ... \"That sunset walk was everything\"",
	Closing:   "Cherishing every moment",
	Signature: "With love",
	Recipient: "You",
}

func PickPreviewPalette(tpl domain.TemplateDefinition) domain.CardPalette {
	if tpl.VariationRules.PaletteMode == "set" && len(tpl.VariationRules.AllowedPalettes) > 0 {
		name := tpl.VariationRules.AllowedPalettes[0]
		if p, ok := GetNamedPalette(name); ok {
			return p
		}
	}

	if tpl.VariationRules.PaletteMood != "" {
		mood := strings.ToLower(tpl.VariationRules.PaletteMood)
		for _, entry := range namedPalettes {
			desc := strings.ToLower(paletteDescriptions[entry.Name])
			if strings.Contains(desc, mood) || strings.Contains(mood, string(entry.Emotions[0])) {
				return entry.Palette
			}
		}
	}

	if len(tpl.Emotions) > 0 {
		for _, entry := range namedPalettes {
			for _, pe := range entry.Emotions {
				for _, te := range tpl.Emotions {
					if string(pe) == string(te) {
						return entry.Palette
					}
				}
			}
		}
	}

	return namedPalettes[0].Palette
}

func BuildPreviewContent(tpl domain.TemplateDefinition) domain.CardContent {
	if tpl.Family == "memory" {
		return memorySampleContent
	}

	occasion := pickBestOccasion(tpl.Occasions)
	base, ok := occasionSampleContent[occasion]
	if !ok {
		base = occasionSampleContent["default"]
	}

	for _, el := range tpl.Elements {
		if el.TextZone == nil {
			continue
		}
		tz := el.TextZone
		switch tz.SemanticRole {
		case "headline":
			if tz.CharMax > 0 && len(base.Headline) > tz.CharMax {
				base.Headline = base.Headline[:tz.CharMax]
			}
		case "body":
			if tz.CharMax > 0 && len(base.Body) > tz.CharMax {
				base.Body = base.Body[:tz.CharMax]
			}
			if tz.Purpose != "" && strings.Contains(strings.ToLower(tz.Purpose), "quote") {
				base.Body = fmt.Sprintf("\"%s\"", base.Body)
			}
		case "closing":
			if tz.CharMax > 0 && len(base.Closing) > tz.CharMax {
				base.Closing = base.Closing[:tz.CharMax]
			}
		case "signature":
			if tz.CharMax > 0 && len(base.Signature) > tz.CharMax {
				base.Signature = base.Signature[:tz.CharMax]
			}
		}
	}

	base.OccasionKey = string(occasion)
	return base
}

func PickPreviewFonts(tpl domain.TemplateDefinition) map[string]string {
	choices := make(map[string]string)
	for _, el := range tpl.Elements {
		if el.TextZone != nil && len(el.TextZone.FontOptions) > 0 {
			choices[el.ID] = el.TextZone.FontOptions[0]
		}
	}
	return choices
}

func pickBestOccasion(occasions []domain.OccasionKey) domain.OccasionKey {
	priority := []domain.OccasionKey{"birthday", "mothers_day", "anniversary", "friendship"}
	for _, p := range priority {
		for _, o := range occasions {
			if o == p {
				return o
			}
		}
	}
	if len(occasions) > 0 {
		return occasions[0]
	}
	return "default"
}

var contentVariants = map[domain.OccasionKey][]domain.CardContent{
	"birthday": {
		{Headline: "Happy Birthday, Alex!", Body: "Another year of your incredible journey. You bring warmth and laughter to everyone around you.", Closing: "With love", Signature: "Always yours", Recipient: "Alex"},
		{Headline: "Cheers to You, Alex!", Body: "Today we celebrate the amazing person you are. Your kindness lights up every room you enter.", Closing: "Warmly", Signature: "Your biggest fan", Recipient: "Alex"},
		{Headline: "It's Your Day, Alex!", Body: "May this year bring you everything your heart desires. You deserve all the happiness in the world.", Closing: "With joy", Signature: "Yours forever", Recipient: "Alex"},
	},
	"mothers_day": {
		{Headline: "Happy Mother's Day!", Body: "Thank you for filling our home with warmth, wisdom, and unconditional love every single day.", Closing: "With all my love", Signature: "Your grateful child", Recipient: "Mom"},
		{Headline: "For the Best Mom!", Body: "Your strength, grace, and endless patience inspire me every day. Thank you for being you.", Closing: "Forever grateful", Signature: "With admiration", Recipient: "Mom"},
	},
	"anniversary": {
		{Headline: "Happy Anniversary!", Body: "Every moment with you is a chapter in the most beautiful story ever written. Here's to forever.", Closing: "Eternally yours", Signature: "With all my heart", Recipient: "Love"},
		{Headline: "To Us, My Love!", Body: "Through every season, every challenge, every joy — you are my favorite person to share it all with.", Closing: "Always together", Signature: "Yours completely", Recipient: "Love"},
	},
	"friendship": {
		{Headline: "For You, Dear Friend!", Body: "Some people make the world brighter just by being in it. You are one of those rare, wonderful people.", Closing: "Always grateful", Signature: "Your friend", Recipient: "Friend"},
		{Headline: "Thank You, Friend!", Body: "For the laughs, the late-night talks, and the unspoken understanding — you make life so much richer.", Closing: "Cheers to us", Signature: "With appreciation", Recipient: "Friend"},
	},
	"default": {
		{Headline: "Thinking of You", Body: "Just a note to say how much you mean to everyone around you. You make the ordinary extraordinary.", Closing: "With warmth", Signature: "Yours truly", Recipient: "You"},
		{Headline: "A Note for You", Body: "Sometimes the simplest words carry the deepest meaning. You are appreciated more than you know.", Closing: "Fondly", Signature: "With care", Recipient: "You"},
	},
}

var memoryVariants = []domain.CardContent{
	{Headline: "Our Favorite Moments", Body: "\"Remember that rainy afternoon we spent laughing?\" ... \"You always know exactly what to say\" ... \"That sunset walk was everything\"", Closing: "Cherishing every moment", Signature: "With love", Recipient: "You"},
	{Headline: "Memories We Treasure", Body: "\"The way you smiled that morning\" ... \"Our kitchen dance at midnight\" ... \"You made it all worthwhile\"", Closing: "Forever in my heart", Signature: "With gratitude", Recipient: "You"},
}

func PickPreviewPaletteVaried(tpl domain.TemplateDefinition, seed int64) domain.CardPalette {
	rng := rand.New(rand.NewSource(seed))

	if tpl.VariationRules.PaletteMode == "set" && len(tpl.VariationRules.AllowedPalettes) > 0 {
		name := tpl.VariationRules.AllowedPalettes[rng.Intn(len(tpl.VariationRules.AllowedPalettes))]
		if p, ok := GetNamedPalette(name); ok {
			return p
		}
	}

	if len(tpl.Emotions) > 0 {
		var matching []domain.CardPalette
		for _, entry := range namedPalettes {
			for _, pe := range entry.Emotions {
				for _, te := range tpl.Emotions {
					if string(pe) == string(te) {
						matching = append(matching, entry.Palette)
					}
				}
			}
		}
		if len(matching) > 0 {
			return matching[rng.Intn(len(matching))]
		}
	}

	return namedPalettes[rng.Intn(len(namedPalettes))].Palette
}

func BuildPreviewContentVaried(tpl domain.TemplateDefinition, seed int64) domain.CardContent {
	rng := rand.New(rand.NewSource(seed))

	if tpl.Family == "memory" {
		return memoryVariants[rng.Intn(len(memoryVariants))]
	}

	occasion := pickBestOccasion(tpl.Occasions)
	variants, ok := contentVariants[occasion]
	if !ok {
		variants = contentVariants["default"]
	}
	base := variants[rng.Intn(len(variants))]

	for _, el := range tpl.Elements {
		if el.TextZone == nil {
			continue
		}
		tz := el.TextZone
		switch tz.SemanticRole {
		case "headline":
			if tz.CharMax > 0 && len(base.Headline) > tz.CharMax {
				base.Headline = base.Headline[:tz.CharMax]
			}
		case "body":
			if tz.CharMax > 0 && len(base.Body) > tz.CharMax {
				base.Body = base.Body[:tz.CharMax]
			}
			if tz.Purpose != "" && strings.Contains(strings.ToLower(tz.Purpose), "quote") {
				base.Body = fmt.Sprintf("\"%s\"", base.Body)
			}
		case "closing":
			if tz.CharMax > 0 && len(base.Closing) > tz.CharMax {
				base.Closing = base.Closing[:tz.CharMax]
			}
		case "signature":
			if tz.CharMax > 0 && len(base.Signature) > tz.CharMax {
				base.Signature = base.Signature[:tz.CharMax]
			}
		}
	}

	base.OccasionKey = string(occasion)
	return base
}

func PickPreviewFontsVaried(tpl domain.TemplateDefinition, seed int64) map[string]string {
	rng := rand.New(rand.NewSource(seed))
	choices := make(map[string]string)
	for _, el := range tpl.Elements {
		if el.TextZone != nil && len(el.TextZone.FontOptions) > 0 {
			choices[el.ID] = el.TextZone.FontOptions[rng.Intn(len(el.TextZone.FontOptions))]
		}
	}
	return choices
}
