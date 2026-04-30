package tests

import (
	"context"
	"testing"

	legacy "computility-ops/backend/internal/domain"
	faapp "computility-ops/backend/internal/modules/failure-analytics/application"
)

type fakeDatasetRepo struct{}

func (fakeDatasetRepo) ListOverallFailureRates(ctx context.Context) ([]legacy.FailureRateSummary, error) {
	_ = ctx
	return []legacy.FailureRateSummary{{Period: "history", Segment: "all", FaultCount: 10}}, nil
}
func (fakeDatasetRepo) ListFailureOverviewCards(ctx context.Context) ([]legacy.FailureOverviewCard, error) {
	_ = ctx
	return []legacy.FailureOverviewCard{{Segment: "all", Year: 2026}}, nil
}
func (fakeDatasetRepo) ListFailureAgeTrendPoints(ctx context.Context) ([]legacy.FailureAgeTrendPoint, error) {
	_ = ctx
	return []legacy.FailureAgeTrendPoint{{Segment: "all", AgeBucket: 2}}, nil
}
func (fakeDatasetRepo) ListFailureFeatureFacts(ctx context.Context) ([]legacy.FailureFeatureFact, error) {
	_ = ctx
	return []legacy.FailureFeatureFact{{RecordYearIndex: 1, Scope: "product"}}, nil
}
func (fakeDatasetRepo) ListStorageTopServerRates(ctx context.Context) ([]legacy.StorageTopServerRate, error) {
	_ = ctx
	return []legacy.StorageTopServerRate{{SN: "S1", FaultRate: 0.1}}, nil
}

func TestListOverallFailureRates(t *testing.T) {
	svc := faapp.NewService(fakeDatasetRepo{})
	rows, err := svc.ListOverallFailureRates(context.Background())
	if err != nil {
		t.Fatalf("ListOverallFailureRates err=%v", err)
	}
	if len(rows) != 1 || rows[0].FaultCount != 10 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}
