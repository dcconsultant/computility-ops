package application

import (
	"context"
	"strings"

	legacy "computility-ops/backend/internal/domain"
	renewaldomain "computility-ops/backend/internal/modules/renewal/domain"
)

type Repository interface {
	ListPlans(ctx context.Context) ([]legacy.RenewalPlan, error)
	GetPlan(ctx context.Context, planID string) (legacy.RenewalPlan, error)
	GetSettings(ctx context.Context) (legacy.RenewalPlanSettings, bool, error)
	ListUnitPrices(ctx context.Context) ([]legacy.RenewalUnitPrice, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPlans(ctx context.Context) ([]renewaldomain.RenewalPlan, error) {
	rows, err := s.repo.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]renewaldomain.RenewalPlan, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapPlan(row))
	}
	return out, nil
}

func (s *Service) GetPlan(ctx context.Context, planID string) (renewaldomain.RenewalPlan, error) {
	row, err := s.repo.GetPlan(ctx, strings.TrimSpace(planID))
	if err != nil {
		return renewaldomain.RenewalPlan{}, err
	}
	return mapPlan(row), nil
}

func (s *Service) GetSettings(ctx context.Context) (renewaldomain.RenewalPlanSettings, bool, error) {
	settings, exists, err := s.repo.GetSettings(ctx)
	if err != nil {
		return renewaldomain.RenewalPlanSettings{}, false, err
	}
	if !exists {
		return renewaldomain.RenewalPlanSettings{}, false, nil
	}
	return mapSettings(settings), true, nil
}

func (s *Service) ListUnitPrices(ctx context.Context) ([]renewaldomain.RenewalUnitPrice, error) {
	rows, err := s.repo.ListUnitPrices(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]renewaldomain.RenewalUnitPrice, 0, len(rows))
	for _, row := range rows {
		out = append(out, renewaldomain.RenewalUnitPrice{
			Country:       row.Country,
			SceneCategory: row.SceneCategory,
			UnitPrice:     row.UnitPrice,
		})
	}
	return out, nil
}

func mapPlan(in legacy.RenewalPlan) renewaldomain.RenewalPlan {
	return renewaldomain.RenewalPlan{
		PlanID:               in.PlanID,
		TargetDate:           in.TargetDate,
		ExcludedEnvironments: append([]string{}, in.ExcludedEnvironments...),
		ExcludedPSAs:         append([]string{}, in.ExcludedPSAs...),
		TargetCores:          in.TargetCores,
		WarmTargetStorageTB:  in.WarmTargetStorageTB,
		HotTargetStorageTB:   in.HotTargetStorageTB,
		DomesticBudget:       in.DomesticBudget,
		IndiaBudget:          in.IndiaBudget,
		Requirements:         mapRequirements(in.Requirements),
		SelectedCores:        in.SelectedCores,
		SelectedStorageTB:    in.SelectedStorageTB,
		SelectedCount:        in.SelectedCount,
	}
}

func mapSettings(in legacy.RenewalPlanSettings) renewaldomain.RenewalPlanSettings {
	return renewaldomain.RenewalPlanSettings{
		TargetDate:           in.TargetDate,
		ExcludedEnvironments: append([]string{}, in.ExcludedEnvironments...),
		ExcludedPSAs:         append([]string{}, in.ExcludedPSAs...),
		Requirements:         mapRequirements(in.Requirements),
		DomesticBudget:       in.DomesticBudget,
		IndiaBudget:          in.IndiaBudget,
	}
}

func mapRequirements(in legacy.RenewalRequirements) renewaldomain.RenewalRequirements {
	return renewaldomain.RenewalRequirements{
		Domestic: mapRegionTargets(in.Domestic),
		India:    mapRegionTargets(in.India),
	}
}

func mapRegionTargets(in legacy.RenewalRegionTargets) renewaldomain.RenewalRegionTargets {
	return renewaldomain.RenewalRegionTargets{
		Compute:     mapSceneTarget(in.Compute),
		WarmStorage: mapSceneTarget(in.WarmStorage),
		HotStorage:  mapSceneTarget(in.HotStorage),
		GPU:         mapSceneTarget(in.GPU),
	}
}

func mapSceneTarget(in legacy.RenewalSceneTarget) renewaldomain.RenewalSceneTarget {
	return renewaldomain.RenewalSceneTarget{Mode: renewaldomain.RenewalTargetMode(in.Mode), Target: in.Target}
}
