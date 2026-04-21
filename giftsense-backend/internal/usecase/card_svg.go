package usecase

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/giftsense/backend/internal/domain"
)

var svgTemplate = template.Must(template.New("card").Funcs(template.FuncMap{
	"wrapLines": func(text string, maxChars int) []string { return wrapText(text, maxChars) },
	"add":       func(a, b int) int { return a + b },
	"mul":       func(a, b int) int { return a * b },
}).Parse(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 420 595" width="420" height="595">
  <rect width="420" height="595" fill="{{.Palette.Background}}"/>
  <rect x="12" y="12" width="396" height="571" fill="none" stroke="{{.Palette.Primary}}" stroke-width="4" rx="4"/>
  {{if eq .Motif "birthday-confetti"}}
  <g opacity="0.22">
    <circle cx="390" cy="52" r="22" fill="{{.Palette.Accent}}" stroke="{{.Palette.Primary}}" stroke-width="1.5"/>
    <path d="M390 74 L388 82 L390 80 L392 82 Z" fill="{{.Palette.Primary}}"/>
    <path d="M390 82 Q384 96 388 108" stroke="{{.Palette.Primary}}" stroke-width="1.2" fill="none"/>
    <circle cx="355" cy="36" r="15" fill="{{.Palette.Primary}}"/>
    <path d="M355 51 L353 57 L355 56 L357 57 Z" fill="{{.Palette.Primary}}"/>
    <path d="M355 57 Q350 68 352 76" stroke="{{.Palette.Primary}}" stroke-width="1" fill="none"/>
    <circle cx="35" cy="25" r="5" fill="{{.Palette.Accent}}"/>
    <circle cx="85" cy="18" r="4" fill="{{.Palette.Primary}}"/>
    <circle cx="155" cy="24" r="5" fill="{{.Palette.Accent}}"/>
    <circle cx="225" cy="16" r="3" fill="{{.Palette.Primary}}"/>
    <circle cx="295" cy="22" r="5" fill="{{.Palette.Accent}}"/>
    <rect x="58" y="42" width="7" height="7" rx="1" fill="{{.Palette.Primary}}" transform="rotate(30 61 46)"/>
    <rect x="175" y="38" width="6" height="6" rx="1" fill="{{.Palette.Accent}}" transform="rotate(45 178 41)"/>
    <rect x="260" y="40" width="7" height="7" rx="1" fill="{{.Palette.Primary}}" transform="rotate(-20 263 43)"/>
    <path d="M115 30 L117 24 L119 30 L125 28 L120 33 L122 39 L117 35 L112 39 L114 33 L109 28 Z" fill="{{.Palette.Accent}}"/>
    <path d="M310 16 L311.5 11 L313 16 L318 14 L314 18 L315.5 23 L311.5 20 L307.5 23 L309 18 L305 14 Z" fill="{{.Palette.Primary}}"/>
  </g>
  {{else if eq .Motif "mothers-floral"}}
  <g opacity="0.20" fill="{{.Palette.Accent}}">
    <ellipse cx="400" cy="30" rx="10" ry="16" transform="rotate(-30 400 30)"/>
    <ellipse cx="395" cy="32" rx="10" ry="16" transform="rotate(10 395 32)"/>
    <ellipse cx="385" cy="25" rx="10" ry="16" transform="rotate(50 385 25)"/>
    <ellipse cx="380" cy="38" rx="10" ry="16" transform="rotate(-60 380 38)"/>
    <circle cx="390" cy="30" r="6" fill="{{.Palette.Primary}}"/>
    <path d="M370 50 C355 55 342 65 330 80" stroke="{{.Palette.Primary}}" stroke-width="2" fill="none"/>
    <ellipse cx="330" cy="82" rx="7" ry="11" transform="rotate(-20 330 82)"/>
    <ellipse cx="327" cy="84" rx="7" ry="11" transform="rotate(20 327 84)"/>
    <ellipse cx="322" cy="78" rx="7" ry="11" transform="rotate(55 322 78)"/>
    <circle cx="326" cy="80" r="4" fill="{{.Palette.Primary}}"/>
    <path d="M20 30 C35 35 45 48 40 60" stroke="{{.Palette.Accent}}" stroke-width="2" fill="none"/>
    <ellipse cx="22" cy="28" rx="7" ry="12" transform="rotate(25 22 28)"/>
    <ellipse cx="18" cy="32" rx="7" ry="12" transform="rotate(-30 18 32)"/>
    <circle cx="20" cy="28" r="4" fill="{{.Palette.Primary}}"/>
    <path d="M380 565 C370 555 355 552 340 558" stroke="{{.Palette.Accent}}" stroke-width="1.5" fill="none"/>
    <ellipse cx="382" cy="568" rx="6" ry="10" transform="rotate(15 382 568)"/>
    <ellipse cx="378" cy="571" rx="6" ry="10" transform="rotate(-25 378 571)"/>
    <circle cx="380" cy="566" r="3.5" fill="{{.Palette.Primary}}"/>
  </g>
  {{else if eq .Motif "anniversary-rings"}}
  <g opacity="0.18" stroke="{{.Palette.Primary}}" fill="none">
    <circle cx="195" cy="55" r="32" stroke-width="3"/>
    <circle cx="225" cy="55" r="32" stroke-width="3"/>
    <circle cx="195" cy="55" r="32" fill="{{.Palette.Accent}}" opacity="0.25" stroke="none"/>
    <circle cx="225" cy="55" r="32" fill="{{.Palette.Primary}}" opacity="0.12" stroke="none"/>
    <path d="M380 30 L382 23 L384 30 L391 27.5 L386 33 L388 40 L382 36.5 L376 40 L378 33 L373 27.5 Z" fill="{{.Palette.Accent}}" stroke="none"/>
    <path d="M35 40 L36.5 34.5 L38 40 L44 38 L40 42.5 L41.5 48 L36.5 45 L31.5 48 L33 42.5 L29 38 Z" fill="{{.Palette.Primary}}" stroke="none"/>
    <path d="M395 565 L396.5 559 L398 565 L404 562.5 L400 567 L401.5 573 L396.5 570 L391.5 573 L393 567 L389 562.5 Z" fill="{{.Palette.Accent}}" stroke="none"/>
    <circle cx="210" cy="25" r="3" fill="{{.Palette.Primary}}" stroke="none"/>
    <circle cx="220" cy="85" r="2.5" fill="{{.Palette.Accent}}" stroke="none"/>
    <circle cx="180" cy="20" r="2" fill="{{.Palette.Accent}}" stroke="none"/>
    <circle cx="242" cy="20" r="2" fill="{{.Palette.Primary}}" stroke="none"/>
  </g>
  {{else if eq .Motif "friendship-stars"}}
  <g opacity="0.20">
    <path d="M390 25 L393 15 L396 25 L406 22 L399 30 L402 40 L393 35 L384 40 L387 30 L380 22 Z" fill="{{.Palette.Primary}}"/>
    <path d="M35 35 L37 27 L39 35 L47 32.5 L41 39 L43 47 L37 43 L31 47 L33 39 L27 32.5 Z" fill="{{.Palette.Accent}}"/>
    <path d="M200 20 L202 13 L204 20 L211 17.5 L206 23 L208 30 L202 26.5 L196 30 L198 23 L193 17.5 Z" fill="{{.Palette.Primary}}"/>
    <path d="M370 570 L372 563 L374 570 L381 567.5 L376 573 L378 580 L372 576.5 L366 580 L368 573 L363 567.5 Z" fill="{{.Palette.Accent}}"/>
    <path d="M50 565 L51.5 559 L53 565 L59 562.5 L55 567 L56.5 573 L51.5 570 L46.5 573 L48 567 L44 562.5 Z" fill="{{.Palette.Primary}}"/>
    <circle cx="120" cy="22" r="4" fill="{{.Palette.Accent}}"/>
    <circle cx="300" cy="18" r="3" fill="{{.Palette.Primary}}"/>
    <circle cx="400" cy="100" r="3" fill="{{.Palette.Accent}}"/>
    <circle cx="20" cy="150" r="3" fill="{{.Palette.Primary}}"/>
    <path d="M14 580 Q100 560 210 570 Q320 580 408 560" stroke="{{.Palette.Primary}}" stroke-width="1.5" fill="none" opacity="0.4"/>
    <path d="M14 570 Q100 552 210 562 Q320 572 408 552" stroke="{{.Palette.Accent}}" stroke-width="1" fill="none" opacity="0.3"/>
  </g>
  {{else if eq .Motif "sunburst"}}
  <g opacity="0.15" stroke="{{.Palette.Primary}}" fill="none">
    <line x1="390" y1="20" x2="390" y2="50" stroke-width="2"/>
    <line x1="390" y1="20" x2="415" y2="38" stroke-width="2"/>
    <line x1="390" y1="20" x2="418" y2="20" stroke-width="2"/>
    <line x1="390" y1="20" x2="412" y2="5" stroke-width="2"/>
    <line x1="390" y1="20" x2="370" y2="5" stroke-width="2"/>
    <line x1="390" y1="20" x2="365" y2="38" stroke-width="2"/>
    <circle cx="390" cy="20" r="10" fill="{{.Palette.Accent}}" stroke="none"/>
    <circle cx="390" cy="20" r="6" fill="{{.Palette.Primary}}" stroke="none"/>
  </g>
  {{end}}
  <text x="210" y="175" text-anchor="middle" font-family="Georgia, serif" font-size="24" font-weight="bold" fill="{{.Palette.Primary}}">{{.Content.Headline}}</text>
  {{range $i, $line := wrapLines .Content.Body 50}}
  <text x="210" y="{{add 215 (mul $i 26)}}" text-anchor="middle" font-family="Georgia, serif" font-size="15" fill="{{$.Palette.Ink}}">{{$line}}</text>
  {{end}}
  <text x="210" y="430" text-anchor="middle" font-family="Georgia, serif" font-size="14" fill="{{.Palette.Muted}}">{{.Content.Closing}}</text>
  <text x="210" y="452" text-anchor="middle" font-family="Georgia, serif" font-size="13" font-style="italic" fill="{{.Palette.Muted}}">{{.Content.Signature}}</text>
  <text x="210" y="555" text-anchor="middle" font-family="Georgia, serif" font-size="14" fill="{{.Palette.Accent}}">For {{.Content.Recipient}}</text>
</svg>`))

type svgData struct {
	Motif   string
	Palette domain.CardPalette
	Content domain.CardContent
}

func RenderSVG(motif string, palette domain.CardPalette, content domain.CardContent) (string, error) {
	var buf bytes.Buffer
	if err := svgTemplate.Execute(&buf, svgData{Motif: motif, Palette: palette, Content: content}); err != nil {
		return "", fmt.Errorf("%w: %s", domain.ErrCardRenderFailed, err.Error())
	}
	return buf.String(), nil
}

func wrapText(text string, maxChars int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]

	for _, word := range words[1:] {
		if len(current)+1+len(word) <= maxChars {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	lines = append(lines, current)
	return lines
}
