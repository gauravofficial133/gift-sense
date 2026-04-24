package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/giftsense/backend/internal/port"
)

type AssetEntry struct {
	ID           string   `json:"id"`
	Prompt       string   `json:"prompt"`
	Slot         string   `json:"slot"`
	Occasion     string   `json:"occasion"`
	Emotions     []string `json:"emotions"`
	Style        string   `json:"style"`
	Tags         []string `json:"tags,omitempty"`
	Source       string   `json:"source,omitempty"`
	ThumbnailB64 string   `json:"thumbnail_b64,omitempty"`
	FilePath     string   `json:"file_path"`
	UsageCount   int      `json:"usage_count"`
	CreatedAt    string   `json:"created_at"`
}

type AssetIndex struct {
	Illustrations []AssetEntry `json:"illustrations"`
}

type AssetLibrary struct {
	mu       sync.RWMutex
	index    AssetIndex
	indexDir string
	imageGen port.ImageGenerator
}

func NewAssetLibrary(indexDir string, imageGen port.ImageGenerator) (*AssetLibrary, error) {
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("create asset directory: %w", err)
	}
	lib := &AssetLibrary{indexDir: indexDir, imageGen: imageGen}
	if err := lib.loadIndex(); err != nil {
		return nil, fmt.Errorf("load asset index: %w", err)
	}
	return lib, nil
}

func (lib *AssetLibrary) loadIndex() error {
	data, err := os.ReadFile(filepath.Join(lib.indexDir, "index.json"))
	if err != nil {
		if os.IsNotExist(err) {
			lib.index = AssetIndex{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &lib.index)
}

func (lib *AssetLibrary) saveIndex() error {
	data, err := json.MarshalIndent(lib.index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(lib.indexDir, "index.json"), data, 0644)
}

func (lib *AssetLibrary) FindByID(id string) *string {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	for _, entry := range lib.index.Illustrations {
		if entry.ID == id {
			data, err := os.ReadFile(filepath.Join(lib.indexDir, entry.FilePath))
			if err != nil {
				return nil
			}
			s := string(data)
			return &s
		}
	}
	return nil
}

func (lib *AssetLibrary) ListAssets(tags []string, style string) []AssetEntry {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	var results []AssetEntry
	for _, entry := range lib.index.Illustrations {
		if style != "" && entry.Style != style {
			continue
		}
		if len(tags) > 0 {
			matched := false
			for _, tag := range tags {
				for _, et := range entry.Tags {
					if strings.EqualFold(tag, et) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}
		results = append(results, entry)
	}
	return results
}

func (lib *AssetLibrary) SaveUpload(id, style string, tags []string, pngBase64 string) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	fileName := id + ".b64"
	if err := os.WriteFile(filepath.Join(lib.indexDir, fileName), []byte(pngBase64), 0644); err != nil {
		return fmt.Errorf("write asset file: %w", err)
	}

	lib.index.Illustrations = append(lib.index.Illustrations, AssetEntry{
		ID:        id,
		Style:     style,
		Tags:      tags,
		Source:    "uploaded",
		FilePath:  fileName,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})

	return lib.saveIndex()
}

func (lib *AssetLibrary) FindMatch(occasion string, emotions []string, slot string) *AssetEntry {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	var best *AssetEntry
	bestScore := 0

	for i := range lib.index.Illustrations {
		entry := &lib.index.Illustrations[i]
		if entry.Slot != slot {
			continue
		}
		score := 0
		if entry.Occasion == occasion {
			score += 3
		}
		for _, e := range emotions {
			for _, ie := range entry.Emotions {
				if strings.EqualFold(e, ie) {
					score += 2
				}
			}
		}
		if score > bestScore {
			bestScore = score
			best = entry
		}
	}

	if bestScore < 3 {
		return nil
	}
	return best
}

func (lib *AssetLibrary) GetOrGenerate(ctx context.Context, prompt, occasion string, emotions []string, slot string, width, height int) (string, error) {
	emotionStrs := make([]string, len(emotions))
	copy(emotionStrs, emotions)

	if match := lib.FindMatch(occasion, emotionStrs, slot); match != nil {
		data, err := os.ReadFile(filepath.Join(lib.indexDir, match.FilePath))
		if err == nil {
			lib.mu.Lock()
			match.UsageCount++
			lib.saveIndex()
			lib.mu.Unlock()
			return string(data), nil
		}
	}

	if lib.imageGen == nil {
		return "", fmt.Errorf("image generator not available")
	}

	result, err := lib.imageGen.Generate(ctx, port.ImageRequest{
		Prompt: prompt,
		Width:  width,
		Height: height,
	})
	if err != nil {
		return "", fmt.Errorf("generate illustration: %w", err)
	}

	lib.mu.Lock()
	defer lib.mu.Unlock()

	id := fmt.Sprintf("ill_%d", time.Now().UnixMilli())
	fileName := id + ".b64"
	if err := os.WriteFile(filepath.Join(lib.indexDir, fileName), []byte(result.PNGBase64), 0644); err != nil {
		log.Printf("failed to cache illustration: %v", err)
		return result.PNGBase64, nil
	}

	lib.index.Illustrations = append(lib.index.Illustrations, AssetEntry{
		ID:         id,
		Prompt:     prompt,
		Slot:       slot,
		Occasion:   occasion,
		Emotions:   emotionStrs,
		FilePath:   fileName,
		UsageCount: 1,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	})

	if err := lib.saveIndex(); err != nil {
		log.Printf("failed to save asset index: %v", err)
	}

	return result.PNGBase64, nil
}
