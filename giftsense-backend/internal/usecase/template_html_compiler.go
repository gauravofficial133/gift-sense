package usecase

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/giftsense/backend/internal/adapter/cardrender"
	"github.com/giftsense/backend/internal/domain"
)

type HTMLCompiler struct {
	fonts    map[string]cardrender.FontData
	assetLib *AssetLibrary
}

type CompileInput struct {
	Template      domain.TemplateDefinition
	Palette       domain.CardPalette
	Content       domain.CardContent
	Illustrations map[string]string
	DataFields    map[string]string
	Photos        map[string]string
	FontChoices   map[string]string
	BackgroundIdx int
	Seed          int64
}

func NewHTMLCompiler(fonts map[string]cardrender.FontData, assetLib *AssetLibrary) *HTMLCompiler {
	return &HTMLCompiler{fonts: fonts, assetLib: assetLib}
}

func (c *HTMLCompiler) Compile(input CompileInput) (string, error) {
	tpl := input.Template
	rng := rand.New(rand.NewSource(input.Seed))

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><style>`)

	c.writeFontFaces(&sb)
	c.writePaletteVars(&sb, input.Palette)
	c.writeBaseStyles(&sb, tpl.Canvas)

	elements := make([]domain.Element, len(tpl.Elements))
	copy(elements, tpl.Elements)
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].ZIndex < elements[j].ZIndex
	})

	for _, el := range elements {
		c.writeElementStyle(&sb, el, tpl.VariationRules.LayoutJitter, rng, input.FontChoices)
	}

	sb.WriteString(`</style></head><body>`)

	for _, el := range elements {
		if err := c.writeElement(&sb, el, input); err != nil {
			return "", fmt.Errorf("render element %s: %w", el.ID, err)
		}
	}

	sb.WriteString(`</body></html>`)
	return sb.String(), nil
}

func (c *HTMLCompiler) CompileInsidePage(input CompileInput) (string, error) {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><style>`)
	c.writeFontFaces(&sb)
	c.writePaletteVars(&sb, input.Palette)

	fmt.Fprintf(&sb, `* { margin: 0; padding: 0; box-sizing: border-box; }
html, body { width: %dpx; height: %dpx; }
body {
  background: var(--bg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 15%%;
  text-align: center;
}
.inside-body {
  font-family: 'Cormorant Garamond', serif;
  font-size: 16px;
  line-height: 1.8;
  color: var(--ink);
  margin-bottom: 32px;
}
.inside-closing {
  font-family: 'Cormorant Garamond', serif;
  font-size: 14px;
  color: var(--muted);
  margin-bottom: 12px;
}
.inside-signature {
  font-family: 'Great Vibes', cursive;
  font-size: 28px;
  color: var(--primary);
}
.inside-divider {
  width: 40px;
  height: 2px;
  background: var(--accent);
  margin: 20px auto;
  border-radius: 1px;
}`, input.Template.Canvas.Width, input.Template.Canvas.Height)

	sb.WriteString(`</style></head><body>`)
	fmt.Fprintf(&sb, `<div class="inside-body">%s</div>`, escapeHTML(input.Content.Body))
	sb.WriteString(`<div class="inside-divider"></div>`)
	fmt.Fprintf(&sb, `<div class="inside-closing">%s</div>`, escapeHTML(input.Content.Closing))
	fmt.Fprintf(&sb, `<div class="inside-signature">%s</div>`, escapeHTML(input.Content.Signature))
	sb.WriteString(`</body></html>`)

	return sb.String(), nil
}

func (c *HTMLCompiler) writeFontFaces(sb *strings.Builder) {
	for _, f := range c.fonts {
		fmt.Fprintf(sb, `@font-face {
  font-family: '%s';
  font-weight: %s;
  font-style: %s;
  font-display: block;
  src: url(data:font/woff2;base64,%s) format('woff2');
}
`, f.Family, f.Weight, f.Style, f.Base64)
	}
}

func (c *HTMLCompiler) writePaletteVars(sb *strings.Builder, p domain.CardPalette) {
	fmt.Fprintf(sb, `:root {
  --bg: %s;
  --bg2: %s;
  --primary: %s;
  --accent: %s;
  --ink: %s;
  --muted: %s;
  --overlay: %s;
}
`, p.Background, p.BackgroundSecondary, p.Primary, p.Accent, p.Ink, p.Muted, p.Overlay)
}

func (c *HTMLCompiler) writeBaseStyles(sb *strings.Builder, canvas domain.CanvasSpec) {
	fmt.Fprintf(sb, `* { margin: 0; padding: 0; box-sizing: border-box; }
html, body { width: %dpx; height: %dpx; overflow: hidden; }
body { background: var(--bg); position: relative; }
`, canvas.Width, canvas.Height)
}

