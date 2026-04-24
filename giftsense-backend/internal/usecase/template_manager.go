package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type TemplateManager struct {
	store port.TemplateStore
}

func NewTemplateManager(store port.TemplateStore) *TemplateManager {
	return &TemplateManager{store: store}
}

func (m *TemplateManager) List(ctx context.Context) ([]domain.TemplateDefinition, error) {
	return m.store.List(ctx)
}

func (m *TemplateManager) Get(ctx context.Context, id string) (*domain.TemplateDefinition, error) {
	return m.store.Get(ctx, id)
}

func (m *TemplateManager) Create(ctx context.Context, tpl domain.TemplateDefinition) (*domain.TemplateDefinition, error) {
	if tpl.ID == "" {
		return nil, fmt.Errorf("template id required")
	}
	if tpl.Name == "" {
		return nil, fmt.Errorf("template name required")
	}
	if tpl.Canvas.Width == 0 || tpl.Canvas.Height == 0 {
		return nil, fmt.Errorf("canvas dimensions required")
	}

	now := time.Now().UTC()
	tpl.Version = 1
	tpl.CreatedAt = now
	tpl.UpdatedAt = now

	if err := m.store.Save(ctx, tpl); err != nil {
		return nil, fmt.Errorf("save template: %w", err)
	}
	return &tpl, nil
}

func (m *TemplateManager) Update(ctx context.Context, id string, tpl domain.TemplateDefinition) (*domain.TemplateDefinition, error) {
	existing, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	tpl.ID = id
	tpl.CreatedAt = existing.CreatedAt
	tpl.UpdatedAt = time.Now().UTC()
	tpl.Version = existing.Version + 1

	if err := m.store.Save(ctx, tpl); err != nil {
		return nil, fmt.Errorf("update template: %w", err)
	}
	return &tpl, nil
}

func (m *TemplateManager) Delete(ctx context.Context, id string) error {
	return m.store.Delete(ctx, id)
}

func (m *TemplateManager) Duplicate(ctx context.Context, id string) (*domain.TemplateDefinition, error) {
	src, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("source template not found: %w", err)
	}

	dup := *src
	dup.ID = id + "-copy"
	dup.Name = src.Name + " (Copy)"

	now := time.Now().UTC()
	dup.CreatedAt = now
	dup.UpdatedAt = now
	dup.Version = 1

	if err := m.store.Save(ctx, dup); err != nil {
		return nil, fmt.Errorf("save duplicate: %w", err)
	}
	return &dup, nil
}
