package usecase

import (
	"context"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type DashboardService struct {
	interactions *InteractionService
	tplStore     port.TemplateStore
}

func NewDashboardService(interactions *InteractionService, tplStore port.TemplateStore) *DashboardService {
	return &DashboardService{interactions: interactions, tplStore: tplStore}
}

func (s *DashboardService) Overview(ctx context.Context) (domain.DashboardOverview, error) {
	overview := domain.DashboardOverview{
		TemplatePopularity: make(map[string]int),
		PalettePopularity:  make(map[string]int),
		FamilyUsage:        make(map[string]int),
	}

	if s.interactions != nil {
		stats, err := s.interactions.GetStats(ctx)
		if err == nil {
			overview.TotalDownloads = stats.TotalDownloads
			overview.TotalCardsGenerated = stats.TotalViews
			overview.TemplatePopularity = stats.TemplatePopularity
			overview.PalettePopularity = stats.PalettePopularity
		}
	}

	if s.tplStore != nil {
		tpls, err := s.tplStore.List(ctx)
		if err == nil {
			for _, t := range tpls {
				if t.Family != "" {
					overview.FamilyUsage[t.Family]++
				}
			}
		}
	}

	return overview, nil
}

func (s *DashboardService) InteractionFeed(ctx context.Context, limit int) ([]domain.CardInteraction, error) {
	if s.interactions == nil {
		return nil, nil
	}
	return s.interactions.GetRecent(ctx, limit)
}

type FamilyInfo struct {
	Name      string   `json:"name"`
	Templates []string `json:"templates"`
}

func (s *DashboardService) ListFamilies(ctx context.Context) ([]FamilyInfo, error) {
	if s.tplStore == nil {
		return nil, nil
	}
	tpls, err := s.tplStore.List(ctx)
	if err != nil {
		return nil, err
	}
	familyMap := make(map[string][]string)
	for _, t := range tpls {
		family := t.Family
		if family == "" {
			family = "uncategorized"
		}
		familyMap[family] = append(familyMap[family], t.ID)
	}
	result := make([]FamilyInfo, 0, len(familyMap))
	for name, templates := range familyMap {
		result = append(result, FamilyInfo{Name: name, Templates: templates})
	}
	return result, nil
}