func (c *HTMLCompiler) writeElementStyle(sb *strings.Builder, el domain.Element, jitter domain.LayoutJitter, rng *rand.Rand, fontChoices map[string]string) {
	if el.Position == nil || el.Size == nil {
		if el.Type == "background" {
			fmt.Fprintf(sb, `#%s { position: absolute; inset: 0; z-index: %d; }
`, el.ID, el.ZIndex)
		}
		return
	}

	x := el.Position.X
	y := el.Position.Y
	w := el.Size.W
	h := el.Size.H

	if jitter.PositionRangePx > 0 {
		x += rng.Intn(jitter.PositionRangePx*2+1) - jitter.PositionRangePx
		y += rng.Intn(jitter.PositionRangePx*2+1) - jitter.PositionRangePx
	}
	if jitter.SizeRangePct > 0 {
		pct := 1.0 + float64(rng.Intn(jitter.SizeRangePct*2+1)-jitter.SizeRangePct)/100.0
		w = int(float64(w) * pct)
		h = int(float64(h) * pct)
	}

	fmt.Fprintf(sb, `#%s { position: absolute; left: %dpx; top: %dpx; width: %dpx; height: %dpx; z-index: %d;`,
		el.ID, x, y, w, h, el.ZIndex)

	if el.Rotation != 0 {
		fmt.Fprintf(sb, ` transform: rotate(%.1fdeg);`, el.Rotation)
	}

	if el.TextZone != nil {
		font := el.TextZone.FontOptions[0]
		if chosen, ok := fontChoices[el.ID]; ok {
			font = chosen
		}
		fmt.Fprintf(sb, ` font-family: '%s', serif; font-weight: %s; text-align: %s; color: %s;`,
			font, el.TextZone.FontWeight, el.TextZone.Alignment, resolveColor(el.TextZone.ColorSource))
		fontSize := (el.TextZone.FontSizeRange.Min + el.TextZone.FontSizeRange.Max) / 2
		fmt.Fprintf(sb, ` font-size: %dpx; overflow: hidden; display: flex; align-items: center; justify-content: center;`, fontSize)
	}

	if el.DataSlot != nil {
		font := "sans-serif"
		if len(el.DataSlot.FontOptions) > 0 {
			font = "'" + el.DataSlot.FontOptions[0] + "', serif"
		}
		fmt.Fprintf(sb, ` font-family: %s; font-size: %dpx; text-align: %s; color: %s; display: flex; align-items: center; justify-content: center;`,
			font, el.DataSlot.FontSize, el.DataSlot.Alignment, resolveColor(el.DataSlot.ColorSource))
	}

	sb.WriteString(" }\n")
}

func (c *HTMLCompiler) writeElement(sb *strings.Builder, el domain.Element, input CompileInput) error {
	switch el.Type {
	case "background":
		c.writeBackground(sb, el, input)
	case "text_zone":
		c.writeTextZone(sb, el, input.Content)
	case "photo_slot":
		c.writePhotoSlot(sb, el, input.Photos)
	case "data_slot":
		c.writeDataSlot(sb, el, input.DataFields)
	case "decorative":
		c.writeDecorative(sb, el)
	case "ai_illustration_slot":
		c.writeIllustration(sb, el, input.Illustrations)
	}
	return nil
}

func (c *HTMLCompiler) writeBackground(sb *strings.Builder, el domain.Element, input CompileInput) {
	tpl := input.Template
	bgIdx := input.BackgroundIdx
	if bgIdx < 0 || bgIdx >= len(tpl.VariationRules.BackgroundOptions) {
		bgIdx = 0
	}

	fmt.Fprintf(sb, `<div id="%s" style="`, el.ID)

	if len(tpl.VariationRules.BackgroundOptions) == 0 {
		sb.WriteString(`background: var(--bg);"></div>`)
		return
	}

	opt := tpl.VariationRules.BackgroundOptions[bgIdx]
	switch opt.Type {
	case "gradient":
		dir := opt.Direction
		if dir == "" {
			dir = "to bottom"
		}
		fmt.Fprintf(sb, `background: linear-gradient(%s, var(--bg), var(--bg2));`, dir)
	case "texture":
		sb.WriteString(`background: var(--bg);`)
	default:
		sb.WriteString(`background: var(--bg);`)
	}
	sb.WriteString(`"></div>`)
}

func (c *HTMLCompiler) writeTextZone(sb *strings.Builder, el domain.Element, content domain.CardContent) {
	text := ""
	if el.TextZone != nil {
		switch el.TextZone.SemanticRole {
		case "headline":
			text = content.Headline
		case "body":
			text = content.Body
		case "closing":
			text = content.Closing
		case "signature":
			text = content.Signature
		case "recipient":
			text = content.Recipient
		}
	}
	fmt.Fprintf(sb, `<div id="%s">%s</div>`, el.ID, escapeHTML(text))
}

