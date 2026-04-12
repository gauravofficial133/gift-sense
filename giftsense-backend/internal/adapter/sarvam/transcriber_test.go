package sarvam_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/giftsense/backend/internal/adapter/sarvam"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// batchServer spins up a fake Sarvam batch STT server that serves the full
// 5-step flow. Callers can inject the transcript and language code it returns.
func batchServer(t *testing.T, transcript, langCode string) *httptest.Server {
	t.Helper()
	const (
		jobID      = "test-job-123"
		outputFile = "output.json"
	)

	mux := http.NewServeMux()

	// Step 1 — initiate job
	mux.HandleFunc("/speech-to-text/job/v1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"job_id":    jobID,
			"job_state": "Accepted",
		})
	})

	// Step 2 — get upload URLs
	mux.HandleFunc("/speech-to-text/job/v1/upload-files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		files, _ := req["files"].([]interface{})
		uploadURLs := map[string]interface{}{}
		for _, f := range files {
			fname := f.(string)
			uploadURLs[fname] = map[string]string{
				"file_url": "http://" + r.Host + "/upload/" + fname,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id":     jobID,
			"upload_urls": uploadURLs,
		})
	})

	// Step 3 — presigned upload (PUT)
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	// Step 4 — start job
	mux.HandleFunc("/speech-to-text/job/v1/"+jobID+"/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job_state": "Running"})
	})

	// Step 5 — status (returns Completed immediately)
	mux.HandleFunc("/speech-to-text/job/v1/"+jobID+"/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id":    jobID,
			"job_state": "Completed",
			"job_details": []map[string]interface{}{
				{
					"state": "Success",
					"outputs": []map[string]string{
						{"file_name": outputFile},
					},
				},
			},
		})
	})

	// Step 6 — download URLs
	mux.HandleFunc("/speech-to-text/job/v1/download-files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id": jobID,
			"download_urls": map[string]interface{}{
				outputFile: map[string]string{
					"file_url": "http://" + r.Host + "/download/" + outputFile,
				},
			},
		})
	})

	// Step 7 — transcript JSON
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"transcript":    transcript,
			"language_code": langCode,
		})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTranscriber(baseURL string) *sarvam.Transcriber {
	return sarvam.NewTranscriber("test-key").WithBaseURL(baseURL + "/speech-to-text/job/v1")
}

// ── constructor ───────────────────────────────────────────────────────────────

func TestNewTranscriber_ShouldReturnNonNilTranscriber_WhenAPIKeyProvided(t *testing.T) {
	tr := sarvam.NewTranscriber("sk-sarvam-test")
	assert.NotNil(t, tr)
}

func TestNewTranscriber_ShouldReturnNonNilTranscriber_WhenAPIKeyIsEmpty(t *testing.T) {
	tr := sarvam.NewTranscriber("")
	assert.NotNil(t, tr)
}

// ── happy path ────────────────────────────────────────────────────────────────

func TestTranscribe_ShouldReturnTranscript_WhenBatchJobSucceeds(t *testing.T) {
	srv := batchServer(t, "Tujhe kitna chahne lage hum", "hi-IN")

	result, err := newTranscriber(srv.URL).Transcribe(context.Background(), port.TranscribeRequest{
		Data:     []byte("fake-audio-bytes"),
		Filename: "voice.mp3",
	})

	require.NoError(t, err)
	assert.Equal(t, "Tujhe kitna chahne lage hum", result.Transcript)
	assert.Equal(t, "hi-IN", result.LanguageCode)
}

func TestTranscribe_ShouldHandleUnicodeTranscript_WhenDevanagariText(t *testing.T) {
	hindiText := "तुझे कितना चाहने लगे हम, अभी जान के भी ना जाना तुम"
	srv := batchServer(t, hindiText, "hi-IN")

	result, err := newTranscriber(srv.URL).Transcribe(context.Background(), port.TranscribeRequest{
		Data:     []byte("audio"),
		Filename: "hindi.mp3",
	})

	require.NoError(t, err)
	assert.Equal(t, hindiText, result.Transcript)
}

