package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

var allowedAudioExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".ogg":  true,
	".opus": true,
	".m4a":  true,
}

// ValidateAudioFile checks extension and size, then returns the raw bytes.
func ValidateAudioFile(fh *multipart.FileHeader, maxBytes int64) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedAudioExtensions[ext] {
		return nil, domain.ErrAudioInvalidFormat
	}
	if fh.Size > maxBytes {
		return nil, domain.ErrAudioFileTooLarge
	}
	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("open audio file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read audio file: %w", err)
	}
	return data, nil
}

// ValidateConversationFile checks extension, size, and reads content.
func ValidateConversationFile(fh *multipart.FileHeader, maxBytes int64) (string, error) {
	if strings.ToLower(filepath.Ext(fh.Filename)) != ".txt" {
		return "", domain.ErrInvalidFileType
	}
	if fh.Size > maxBytes {
		return "", domain.ErrFileTooLarge
	}
	f, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer f.Close()

	head := make([]byte, 512)
	n, err := f.Read(head)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read file header: %w", err)
	}
	contentType := http.DetectContentType(head[:n])
	if !strings.HasPrefix(contentType, "text/") {
		return "", domain.ErrInvalidFileType
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek uploaded file: %w", err)
	}

	raw, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read uploaded file: %w", err)
	}
	return string(raw), nil
}
