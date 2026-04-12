package sarvam

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

const (
	sarvamBatchBaseURL = "https://api.sarvam.ai/speech-to-text/job/v1"
	sarvamSyncURL      = "https://api.sarvam.ai/speech-to-text"
	pollInterval       = 3 * time.Second
	// syncMaxBytes is the threshold below which the fast synchronous API is used.
	// Files above this size fall back to the batch job pipeline.
	// ~25 MB covers roughly 25 minutes of compressed audio at 128 kbps.
	syncMaxBytes = 25 * 1024 * 1024
)

// Transcriber uses the Sarvam STT APIs.
// Short audio (≤ syncMaxBytes) uses the synchronous endpoint (2–5 s round-trip).
// Longer audio falls back to the batch job pipeline.
type Transcriber struct {
	apiKey     string
	baseURL    string // batch base URL
	syncURL    string // synchronous endpoint
	httpClient *http.Client
}

func NewTranscriber(apiKey string) *Transcriber {
	return &Transcriber{
		apiKey:  apiKey,
		baseURL: sarvamBatchBaseURL,
		syncURL: sarvamSyncURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// WithBaseURL overrides the batch API base URL. Used in tests.
func (t *Transcriber) WithBaseURL(url string) *Transcriber {
	t.baseURL = url
	return t
}

// WithSyncURL overrides the synchronous API URL. Used in tests.
func (t *Transcriber) WithSyncURL(url string) *Transcriber {
	t.syncURL = url
	return t
}

// WithEndpoint is an alias for WithBaseURL kept for test compatibility.
func (t *Transcriber) WithEndpoint(url string) *Transcriber {
	return t.WithBaseURL(url)
}

// ── request/response types ───────────────────────────────────────────────────

type initiateJobReq struct {
	JobParameters batchJobParams `json:"job_parameters"`
}

type batchJobParams struct {
	LanguageCode string `json:"language_code"`
	Model        string `json:"model"`
}

type jobInitResp struct {
	JobID                string `json:"job_id"`
	JobState             string `json:"job_state"`
	StorageContainerType string `json:"storage_container_type"`
}

type uploadFilesReq struct {
	JobID string   `json:"job_id"`
	Files []string `json:"files"`
}

type signedURLDetails struct {
	FileURL string `json:"file_url"`
}

type uploadFilesResp struct {
	UploadURLs           map[string]signedURLDetails `json:"upload_urls"`
	StorageContainerType string                      `json:"storage_container_type"`
}

type taskFileDetail struct {
	FileName string `json:"file_name"`
}

type taskDetail struct {
	Outputs      []taskFileDetail `json:"outputs"`
	State        string           `json:"state"`
	ErrorMessage string           `json:"error_message"`
}

type jobStatusResp struct {
	JobID        string       `json:"job_id"`
	JobState     string       `json:"job_state"`
	JobDetails   []taskDetail `json:"job_details"`
	ErrorMessage string       `json:"error_message"`
}

type downloadFilesReq struct {
	JobID string   `json:"job_id"`
	Files []string `json:"files"`
}

type downloadFilesResp struct {
	DownloadURLs map[string]signedURLDetails `json:"download_urls"`
}

type transcriptJSON struct {
	Transcript   string       `json:"transcript"`
	LanguageCode string       `json:"language_code"`
	Segments     []sttSegment `json:"segments"`
}

type sttSegment struct {
	Text string `json:"text"`
}

// ── main entry point ─────────────────────────────────────────────────────────

// errDurationExceeded is a sentinel used internally to signal that the sync
// API rejected the file for exceeding the 30-second duration limit.
type errDurationExceeded struct{ body string }

func (e errDurationExceeded) Error() string {
	return "sync API: audio duration exceeds 30 s limit"
}

// Transcribe transcribes audio. Files ≤ syncMaxBytes try the fast synchronous
// endpoint first; if the sync API rejects the file for exceeding the 30-second
// duration limit (or the file is larger than syncMaxBytes), the batch job
// pipeline is used automatically.
func (t *Transcriber) Transcribe(ctx context.Context, req port.TranscribeRequest) (port.TranscribeResult, error) {
	if len(req.Data) <= syncMaxBytes {
		result, err := t.transcribeSync(ctx, req)
		if err == nil {
			return result, nil
		}
		var durErr errDurationExceeded
		if !isErrDurationExceeded(err, &durErr) {
			return port.TranscribeResult{}, err
		}
		// Audio is short in bytes but long in duration — fall through to batch.
	}
	return t.transcribeBatch(ctx, req)
}

// isErrDurationExceeded unwraps err to find an errDurationExceeded value.
func isErrDurationExceeded(err error, target *errDurationExceeded) bool {
	var e errDurationExceeded
	if errors.As(err, &e) {
		if target != nil {
			*target = e
		}
		return true
	}
	// Also check the raw error string so the test/mock path works without
	// wrapping through fmt.Errorf("%w", errDurationExceeded{}).
	return strings.Contains(err.Error(), "Audio duration exceeds the maximum limit")
}

// transcribeSync calls the synchronous Sarvam STT endpoint (POST /speech-to-text).
// Returns in 2–5 seconds for short clips.
func (t *Transcriber) transcribeSync(ctx context.Context, req port.TranscribeRequest) (port.TranscribeResult, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", req.Filename)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: create form file: %s", domain.ErrTranscriptionFailed, err)
	}
	if _, err = fw.Write(req.Data); err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: write audio data: %s", domain.ErrTranscriptionFailed, err)
	}

	lc := req.LanguageCode
	if lc == "" {
		lc = "unknown"
	}
	if err = mw.WriteField("language_code", lc); err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: write language_code: %s", domain.ErrTranscriptionFailed, err)
	}
	if err = mw.WriteField("model", "saarika:v2.5"); err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: write model: %s", domain.ErrTranscriptionFailed, err)
	}
	mw.Close()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.syncURL, &buf)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: create sync request: %s", domain.ErrTranscriptionFailed, err)
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	httpReq.Header.Set("api-subscription-key", t.apiKey)

	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: sync request: %s", domain.ErrTranscriptionFailed, err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: read sync response: %s", domain.ErrTranscriptionFailed, err)
	}
	if resp.StatusCode != http.StatusOK {
		body := string(b)
		if resp.StatusCode == http.StatusBadRequest && strings.Contains(body, "Audio duration exceeds the maximum limit") {
			return port.TranscribeResult{}, errDurationExceeded{body: body}
		}
		return port.TranscribeResult{}, fmt.Errorf("%w: sync API returned %d: %s", domain.ErrTranscriptionFailed, resp.StatusCode, body)
	}

	var result transcriptJSON
	if err = json.Unmarshal(b, &result); err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: invalid sync transcript JSON", domain.ErrTranscriptionFailed)
	}

	transcript := result.Transcript
	if transcript == "" && len(result.Segments) > 0 {
		parts := make([]string, 0, len(result.Segments))
		for _, s := range result.Segments {
			if s.Text != "" {
				parts = append(parts, s.Text)
			}
		}
		transcript = strings.Join(parts, " ")
	}
	if transcript == "" {
		return port.TranscribeResult{}, domain.ErrTranscriptionFailed
	}

	return port.TranscribeResult{
		Transcript:   transcript,
		LanguageCode: result.LanguageCode,
	}, nil
}