func (c *HTMLCompiler) writePhotoSlot(sb *strings.Builder, el domain.Element, photos map[string]string) {
	if b64, ok := photos[el.ID]; ok && b64 != "" {
		style := ""
		if el.PhotoSlot != nil && el.PhotoSlot.Shape == "circle" {
			style = "border-radius: 50%; "
		}
		if el.PhotoSlot != nil && el.PhotoSlot.BorderWidth > 0 {
			style += fmt.Sprintf("border: %dpx solid %s; ", el.PhotoSlot.BorderWidth, resolveColor(el.PhotoSlot.BorderColorSource))
		}
		fmt.Fprintf(sb, `<div id="%s"><img src="data:image/png;base64,%s" style="%swidth:100%%;height:100%%;object-fit:cover;" /></div>`, el.ID, b64, style)
		return
	}

	placeholder := "Photo"
	if el.PhotoSlot != nil && el.PhotoSlot.PlaceholderText != "" {
		placeholder = el.PhotoSlot.PlaceholderText
	}
	style := "display:flex;align-items:center;justify-content:center;background:var(--overlay);color:var(--muted);font-size:12px;"
	if el.PhotoSlot != nil && el.PhotoSlot.Shape == "circle" {
		style += "border-radius:50%;"
	}
	fmt.Fprintf(sb, `<div id="%s" style="%s">%s</div>`, el.ID, style, escapeHTML(placeholder))
}

func (c *HTMLCompiler) writeDataSlot(sb *strings.Builder, el domain.Element, dataFields map[string]string) {
	text := ""
	if el.DataSlot != nil {
		value := dataFields[el.DataSlot.Field]
		if value != "" && el.DataSlot.FormatTemplate != "" {
			text = strings.ReplaceAll(el.DataSlot.FormatTemplate, "{{value}}", value)
		} else {
			text = value
		}
	}
	fmt.Fprintf(sb, `<div id="%s">%s</div>`, el.ID, escapeHTML(text))
}

func (c *HTMLCompiler) writeDecorative(sb *strings.Builder, el domain.Element) {
	if el.Decorative == nil {
		fmt.Fprintf(sb, `<div id="%s"></div>`, el.ID)
		return
	}

	style := ""
	if el.Decorative.Opacity < 1 && el.Decorative.Opacity > 0 {
		style += fmt.Sprintf("opacity:%.2f;", el.Decorative.Opacity)
	}
	if el.Decorative.FlipX || el.Decorative.FlipY {
		sx, sy := "1", "1"
		if el.Decorative.FlipX {
			sx = "-1"
		}
		if el.Decorative.FlipY {
			sy = "-1"
		}
		style += fmt.Sprintf("transform:scale(%s,%s);", sx, sy)
	}

	if c.assetLib != nil {
		match := c.assetLib.FindByID(el.Decorative.AssetID)
		if match != nil {
			fmt.Fprintf(sb, `<div id="%s" style="%s"><img src="data:image/png;base64,%s" style="width:100%%;height:100%%;object-fit:contain;" /></div>`, el.ID, style, *match)
			return
		}
	}

	fmt.Fprintf(sb, `<div id="%s" style="%s"></div>`, el.ID, style)
}

func (c *HTMLCompiler) writeIllustration(sb *strings.Builder, el domain.Element, illustrations map[string]string) {
	slotName := ""
	style := ""
	if el.IllustrationSlot != nil {
		slotName = el.IllustrationSlot.SlotName
		if el.IllustrationSlot.Opacity < 1 && el.IllustrationSlot.Opacity > 0 {
			style += fmt.Sprintf("opacity:%.2f;", el.IllustrationSlot.Opacity)
		}
		if el.IllustrationSlot.Shape == "circle" {
			style += "border-radius:50%;overflow:hidden;"
		}
	}

	if b64, ok := illustrations[slotName]; ok && b64 != "" {
		fmt.Fprintf(sb, `<div id="%s" style="%s"><img src="data:image/png;base64,%s" style="width:100%%;height:100%%;object-fit:cover;" /></div>`, el.ID, style, b64)
		return
	}

	fmt.Fprintf(sb, `<div id="%s" style="%s"></div>`, el.ID, style)
}

func resolveColor(source string) string {
	switch source {
	case "palette.primary":
		return "var(--primary)"
	case "palette.accent":
		return "var(--accent)"
	case "palette.ink":
		return "var(--ink)"
	case "palette.muted":
		return "var(--muted)"
	case "palette.background":
		return "var(--bg)"
	default:
		return "var(--ink)"
	}
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
