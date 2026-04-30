package api

import (
	"context"
	"strings"

	renewalapp "computility-ops/backend/internal/modules/renewal/application"
	renewaldomain "computility-ops/backend/internal/modules/renewal/domain"
)

// LegacyQueryAdapter is a compatibility adapter for incremental renewal read-path migration.
type LegacyQueryAdapter struct {
	svc *renewalapp.Service
}

func NewLegacyQueryAdapter(svc *renewalapp.Service) *LegacyQueryAdapter {
	return &LegacyQueryAdapter{svc: svc}
}

func (a *LegacyQueryAdapter) ListPlans(ctx context.Context) ([]renewaldomain.RenewalPlan, error) {
	return a.svc.ListPlans(ctx)
}

func (a *LegacyQueryAdapter) GetPlan(ctx context.Context, planID string) (renewaldomain.RenewalPlan, error) {
	return a.svc.GetPlan(ctx, strings.TrimSpace(planID))
}

func (a *LegacyQueryAdapter) GetSettings(ctx context.Context) (renewaldomain.RenewalPlanSettings, bool, error) {
	return a.svc.GetSettings(ctx)
}

func (a *LegacyQueryAdapter) ListUnitPrices(ctx context.Context) ([]renewaldomain.RenewalUnitPrice, error) {
	return a.svc.ListUnitPrices(ctx)
}
