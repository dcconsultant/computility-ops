package api

import (
	"context"

	faapp "computility-ops/backend/internal/modules/failure-analytics/application"
	fadomain "computility-ops/backend/internal/modules/failure-analytics/domain"
)

// LegacyQueryAdapter is a compatibility layer for incremental handler migration.
type LegacyQueryAdapter struct {
	svc *faapp.Service
}

func NewLegacyQueryAdapter(svc *faapp.Service) *LegacyQueryAdapter {
	return &LegacyQueryAdapter{svc: svc}
}

func (a *LegacyQueryAdapter) ListOverallFailureRates(ctx context.Context) ([]fadomain.FailureRateSummary, error) {
	return a.svc.ListOverallFailureRates(ctx)
}

func (a *LegacyQueryAdapter) ListFailureOverviewCards(ctx context.Context) ([]fadomain.FailureOverviewCard, error) {
	return a.svc.ListFailureOverviewCards(ctx)
}

func (a *LegacyQueryAdapter) ListFailureAgeTrendPoints(ctx context.Context) ([]fadomain.FailureAgeTrendPoint, error) {
	return a.svc.ListFailureAgeTrendPoints(ctx)
}

func (a *LegacyQueryAdapter) ListFailureFeatureFacts(ctx context.Context) ([]fadomain.FailureFeatureFact, error) {
	return a.svc.ListFailureFeatureFacts(ctx)
}

func (a *LegacyQueryAdapter) ListStorageTopServerRates(ctx context.Context) ([]fadomain.StorageTopServerRate, error) {
	return a.svc.ListStorageTopServerRates(ctx)
}
