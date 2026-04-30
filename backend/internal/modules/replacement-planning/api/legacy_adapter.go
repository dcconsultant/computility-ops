package api

import (
	"context"

	rpapp "computility-ops/backend/internal/modules/replacement-planning/application"
	rpdomain "computility-ops/backend/internal/modules/replacement-planning/domain"
)

type LegacyQueryAdapter struct{ svc *rpapp.Service }

func NewLegacyQueryAdapter(svc *rpapp.Service) *LegacyQueryAdapter { return &LegacyQueryAdapter{svc: svc} }

func (a *LegacyQueryAdapter) ListSuggestions(ctx context.Context) ([]rpdomain.ReplacementSuggestion, error) {
	return a.svc.ListSuggestions(ctx)
}
