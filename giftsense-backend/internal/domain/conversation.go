package domain

import "time"

type Message struct {
	Index     int
	Sender    string
	Text      string
	Timestamp time.Time
	IsMedia   bool
}

type ChunkMetadata struct {
	Topics           []string
	EmotionalMarkers []string
	HasPreference    bool
	HasWish          bool
}

type Chunk struct {
	ID             string
	SessionID      string
	AnonymizedText string
	StartIndex     int
	EndIndex       int
	Metadata       ChunkMetadata
}
