package domain

type OccasionKey string
type EmotionGroup string

type CardPalette struct {
	Background string
	Primary    string
	Accent     string
	Ink        string
	Muted      string
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

type CardRender struct {
	SVG       string      `json:"svg"`
	PDFBase64 string      `json:"pdf_base64"`
	ThemeID   string      `json:"theme_id"`
	Content   CardContent `json:"content"`
}
