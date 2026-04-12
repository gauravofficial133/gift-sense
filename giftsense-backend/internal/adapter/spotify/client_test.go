package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestSearch_ShouldReturnTracks_WhenQueryMatches(t *testing.T) {
	tokenServer := newTokenServer(t)
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query().Get("q")
		if q != "tum hi ho" {
			t.Fatalf("expected query 'tum hi ho', got %q", q)
		}
		json.NewEncoder(w).Encode(spotifySearchResponse{
			Tracks: struct {
				Items []spotifyTrackItem `json:"items"`
			}{
				Items: []spotifyTrackItem{
					{
						ID:   "abc123",
						Name: "Tum Hi Ho",
						Artists: []spotifyArtist{{Name: "Arijit Singh"}},
						Album: spotifyAlbum{Images: []spotifyImage{{URL: "https://img.spotify.com/abc.jpg"}}},
						URI:  "spotify:track:abc123",
					},
				},
			},
		})
	}))
	defer apiServer.Close()

	client := NewClient("test-id", "test-secret").WithBaseURLs(tokenServer.URL, apiServer.URL)

	tracks, err := client.Search(context.Background(), "tum hi ho", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Name != "Tum Hi Ho" {
		t.Errorf("expected track name 'Tum Hi Ho', got %q", tracks[0].Name)
	}
	if tracks[0].Artist != "Arijit Singh" {
		t.Errorf("expected artist 'Arijit Singh', got %q", tracks[0].Artist)
	}
	if tracks[0].AlbumArt != "https://img.spotify.com/abc.jpg" {
		t.Errorf("expected album art URL, got %q", tracks[0].AlbumArt)
	}
}

func TestSearch_ShouldReturnError_WhenAPIFails(t *testing.T) {
	tokenServer := newTokenServer(t)
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer apiServer.Close()

	client := NewClient("test-id", "test-secret").WithBaseURLs(tokenServer.URL, apiServer.URL)

	_, err := client.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetAudioFeatures_ShouldReturnFeatures_WhenTrackExists(t *testing.T) {
	tokenServer := newTokenServer(t)
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio-features/abc123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(audioFeaturesResponse{
			Valence:      0.8,
			Energy:       0.6,
			Danceability: 0.7,
			Tempo:        120.5,
		})
	}))
	defer apiServer.Close()

	client := NewClient("test-id", "test-secret").WithBaseURLs(tokenServer.URL, apiServer.URL)

	features, err := client.GetAudioFeatures(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if features.Valence != 0.8 {
		t.Errorf("expected valence 0.8, got %f", features.Valence)
	}
	if features.Energy != 0.6 {
		t.Errorf("expected energy 0.6, got %f", features.Energy)
	}
	if features.Tempo != 120.5 {
		t.Errorf("expected tempo 120.5, got %f", features.Tempo)
	}
}

func TestGetAudioFeatures_ShouldReturnZero_WhenEndpointDeprecated(t *testing.T) {
	tokenServer := newTokenServer(t)
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer apiServer.Close()

	client := NewClient("test-id", "test-secret").WithBaseURLs(tokenServer.URL, apiServer.URL)

	features, err := client.GetAudioFeatures(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("expected no error on 403, got: %v", err)
	}
	if features.Valence != 0 || features.Energy != 0 {
		t.Errorf("expected zero-valued features on 403, got %+v", features)
	}
}

func TestEnsureToken_ShouldCacheToken_WhenCalledMultipleTimes(t *testing.T) {
	var callCount int32
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "cached-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(spotifySearchResponse{})
	}))
	defer apiServer.Close()

	client := NewClient("test-id", "test-secret").WithBaseURLs(tokenServer.URL, apiServer.URL)

	// Make two search calls — token should only be fetched once.
	client.Search(context.Background(), "a", 1)
	client.Search(context.Background(), "b", 1)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 token request (cached), got %d", callCount)
	}
}

func TestEnsureToken_ShouldReturnError_WhenAuthFails(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer tokenServer.Close()

	client := NewClient("bad-id", "bad-secret").WithBaseURLs(tokenServer.URL, "http://unused")

	_, err := client.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error on 401, got nil")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func newTokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
}