// transcribeBatch runs the full async Sarvam batch job pipeline.
// Use for large files (> syncMaxBytes) where the sync endpoint may time out.
func (t *Transcriber) transcribeBatch(ctx context.Context, req port.TranscribeRequest) (port.TranscribeResult, error) {
	jobID, err := t.initiateJob(ctx, req.LanguageCode)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: initiate: %s", domain.ErrTranscriptionFailed, err)
	}

	uploadURL, err := t.getUploadURL(ctx, jobID, req.Filename)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: upload URL: %s", domain.ErrTranscriptionFailed, err)
	}

	if err = t.uploadAudio(ctx, uploadURL, req.Data, req.Filename); err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: upload audio: %s", domain.ErrTranscriptionFailed, err)
	}

	if err = t.startJob(ctx, jobID); err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: start job: %s", domain.ErrTranscriptionFailed, err)
	}

	outputFile, err := t.pollUntilComplete(ctx, jobID)
	if err != nil {
		return port.TranscribeResult{}, err
	}

	downloadURL, err := t.getDownloadURL(ctx, jobID, outputFile)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: download URL: %s", domain.ErrTranscriptionFailed, err)
	}

	return t.fetchTranscript(ctx, downloadURL)
}

// ── step implementations ─────────────────────────────────────────────────────

func (t *Transcriber) initiateJob(ctx context.Context, languageCode string) (string, error) {
	lc := languageCode
	if lc == "" {
		lc = "unknown"
	}
	body := initiateJobReq{
		JobParameters: batchJobParams{
			LanguageCode: lc,
			Model:        "saarika:v2.5",
		},
	}
	var resp jobInitResp
	if err := t.postJSON(ctx, t.baseURL, body, &resp); err != nil {
		return "", err
	}
	if resp.JobID == "" {
		return "", fmt.Errorf("empty job_id in response")
	}
	return resp.JobID, nil
}

