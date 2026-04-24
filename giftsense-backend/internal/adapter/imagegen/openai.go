package imagegen

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/giftsense/backend/internal/port"
)

type DallEGenerator struct {
	client openai.Client
	model  string
}

func NewDallEGenerator(apiKey, model string) (*DallEGenerator, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required for image generation")
	}
	if model == "" {
		model = "dall-e-3"
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &DallEGenerator{client: client, model: model}, nil
}

func (g *DallEGenerator) Generate(ctx context.Context, req port.ImageRequest) (*port.ImageResult, error) {
	size := pickSize(req.Width, req.Height)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := g.client.Images.Generate(ctx, openai.ImageGenerateParams{
		Prompt:         req.Prompt,
		Model:          openai.ImageModel(g.model),
		N:              openai.Int(1),
		Size:           openai.ImageGenerateParamsSize(size),
		ResponseFormat: openai.ImageGenerateParamsResponseFormatB64JSON,
		Quality:        openai.ImageGenerateParamsQualityStandard,
	})
	if err != nil {
		return nil, fmt.Errorf("dall-e generate: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("dall-e returned no images")
	}

	return &port.ImageResult{PNGBase64: resp.Data[0].B64JSON}, nil
}

func pickSize(w, h int) string {
	ratio := float64(w) / float64(h)
	switch {
	case ratio > 1.2:
		return "1792x1024"
	case ratio < 0.8:
		return "1024x1792"
	default:
		return "1024x1024"
	}
}
