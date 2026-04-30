package api

import (
	"context"

	rcapp "computility-ops/backend/internal/modules/reconfig-planning/application"
	rcdomain "computility-ops/backend/internal/modules/reconfig-planning/domain"
)

type LegacyQueryAdapter struct{ svc *rcapp.Service }

func NewLegacyQueryAdapter(svc *rcapp.Service) *LegacyQueryAdapter { return &LegacyQueryAdapter{svc: svc} }

func (a *LegacyQueryAdapter) ListSuggestions(ctx context.Context) ([]rcdomain.ReconfigSuggestion, error) {
	return a.svc.ListSuggestions(ctx)
}
