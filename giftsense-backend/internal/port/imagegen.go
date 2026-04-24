package port

import "context"

type ImageRequest struct {
	Prompt string
	Width  int
	Height int
}

type ImageResult struct {
	PNGBase64 string
}

type ImageGenerator interface {
	Generate(ctx context.Context, req ImageRequest) (*ImageResult, error)
}