func (t *Transcriber) getUploadURL(ctx context.Context, jobID, filename string) (string, error) {
	body := uploadFilesReq{
		JobID: jobID,
		Files: []string{filename},
	}
	var resp uploadFilesResp
	if err := t.postJSON(ctx, t.baseURL+"/upload-files", body, &resp); err != nil {
		return "", err
	}
	details, ok := resp.UploadURLs[filename]
	if !ok || details.FileURL == "" {
		return "", fmt.Errorf("no upload URL returned for %q", filename)
	}
	return details.FileURL, nil
}

func (t *Transcriber) uploadAudio(ctx context.Context, uploadURL string, data []byte, filename string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mimeFromFilename(filename))
	req.Header.Set("x-ms-blob-type", "BlockBlob") // required for Azure SAS URLs
	req.ContentLength = int64(len(data))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (t *Transcriber) startJob(ctx context.Context, jobID string) error {
	var raw json.RawMessage
	return t.postJSON(ctx, fmt.Sprintf("%s/%s/start", t.baseURL, jobID), struct{}{}, &raw)
}

func (t *Transcriber) pollUntilComplete(ctx context.Context, jobID string) (string, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("%w: timed out waiting for transcription job to complete", domain.ErrTranscriptionFailed)
		case <-ticker.C:
			outputFile, done, err := t.checkStatus(ctx, jobID)
			if err != nil {
				return "", fmt.Errorf("%w: %s", domain.ErrTranscriptionFailed, err)
			}
			if done {
				return outputFile, nil
			}
		}
	}
}

func (t *Transcriber) checkStatus(ctx context.Context, jobID string) (outputFile string, complete bool, err error) {
	url := fmt.Sprintf("%s/%s/status", t.baseURL, jobID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("api-subscription-key", t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, fmt.Errorf("read status response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("status check returned %d: %s", resp.StatusCode, string(b))
	}

	var status jobStatusResp
	if err = json.Unmarshal(b, &status); err != nil {
		return "", false, fmt.Errorf("invalid status JSON: %w", err)
	}

	switch status.JobState {
	case "Completed":
		for _, detail := range status.JobDetails {
			if detail.State == "Success" && len(detail.Outputs) > 0 {
				return detail.Outputs[0].FileName, true, nil
			}
			if detail.ErrorMessage != "" {
				return "", false, fmt.Errorf("task failed: %s", detail.ErrorMessage)
			}
		}
		return "", false, fmt.Errorf("job completed but no output files found")
	case "Failed":
		msg := status.ErrorMessage
		if msg == "" {
			msg = "job failed without an error message"
		}
		return "", false, fmt.Errorf("job failed: %s", msg)
	default:
		// Accepted, Pending, Running — keep polling
		return "", false, nil
	}
}

func (t *Transcriber) getDownloadURL(ctx context.Context, jobID, outputFile string) (string, error) {
	body := downloadFilesReq{
		JobID: jobID,
		Files: []string{outputFile},
	}
	var resp downloadFilesResp
	if err := t.postJSON(ctx, t.baseURL+"/download-files", body, &resp); err != nil {
		return "", err
	}
	details, ok := resp.DownloadURLs[outputFile]
	if !ok || details.FileURL == "" {
		return "", fmt.Errorf("no download URL returned for %q", outputFile)
	}
	return details.FileURL, nil
}

func (t *Transcriber) fetchTranscript(ctx context.Context, downloadURL string) (port.TranscribeResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: build request: %s", domain.ErrTranscriptionFailed, err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: fetch: %s", domain.ErrTranscriptionFailed, err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: read body: %s", domain.ErrTranscriptionFailed, err)
	}

	var result transcriptJSON
	if err = json.Unmarshal(b, &result); err != nil {
		return port.TranscribeResult{}, fmt.Errorf("%w: invalid transcript JSON", domain.ErrTranscriptionFailed)
	}

	transcript := result.Transcript
	if transcript == "" && len(result.Segments) > 0 {
		parts := make([]string, 0, len(result.Segments))
		for _, s := range result.Segments {
			if s.Text != "" {
				parts = append(parts, s.Text)
			}
		}
		transcript = strings.Join(parts, " ")
	}

	if transcript == "" {
		return port.TranscribeResult{}, domain.ErrTranscriptionFailed
	}

	return port.TranscribeResult{
		Transcript:   transcript,
		LanguageCode: result.LanguageCode,
	}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (t *Transcriber) postJSON(ctx context.Context, url string, body, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-subscription-key", t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBytes))
	}
	if out != nil {
		if err = json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

func mimeFromFilename(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".opus":
		return "audio/opus"
	case ".m4a":
		return "audio/mp4"
	default:
		return "application/octet-stream"
	}
}
