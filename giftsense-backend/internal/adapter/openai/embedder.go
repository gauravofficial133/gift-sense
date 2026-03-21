package openai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Embedder struct {
	client     openai.Client
	model      string
	dimensions int
}

func NewEmbedder(apiKey, model string, dimensions int) (*Embedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("embedding model is required")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &Embedder{client: client, model: model, dimensions: dimensions}, nil
}

func (embedder *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := embedder.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model:      openai.EmbeddingModel(embedder.model),
		Input:      openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
		Dimensions: openai.Int(int64(embedder.dimensions)),
	})
	if err != nil {
		return nil, fmt.Errorf("embedding API call failed: %w", err)
	}

	vectors := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		vec := make([]float32, len(d.Embedding))
		for j, v := range d.Embedding {
			vec[j] = float32(v)
		}
		vectors[i] = vec
	}
	return vectors, nil
}
