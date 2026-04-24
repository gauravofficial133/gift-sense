package cardrender

import (
	"encoding/base64"
)

type FontData struct {
	Family string
	Weight string
	Style  string
	Base64 string
}

type TemplateEngine struct {
	fonts map[string]FontData
}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		fonts: make(map[string]FontData),
	}
}

func (e *TemplateEngine) Fonts() map[string]FontData {
	return e.fonts
}

func (e *TemplateEngine) RegisterFont(family string, weight string, style string, woff2Bytes []byte) {
	e.fonts[family+"_"+weight+"_"+style] = FontData{
		Family: family,
		Weight: weight,
		Style:  style,
		Base64: base64.StdEncoding.EncodeToString(woff2Bytes),
	}
}
