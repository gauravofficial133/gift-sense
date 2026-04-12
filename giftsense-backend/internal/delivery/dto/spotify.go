package dto

import "github.com/giftsense/backend/internal/domain"

// SpotifyTrackDTO is the wire representation of a Spotify track.
type SpotifyTrackDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Artist     string `json:"artist"`
	AlbumArt   string `json:"album_art"`
	PreviewURL string `json:"preview_url"`
	SpotifyURI string `json:"spotify_uri"`
}

// SpotifySearchResponse is returned by GET /api/v1/spotify/search.
type SpotifySearchResponse struct {
	Tracks []SpotifyTrackDTO `json:"tracks"`
}

// SpotifyAudioFeaturesDTO is the wire representation of audio features.
type SpotifyAudioFeaturesDTO struct {
	Valence      float64 `json:"valence"`
	Energy       float64 `json:"energy"`
	Danceability float64 `json:"danceability"`
	Tempo        float64 `json:"tempo"`
}

// AnalyzeSongRequest is the JSON body for POST /api/v1/spotify/analyze-song.
type AnalyzeSongRequest struct {
	TrackID   string `json:"track_id"   binding:"required"`
	TrackName string `json:"track_name" binding:"required"`
	Artist    string `json:"artist"     binding:"required"`
}

// AnalyzeSongResponse is returned by POST /api/v1/spotify/analyze-song.
type AnalyzeSongResponse struct {
	TrackName     string             `json:"track_name"`
	Artist        string             `json:"artist"`
	Emotions      []EmotionSignalDTO `json:"emotions"`
	LyricsSnippet string             `json:"lyrics_snippet"`
	LanguageLabel string             `json:"language_label"`
	Cached        bool               `json:"cached"`
}

// AnalyzeFromSongRequest is the JSON body for POST /api/v1/analyze-from-song.
type AnalyzeFromSongRequest struct {
	SessionID         string             `json:"session_id"          binding:"required,uuid"`
	TrackName         string             `json:"track_name"          binding:"required"`
	Artist            string             `json:"artist"              binding:"required"`
	Name              string             `json:"name"                binding:"required"`
	Relation          string             `json:"relation"`
	Gender            string             `json:"gender"`
	Occasion          string             `json:"occasion"            binding:"required"`
	BudgetTier        string             `json:"budget_tier"         binding:"required,oneof=BUDGET MID_RANGE PREMIUM LUXURY"`
	ConfirmedEmotions []EmotionSignalDTO `json:"confirmed_emotions"`
}

func SpotifyTrackToDTO(t domain.SpotifyTrack) SpotifyTrackDTO {
	return SpotifyTrackDTO{
		ID:         t.ID,
		Name:       t.Name,
		Artist:     t.Artist,
		AlbumArt:   t.AlbumArt,
		PreviewURL: t.PreviewURL,
		SpotifyURI: t.SpotifyURI,
	}
}

func SpotifyTracksToDTO(tracks []domain.SpotifyTrack) []SpotifyTrackDTO {
	dtos := make([]SpotifyTrackDTO, len(tracks))
	for i, t := range tracks {
		dtos[i] = SpotifyTrackToDTO(t)
	}
	return dtos
}

func SpotifyAudioFeaturesToDTO(f domain.SpotifyAudioFeatures) SpotifyAudioFeaturesDTO {
	return SpotifyAudioFeaturesDTO{
		Valence:      f.Valence,
		Energy:       f.Energy,
		Danceability: f.Danceability,
		Tempo:        f.Tempo,
	}
}
