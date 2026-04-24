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
}
