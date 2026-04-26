package port

import (
	"context"

	"github.com/giftsense/backend/internal/domain"
)

type TemplateStore interface {
	List(ctx context.Context) ([]domain.TemplateDefinition, error)
	Get(ctx context.Context, id string) (*domain.TemplateDefinition, error)
	Save(ctx context.Context, tpl domain.TemplateDefinition) error
	Delete(ctx context.Context, id string) error
	SavePreview(ctx context.Context, id string, pngBase64 string) error
	GetPreview(ctx context.Context, id string) (string, error)
}