func TestTranscribe_ShouldSendAuthHeader_WhenAPIKeySet(t *testing.T) {
	var receivedKey string
	srv := batchServer(t, "test", "en-IN")

	// Wrap the server to capture auth header on initiate
	inner := srv
	outer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/speech-to-text/job/v1" && r.Method == http.MethodPost {
			receivedKey = r.Header.Get("api-subscription-key")
		}
		inner.Config.Handler.ServeHTTP(w, r)
	}))
	defer outer.Close()

	sarvam.NewTranscriber("my-secret-key").
		WithBaseURL(outer.URL+"/speech-to-text/job/v1").
		Transcribe(context.Background(), port.TranscribeRequest{
			Data:     []byte("audio"),
			Filename: "a.mp3",
		})

	assert.Equal(t, "my-secret-key", receivedKey)
}

func TestTranscribe_ShouldBuildTranscriptFromSegments_WhenTopLevelTranscriptEmpty(t *testing.T) {
	mux := http.NewServeMux()
	jobID := "seg-job"
	mux.HandleFunc("/speech-to-text/job/v1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_id": jobID, "job_state": "Accepted"})
	})
	mux.HandleFunc("/speech-to-text/job/v1/upload-files", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"upload_urls": map[string]interface{}{
				"a.mp3": map[string]string{"file_url": "http://" + r.Host + "/upload/a.mp3"},
			},
		})
	})
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
	mux.HandleFunc("/speech-to-text/job/v1/"+jobID+"/start", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_state": "Running"})
	})
	mux.HandleFunc("/speech-to-text/job/v1/"+jobID+"/status", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job_state": "Completed",
			"job_details": []map[string]interface{}{
				{"state": "Success", "outputs": []map[string]string{{"file_name": "out.json"}}},
			},
		})
	})
	mux.HandleFunc("/speech-to-text/job/v1/download-files", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"download_urls": map[string]interface{}{
				"out.json": map[string]string{"file_url": "http://" + r.Host + "/download/out.json"},
			},
		})
	})
	mux.HandleFunc("/download/", func(w http.ResponseWriter, _ *http.Request) {
		// transcript is empty, segments contain the text
		json.NewEncoder(w).Encode(map[string]interface{}{
			"transcript":    "",
			"language_code": "hi-IN",
			"segments": []map[string]string{
				{"text": "Hello"},
				{"text": "world"},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	result, err := sarvam.NewTranscriber("k").
		WithBaseURL(srv.URL+"/speech-to-text/job/v1").
		Transcribe(context.Background(), port.TranscribeRequest{Data: []byte("x"), Filename: "a.mp3"})

	require.NoError(t, err)
	assert.Equal(t, "Hello world", result.Transcript)
}

func TestTranscribe_ShouldForwardAudioBytesUnchanged_WhenUploading(t *testing.T) {
	audioPayload := []byte("fake-mp3-bytes-representing-audio-content")
	var receivedBytes []byte

	srv := batchServer(t, "some transcript", "hi-IN")

	// Intercept the PUT upload to capture the bytes
	mux := http.NewServeMux()
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		receivedBytes = b
		w.WriteHeader(http.StatusCreated)
	})
	// Forward everything else to the real batch server
	mux.HandleFunc("/", srv.Config.Handler.ServeHTTP)

	outer := httptest.NewServer(mux)
	defer outer.Close()

	_, err := sarvam.NewTranscriber("k").
		WithBaseURL(outer.URL+"/speech-to-text/job/v1").
		Transcribe(context.Background(), port.TranscribeRequest{
			Data:     audioPayload,
			Filename: "voice.mp3",
		})

	require.NoError(t, err)
	assert.Equal(t, audioPayload, receivedBytes, "audio bytes must be forwarded to upload endpoint unchanged")
}

func TestTranscribe_ShouldPassLanguageCode_WhenProvided(t *testing.T) {
	var receivedLangCode string

	srv := batchServer(t, "hello", "en-IN")

	mux := http.NewServeMux()
	mux.HandleFunc("/speech-to-text/job/v1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if params, ok := body["job_parameters"].(map[string]interface{}); ok {
				receivedLangCode, _ = params["language_code"].(string)
			}
		}
		srv.Config.Handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("/", srv.Config.Handler.ServeHTTP)

	outer := httptest.NewServer(mux)
	defer outer.Close()

	_, err := sarvam.NewTranscriber("k").
		WithBaseURL(outer.URL+"/speech-to-text/job/v1").
		Transcribe(context.Background(), port.TranscribeRequest{
			Data:         []byte("audio"),
			Filename:     "a.mp3",
			LanguageCode: "en-IN",
		})

	require.NoError(t, err)
	assert.Equal(t, "en-IN", receivedLangCode, "language_code from request must be forwarded to job_parameters")
}

func TestTranscribe_ShouldUseUnknownLanguageCode_WhenNotProvided(t *testing.T) {
	var receivedLangCode string

	srv := batchServer(t, "hello", "hi-IN")

	mux := http.NewServeMux()
	mux.HandleFunc("/speech-to-text/job/v1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if params, ok := body["job_parameters"].(map[string]interface{}); ok {
				receivedLangCode, _ = params["language_code"].(string)
			}
		}
		srv.Config.Handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("/", srv.Config.Handler.ServeHTTP)

	outer := httptest.NewServer(mux)
	defer outer.Close()

	_, err := sarvam.NewTranscriber("k").
		WithBaseURL(outer.URL+"/speech-to-text/job/v1").
		Transcribe(context.Background(), port.TranscribeRequest{
			Data:     []byte("audio"),
			Filename: "a.mp3",
			// LanguageCode intentionally empty
		})

	require.NoError(t, err)
	assert.Equal(t, "unknown", receivedLangCode, "language_code should default to 'unknown' when not provided")
}

