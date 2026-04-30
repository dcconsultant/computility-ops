package api

import (
	"context"

	assetapp "computility-ops/backend/internal/modules/asset-config/application"
)

// LegacyQueryAdapter is a Phase-1 placeholder used by legacy handlers.
// It allows incremental migration from internal/service to module application services.
type LegacyQueryAdapter struct {
	svc *assetapp.Service
}

func NewLegacyQueryAdapter(svc *assetapp.Service) *LegacyQueryAdapter {
	return &LegacyQueryAdapter{svc: svc}
}

func (a *LegacyQueryAdapter) ListServers(ctx context.Context) (int, error) {
	rows, err := a.svc.ListServers(ctx)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}
