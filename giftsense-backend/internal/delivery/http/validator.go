package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/giftsense/backend/internal/domain"
)

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
	raw, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read uploaded file: %w", err)
	}
	return string(raw), nil
}
