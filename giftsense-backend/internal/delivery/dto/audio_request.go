package dto

// AnalyzeAudioRequest holds the multipart form fields for POST /api/v1/analyze-audio.
// The audio file itself is read separately via c.FormFile("audio").
type AnalyzeAudioRequest struct {
	SessionID  string `form:"session_id"  binding:"required,uuid"`
	Name       string `form:"name"        binding:"required"`
	Relation   string `form:"relation"`
	Gender     string `form:"gender"`
	Occasion   string `form:"occasion"    binding:"required"`
	BudgetTier string `form:"budget_tier" binding:"required,oneof=BUDGET MID_RANGE PREMIUM LUXURY"`
}

// EmotionSignalDTO is the wire representation of a single detected emotion.
type EmotionSignalDTO struct {
	Name      string  `json:"name"`
	Emoji     string  `json:"emoji"`
	Intensity float64 `json:"intensity"`
}

// AnalyzeFromTranscriptRequest is the JSON body for POST /api/v1/analyze-from-transcript.
type AnalyzeFromTranscriptRequest struct {
	SessionID         string             `json:"session_id"          binding:"required,uuid"`
	Transcript        string             `json:"transcript"          binding:"required,min=10"`
	Name              string             `json:"name"                binding:"required"`
	Relation          string             `json:"relation"`
	Gender            string             `json:"gender"`
	Occasion          string             `json:"occasion"            binding:"required"`
	BudgetTier        string             `json:"budget_tier"         binding:"required,oneof=BUDGET MID_RANGE PREMIUM LUXURY"`
	ConfirmedEmotions []EmotionSignalDTO `json:"confirmed_emotions"`
}
