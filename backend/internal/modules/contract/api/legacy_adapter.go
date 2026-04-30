package api

import (
	"context"
	"strings"

	contractapp "computility-ops/backend/internal/modules/contract/application"
	contractdomain "computility-ops/backend/internal/modules/contract/domain"
)

// LegacyQueryAdapter is a Phase-1/2 compatibility adapter for old handlers.
type LegacyQueryAdapter struct {
	svc *contractapp.Service
}

func NewLegacyQueryAdapter(svc *contractapp.Service) *LegacyQueryAdapter {
	return &LegacyQueryAdapter{svc: svc}
}

func (a *LegacyQueryAdapter) ListContracts(ctx context.Context) ([]contractdomain.Contract, error) {
	return a.svc.ListContracts(ctx)
}

func (a *LegacyQueryAdapter) GetContract(ctx context.Context, contractID string) (contractdomain.Contract, error) {
	return a.svc.GetContract(ctx, strings.TrimSpace(contractID))
}
