package templatestore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/giftsense/backend/internal/domain"
)

type FSStore struct {
	mu      sync.RWMutex
	baseDir string
}

func NewFSStore(baseDir string) (*FSStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create templates dir: %w", err)
	}
	return &FSStore{baseDir: baseDir}, nil
}

func (s *FSStore) List(ctx context.Context) ([]domain.TemplateDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("read templates dir: %w", err)
	}

	var templates []domain.TemplateDefinition
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		tpl, err := s.load(entry.Name())
		if err != nil {
			continue
		}
		templates = append(templates, *tpl)
	}
	return templates, nil
}

func (s *FSStore) Get(ctx context.Context, id string) (*domain.TemplateDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.load(id)
}

func (s *FSStore) Save(ctx context.Context, tpl domain.TemplateDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.baseDir, tpl.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create template dir: %w", err)
	}

	data, err := json.MarshalIndent(tpl, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal template: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, "template.json"), data, 0644)
}

func (s *FSStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.baseDir, id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("template not found: %s", id)
	}
	return os.RemoveAll(dir)
}

func (s *FSStore) SavePreview(ctx context.Context, id string, pngBase64 string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.baseDir, id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("template dir not found: %s", id)
	}
	return os.WriteFile(filepath.Join(dir, "preview.txt"), []byte(pngBase64), 0644)
}

func (s *FSStore) GetPreview(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(s.baseDir, id, "preview.txt"))
	if err != nil {
		return "", fmt.Errorf("read preview %s: %w", id, err)
	}
	return string(data), nil
}

func (s *FSStore) load(id string) (*domain.TemplateDefinition, error) {
	data, err := os.ReadFile(filepath.Join(s.baseDir, id, "template.json"))
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", id, err)
	}
	var tpl domain.TemplateDefinition
	if err := json.Unmarshal(data, &tpl); err != nil {
		return nil, fmt.Errorf("parse template %s: %w", id, err)
	}
	return &tpl, nil
}
