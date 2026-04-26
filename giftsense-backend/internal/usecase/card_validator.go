package usecase

import (
	"fmt"
	"math"
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

func ValidateCard(render *domain.CardRender, tplDef domain.TemplateDefinition, palette domain.CardPalette) domain.ValidationResult {
	var issues []string

	overflow := checkTextOverflow(render.Content, tplDef.Elements)
	if overflow {
		issues = append(issues, "text exceeds zone character limits")
	}

	ratio, contrastOK := checkContrastRatio(palette)
	if !contrastOK {
		issues = append(issues, fmt.Sprintf("contrast ratio %.2f below WCAG AA (4.5)", ratio))
	}

	illustrationOK := checkIllustration(render)
	if render.Illustration != nil && !illustrationOK {
		issues = append(issues, "illustration requested but missing or empty")
	}

	score := computeCompositionScore(overflow, contrastOK, illustrationOK, render.Illustration != nil)

	return domain.ValidationResult{
		TextOverflow:     overflow,
		ContrastRatio:    math.Round(ratio*100) / 100,
		ContrastPassed:   contrastOK,
		IllustrationOK:   illustrationOK,
		CompositionScore: math.Round(score*100) / 100,
		OverallPass:      !overflow && contrastOK && illustrationOK,
		Issues:           issues,
	}
}

func checkTextOverflow(content domain.CardContent, elements []domain.Element) bool {
	for _, el := range elements {
		if el.TextZone == nil {
			continue
		}
		text := textForRole(content, el.TextZone.SemanticRole)
		if el.TextZone.CharMax > 0 && len(text) > el.TextZone.CharMax {
			return true
		}
	}
	return false
}

func textForRole(content domain.CardContent, role string) string {
	switch strings.ToLower(role) {
	case "headline":
		return content.Headline
	case "body":
		return content.Body
	case "closing":
		return content.Closing
	case "signature":
		return content.Signature
	default:
		return ""
	}
}

func checkContrastRatio(palette domain.CardPalette) (float64, bool) {
	bgLum := relativeLuminance(palette.Background)
	inkLum := relativeLuminance(palette.Ink)
	lighter := math.Max(bgLum, inkLum)
	darker := math.Min(bgLum, inkLum)
	ratio := (lighter + 0.05) / (darker + 0.05)
	return ratio, ratio >= 4.5
}

func relativeLuminance(hex string) float64 {
	r, g, b := parseHexColor(hex)
	return 0.2126*linearize(r) + 0.7152*linearize(g) + 0.0722*linearize(b)
}

func linearize(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func parseHexColor(hex string) (float64, float64, float64) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) < 6 {
		return 0, 0, 0
	}
	r := hexToDec(hex[0:2])
	g := hexToDec(hex[2:4])
	b := hexToDec(hex[4:6])
	return float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0
}

func hexToDec(s string) int {
	val := 0
	for _, c := range s {
		val *= 16
		switch {
		case c >= '0' && c <= '9':
			val += int(c - '0')
		case c >= 'a' && c <= 'f':
			val += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			val += int(c-'A') + 10
		}
	}
	return val
}

func checkIllustration(render *domain.CardRender) bool {
	if render.Illustration == nil {
		return true
	}
	return render.Illustration.PNGBase64 != ""
}

func computeCompositionScore(overflow, contrastOK, illustrationOK, hasIllustration bool) float64 {
	score := 0.0
	weights := 0.0

	weights += 0.35
	if !overflow {
		score += 0.35
	}

	weights += 0.35
	if contrastOK {
		score += 0.35
	}

	if hasIllustration {
		weights += 0.30
		if illustrationOK {
			score += 0.30
		}
	} else {
		score += 0.0
		weights += 0.0
	}

	if weights == 0 {
		return 1.0
	}
	return score / weights
}
