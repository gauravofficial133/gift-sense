package domain

import "errors"

// SpotifyTrack holds the metadata for a Spotify track returned by search.
type SpotifyTrack struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Artist     string `json:"artist"`
	AlbumArt   string `json:"album_art"`
	PreviewURL string `json:"preview_url"`
	SpotifyURI string `json:"spotify_uri"`
}

// SpotifyAudioFeatures holds Spotify's audio analysis data for a track.
type SpotifyAudioFeatures struct {
	Valence      float64 `json:"valence"`
	Energy       float64 `json:"energy"`
	Danceability float64 `json:"danceability"`
	Tempo        float64 `json:"tempo"`
}

// CachedSongEmotion is the cached emotion analysis for a Spotify track.
type CachedSongEmotion struct {
	TrackID       string
	TrackName     string
	Artist        string
	Emotions      []EmotionSignal
	LyricsSnippet string
	LanguageLabel string
	AudioFeatures SpotifyAudioFeatures
}

var (
	ErrSpotifySearchFailed  = errors.New("spotify search failed")
	ErrSpotifyTrackNotFound = errors.New("spotify track not found")
	ErrSpotifyUnavailable   = errors.New("spotify integration is not configured")
)
