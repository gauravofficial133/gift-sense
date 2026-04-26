package domain

type OccasionKey string
type EmotionGroup string

type CardPalette struct {
	Background          string
	BackgroundSecondary string
	Primary             string
	Accent              string
	Ink                 string
	Muted               string
	Overlay             string
}

type OccasionTemplate struct {
	Key      OccasionKey
	Greeting string
	Motif    string
}

type CardContent struct {
	Headline    string
	Body        string
	Closing     string
	Signature   string
	Recipient   string
	Emotions    []EmotionSignal
	OccasionKey string
}

type RecipeID string

type RecipeMetadata struct {
	ID         RecipeID      `json:"id"`
	Name       string        `json:"name"`
	Tier       string        `json:"tier"`
	Occasions  []OccasionKey `json:"occasions"`
	Emotions   []EmotionGroup `json:"emotions"`
	Orientation string       `json:"orientation"`
	HeadlineFont string      `json:"headline_font"`
	BodyFont     string      `json:"body_font"`
}

type CardIllustration struct {
	PNGBase64 string `json:"png_base64,omitempty"`
	Slot      string `json:"slot,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
}

type CardRender struct {
	PreviewPNG    string            `json:"preview_png"`
	PDFBase64     string            `json:"pdf_base64,omitempty"`
	RecipeID      string            `json:"recipe_id"`
	PaletteName   string            `json:"palette_name"`
	Content       CardContent       `json:"content"`
	Model         string            `json:"model"`
	Illustration  *CardIllustration `json:"illustration,omitempty"`
	DataFields    map[string]string `json:"data_fields,omitempty"`
	Photos        map[string]string `json:"photos,omitempty"`
	CardType      string            `json:"card_type,omitempty"`
	Meta          *CardMeta         `json:"meta,omitempty"`
	Evidences     []MemoryEvidence  `json:"evidences,omitempty"`
}

type CardMeta struct {
	Model             string            `json:"model"`
	GenerationMs      int64             `json:"generation_ms"`
	TemplateTier      string            `json:"template_tier"`
	TemplateFamily    string            `json:"template_family"`
	Orientation       string            `json:"orientation"`
	EmotionMatchScore float64           `json:"emotion_match_score"`
	Variation         string            `json:"variation"`
	Validation        *ValidationResult `json:"validation,omitempty"`
	Scoring           *ScoringBreakdown `json:"scoring,omitempty"`
}

type ValidationResult struct {
	TextOverflow     bool    `json:"text_overflow"`
	ContrastRatio    float64 `json:"contrast_ratio"`
	ContrastPassed   bool    `json:"contrast_passed"`
	IllustrationOK   bool    `json:"illustration_ok"`
	CompositionScore float64 `json:"composition_score"`
	OverallPass      bool    `json:"overall_pass"`
	Issues           []string `json:"issues,omitempty"`
}

type ScoringBreakdown struct {
	TemplateOccasionFit float64 `json:"template_occasion_fit"`
	CopyQuality         float64 `json:"copy_quality"`
	VisualHarmony       float64 `json:"visual_harmony"`
	Originality         float64 `json:"originality"`
	TotalScore          float64 `json:"total_score"`
}

type MemoryEvidence struct {
	Quote   string `json:"quote"`
	Context string `json:"context"`
	Emotion string `json:"emotion"`
}
