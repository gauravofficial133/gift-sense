package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/giftsense/backend/internal/domain"
)

const (
	spotifyAccountsURL = "https://accounts.spotify.com/api/token"
	spotifyAPIBaseURL  = "https://api.spotify.com/v1"
	tokenBufferSeconds = 300 // refresh 5 minutes before expiry
)

// Client implements port.SpotifyClient using the Spotify Web API with Client Credentials flow.
type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	accountsURL  string
	apiBaseURL   string

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		accountsURL:  spotifyAccountsURL,
		apiBaseURL:   spotifyAPIBaseURL,
	}
}

// WithBaseURLs overrides both the accounts and API base URLs. Used in tests.
func (c *Client) WithBaseURLs(accountsURL, apiBaseURL string) *Client {
	c.accountsURL = accountsURL
	c.apiBaseURL = apiBaseURL
	return c
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]domain.SpotifyTrack, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrSpotifySearchFailed, err)
	}

	if limit <= 0 || limit > 10 {
		limit = 5
	}

	reqURL := fmt.Sprintf("%s/search?q=%s&type=track&limit=%d",
		c.apiBaseURL, url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrSpotifySearchFailed, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrSpotifySearchFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %s", domain.ErrSpotifySearchFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", domain.ErrSpotifySearchFailed, resp.StatusCode, string(body))
	}

	var searchResp spotifySearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("%w: parse response: %s", domain.ErrSpotifySearchFailed, err)
	}

	tracks := make([]domain.SpotifyTrack, 0, len(searchResp.Tracks.Items))
	for _, item := range searchResp.Tracks.Items {
		artist := ""
		if len(item.Artists) > 0 {
			artist = item.Artists[0].Name
		}
		albumArt := ""
		if len(item.Album.Images) > 0 {
			albumArt = item.Album.Images[0].URL
		}
		tracks = append(tracks, domain.SpotifyTrack{
			ID:         item.ID,
			Name:       item.Name,
			Artist:     artist,
			AlbumArt:   albumArt,
			PreviewURL: item.PreviewURL,
			SpotifyURI: item.URI,
		})
	}

	return tracks, nil
}

func (c *Client) GetAudioFeatures(ctx context.Context, trackID string) (domain.SpotifyAudioFeatures, error) {
	if err := c.ensureToken(ctx); err != nil {
		return domain.SpotifyAudioFeatures{}, fmt.Errorf("spotify audio features: %w", err)
	}

	reqURL := fmt.Sprintf("%s/audio-features/%s", c.apiBaseURL, url.PathEscape(trackID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return domain.SpotifyAudioFeatures{}, fmt.Errorf("spotify audio features: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.SpotifyAudioFeatures{}, fmt.Errorf("spotify audio features: %w", err)
	}
	defer resp.Body.Close()

	// Gracefully degrade if the audio features endpoint is deprecated (403).
	if resp.StatusCode == http.StatusForbidden {
		log.Printf("spotify: audio features endpoint returned 403 (likely deprecated), returning zero values for track %s", trackID)
		return domain.SpotifyAudioFeatures{}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.SpotifyAudioFeatures{}, fmt.Errorf("spotify audio features: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("spotify: audio features returned %d for track %s, returning zero values", resp.StatusCode, trackID)
		return domain.SpotifyAudioFeatures{}, nil
	}

	var featResp audioFeaturesResponse
	if err := json.Unmarshal(body, &featResp); err != nil {
		return domain.SpotifyAudioFeatures{}, fmt.Errorf("spotify audio features: parse: %w", err)
	}

	return domain.SpotifyAudioFeatures{
		Valence:      featResp.Valence,
		Energy:       featResp.Energy,
		Danceability: featResp.Danceability,
		Tempo:        featResp.Tempo,
	}, nil
}

// ensureToken acquires or refreshes the Client Credentials token.
func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	data := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.accountsURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(c.clientID+":"+c.clientSecret),
	))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("parse token response: %w", err)
	}

	c.token = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-tokenBufferSeconds) * time.Second)

	return nil
}

// ── Spotify API response types ──────────────────────────────────────────────

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type spotifySearchResponse struct {
	Tracks struct {
		Items []spotifyTrackItem `json:"items"`
	} `json:"tracks"`
}

type spotifyTrackItem struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Artists    []spotifyArtist  `json:"artists"`
	Album      spotifyAlbum     `json:"album"`
	PreviewURL string           `json:"preview_url"`
	URI        string           `json:"uri"`
}

type spotifyArtist struct {
	Name string `json:"name"`
}

type spotifyAlbum struct {
	Images []spotifyImage `json:"images"`
}

type spotifyImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type audioFeaturesResponse struct {
	Valence      float64 `json:"valence"`
	Energy       float64 `json:"energy"`
	Danceability float64 `json:"danceability"`
	Tempo        float64 `json:"tempo"`
}
