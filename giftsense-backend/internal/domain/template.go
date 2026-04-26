package domain

import "time"

type TemplateDefinition struct {
	Version        int              `json:"version"`
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Family         string           `json:"family,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	Occasions      []OccasionKey    `json:"occasions"`
	Emotions       []EmotionGroup   `json:"emotions"`
	Themes         []string         `json:"themes"`
	Tier           string           `json:"tier"`
	Canvas         CanvasSpec       `json:"canvas"`
	VariationRules VariationRules   `json:"variation_rules"`
	Elements       []Element        `json:"elements"`
}

type CanvasSpec struct {
	Orientation string `json:"orientation"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

type VariationRules struct {
	PaletteMode       string             `json:"palette_mode"`
	AllowedPalettes   []string           `json:"allowed_palettes"`
	PaletteMood       string             `json:"palette_mood"`
	BackgroundOptions []BackgroundOption `json:"background_options"`
	LayoutJitter      LayoutJitter       `json:"layout_jitter"`
}

type BackgroundOption struct {
	Type      string `json:"type"`
	Direction string `json:"direction,omitempty"`
	TextureID string `json:"texture_id,omitempty"`
}

type LayoutJitter struct {
	PositionRangePx int `json:"position_range_px"`
	SizeRangePct    int `json:"size_range_pct"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Size struct {
	W int `json:"w"`
	H int `json:"h"`
}

type Range struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type Element struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	ZIndex           int               `json:"z_index"`
	Position         *Position         `json:"position,omitempty"`
	Size             *Size             `json:"size,omitempty"`
	Rotation         float64           `json:"rotation"`
	TextZone         *TextZone         `json:"text_zone,omitempty"`
	IllustrationSlot *IllustrationSlot `json:"illustration_slot,omitempty"`
	DataSlot         *DataSlot         `json:"data_slot,omitempty"`
	PhotoSlot        *PhotoSlot        `json:"photo_slot,omitempty"`
	Decorative       *DecorativeSpec   `json:"decorative,omitempty"`
}

type TextZone struct {
	Purpose       string   `json:"purpose"`
	Tone          string   `json:"tone"`
	CharMin       int      `json:"char_min"`
	CharMax       int      `json:"char_max"`
	FontOptions   []string `json:"font_options"`
	FontSizeRange Range    `json:"font_size_range"`
	FontWeight    string   `json:"font_weight"`
	ColorSource   string   `json:"color_source"`
	Alignment     string   `json:"alignment"`
	SemanticRole  string   `json:"semantic_role"`
}

type IllustrationSlot struct {
	SlotName  string  `json:"slot_name"`
	StyleHint string  `json:"style_hint"`
	Shape     string  `json:"shape"`
	Opacity   float64 `json:"opacity"`
}

type DataSlot struct {
	Field          string   `json:"field"`
	FormatTemplate string   `json:"format_template"`
	FontOptions    []string `json:"font_options"`
	FontSize       int      `json:"font_size"`
	ColorSource    string   `json:"color_source"`
	Alignment      string   `json:"alignment"`
}

type PhotoSlot struct {
	Shape            string `json:"shape"`
	BorderColorSource string `json:"border_color_source"`
	BorderWidth      int    `json:"border_width"`
	PlaceholderText  string `json:"placeholder_text"`
}

type DecorativeSpec struct {
	AssetID string  `json:"asset_id"`
	Opacity float64 `json:"opacity"`
	FlipX   bool    `json:"flip_x"`
	FlipY   bool    `json:"flip_y"`
}
