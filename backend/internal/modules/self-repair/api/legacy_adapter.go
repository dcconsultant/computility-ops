package api

import (
	"context"

	srapp "computility-ops/backend/internal/modules/self-repair/application"
	srdomain "computility-ops/backend/internal/modules/self-repair/domain"
)

type LegacyQueryAdapter struct{ svc *srapp.Service }

func NewLegacyQueryAdapter(svc *srapp.Service) *LegacyQueryAdapter { return &LegacyQueryAdapter{svc: svc} }

func (a *LegacyQueryAdapter) ListSuggestions(ctx context.Context) ([]srdomain.SelfRepairSuggestion, error) {
	return a.svc.ListSuggestions(ctx)
}