// ── error paths ───────────────────────────────────────────────────────────────

func TestTranscribe_ShouldReturnError_WhenInitiateJobFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_api_key"}`))
	}))
	defer srv.Close()

	_, err := sarvam.NewTranscriber("bad-key").
		WithBaseURL(srv.URL+"/speech-to-text/job/v1").
		Transcribe(context.Background(), port.TranscribeRequest{Data: []byte("x"), Filename: "a.mp3"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTranscriptionFailed)
}

func TestTranscribe_ShouldReturnError_WhenJobFails(t *testing.T) {
	mux := http.NewServeMux()
	jobID := "fail-job"
	mux.HandleFunc("/speech-to-text/job/v1", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_id": jobID, "job_state": "Accepted"})
	})
	mux.HandleFunc("/speech-to-text/job/v1/upload-files", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"upload_urls": map[string]interface{}{
				"a.mp3": map[string]string{"file_url": "http://" + r.Host + "/upload/a.mp3"},
			},
		})
	})
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
	mux.HandleFunc("/speech-to-text/job/v1/"+jobID+"/start", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_state": "Running"})
	})
	mux.HandleFunc("/speech-to-text/job/v1/"+jobID+"/status", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"job_state":     "Failed",
			"error_message": "unsupported audio codec",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, err := sarvam.NewTranscriber("k").
		WithBaseURL(srv.URL+"/speech-to-text/job/v1").
		Transcribe(context.Background(), port.TranscribeRequest{Data: []byte("x"), Filename: "a.mp3"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTranscriptionFailed)
	assert.True(t, strings.Contains(err.Error(), "unsupported audio codec"))
}

func TestTranscribe_ShouldReturnError_WhenTranscriptEmpty(t *testing.T) {
	srv := batchServer(t, "", "hi-IN") // empty transcript

	_, err := newTranscriber(srv.URL).Transcribe(context.Background(), port.TranscribeRequest{
		Data:     []byte("silent"),
		Filename: "silent.wav",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTranscriptionFailed)
}

func TestTranscribe_ShouldReturnError_WhenContextAlreadyCancelled(t *testing.T) {
	srv := batchServer(t, "hello", "en-IN")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := newTranscriber(srv.URL).Transcribe(ctx, port.TranscribeRequest{
		Data:     []byte("audio"),
		Filename: "a.mp3",
	})

	require.Error(t, err)
}
