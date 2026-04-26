package usecase

import (
	"math"
	"sort"
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

func ScoreCards(cards []*domain.CardRender, occasion string, emotions []domain.EmotionSignal, templates []domain.TemplateDefinition) []*domain.CardRender {
	tplByID := make(map[string]domain.TemplateDefinition, len(templates))
	for _, t := range templates {
		tplByID[t.ID] = t
	}
	for _, card := range cards {
		tpl := tplByID[card.RecipeID]
		scoring := scoreOneCard(card, occasion, emotions, tpl, cards)
		if card.Meta == nil {
			card.Meta = &domain.CardMeta{}
		}
		card.Meta.Scoring = &scoring
	}
	sort.Slice(cards, func(i, j int) bool {
		return cardScore(cards[i]) > cardScore(cards[j])
	})
	return cards
}

func cardScore(card *domain.CardRender) float64 {
	if card.Meta == nil || card.Meta.Scoring == nil {
		return 0
	}
	return card.Meta.Scoring.TotalScore
}

func scoreOneCard(card *domain.CardRender, occasion string, emotions []domain.EmotionSignal, tpl domain.TemplateDefinition, allCards []*domain.CardRender) domain.ScoringBreakdown {
	fit := scoreTemplateOccasionFit(occasion, tpl)
	copy := scoreCopyQuality(card)
	harmony := scoreVisualHarmony(card, emotions, tpl)
	originality := scoreOriginality(card, allCards)

	total := fit*0.25 + copy*0.30 + harmony*0.20 + originality*0.25

	return domain.ScoringBreakdown{
		TemplateOccasionFit: round2(fit),
		CopyQuality:         round2(copy),
		VisualHarmony:       round2(harmony),
		Originality:         round2(originality),
		TotalScore:          round2(total),
	}
}

func scoreTemplateOccasionFit(occasion string, tpl domain.TemplateDefinition) float64 {
	if len(tpl.Occasions) == 0 {
		return 0.5
	}
	occ := strings.ToLower(occasion)
	for _, o := range tpl.Occasions {
		if strings.Contains(occ, strings.ToLower(string(o))) {
			return 1.0
		}
	}
	for _, o := range tpl.Occasions {
		if string(o) == "default" || string(o) == "general" {
			return 0.6
		}
	}
	return 0.3
}

func scoreCopyQuality(card *domain.CardRender) float64 {
	score := 0.0
	c := card.Content

	if len(c.Headline) >= 5 && len(c.Headline) <= 40 {
		score += 0.25
	} else if len(c.Headline) > 0 {
		score += 0.10
	}

	if len(c.Body) >= 50 && len(c.Body) <= 240 {
		score += 0.35
	} else if len(c.Body) >= 20 {
		score += 0.15
	}

	if len(c.Closing) >= 2 && len(c.Closing) <= 30 {
		score += 0.20
	}

	if len(c.Signature) >= 1 && len(c.Signature) <= 30 {
		score += 0.20
	}

	return score
}

func scoreVisualHarmony(card *domain.CardRender, emotions []domain.EmotionSignal, tpl domain.TemplateDefinition) float64 {
	score := 0.5

	if card.Meta != nil && card.Meta.Validation != nil {
		if card.Meta.Validation.ContrastPassed {
			score += 0.25
		}
		if !card.Meta.Validation.TextOverflow {
			score += 0.15
		}
		if card.Meta.Validation.IllustrationOK {
			score += 0.10
		}
	} else {
		score += 0.25
	}

	return math.Min(score, 1.0)
}

func scoreOriginality(card *domain.CardRender, allCards []*domain.CardRender) float64 {
	if len(allCards) <= 1 {
		return 1.0
	}

	sameTemplate := 0
	samePalette := 0
	sameFamily := 0

	for _, other := range allCards {
		if other == card {
			continue
		}
		if other.RecipeID == card.RecipeID {
			sameTemplate++
		}
		if other.PaletteName == card.PaletteName {
			samePalette++
		}
		if card.Meta != nil && other.Meta != nil && card.Meta.TemplateFamily != "" {
			if card.Meta.TemplateFamily == other.Meta.TemplateFamily {
				sameFamily++
			}
		}
	}

	score := 1.0
	if sameTemplate > 0 {
		score -= 0.4
	}
	if samePalette > 0 {
		score -= 0.2
	}
	if sameFamily > 0 {
		score -= 0.1
	}

	return math.Max(score, 0.0)
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}
