package infrastructure

import (
	"context"
	"testing"

	"computility-ops/backend/internal/domain"
)

type fakeServerRepo struct{}
type fakeDatasetRepo struct{}

func (fakeServerRepo) ReplaceAll(ctx context.Context, servers []domain.Server) error { _ = ctx; _ = servers; return nil }
func (fakeServerRepo) Clear(ctx context.Context) error { _ = ctx; return nil }
func (fakeServerRepo) List(ctx context.Context) ([]domain.Server, error) {
	_ = ctx
	return []domain.Server{{SN: "S1", Model: "GPU-X", ConfigType: "C1", LaunchDate: "2020-01-01"}}, nil
}

func (fakeDatasetRepo) ReplaceHostPackages(ctx context.Context, rows []domain.HostPackageConfig) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListHostPackages(ctx context.Context) ([]domain.HostPackageConfig, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplaceSpecialRules(ctx context.Context, rows []domain.SpecialRule) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListSpecialRules(ctx context.Context) ([]domain.SpecialRule, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplaceModelFailureRates(ctx context.Context, rows []domain.ModelFailureRate) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListModelFailureRates(ctx context.Context) ([]domain.ModelFailureRate, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplacePackageFailureRates(ctx context.Context, rows []domain.PackageFailureRate) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListPackageFailureRates(ctx context.Context) ([]domain.PackageFailureRate, error) {
	_ = ctx
	return []domain.PackageFailureRate{{ConfigType: "C1", Recent1YFailureRate: 0.2}}, nil
}
func (fakeDatasetRepo) ReplacePackageModelFailureRates(ctx context.Context, rows []domain.PackageModelFailureRate) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListPackageModelFailureRates(ctx context.Context) ([]domain.PackageModelFailureRate, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplaceOverallFailureRates(ctx context.Context, rows []domain.FailureRateSummary) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListOverallFailureRates(ctx context.Context) ([]domain.FailureRateSummary, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplaceFailureOverviewCards(ctx context.Context, rows []domain.FailureOverviewCard) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListFailureOverviewCards(ctx context.Context) ([]domain.FailureOverviewCard, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplaceFailureAgeTrendPoints(ctx context.Context, rows []domain.FailureAgeTrendPoint) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListFailureAgeTrendPoints(ctx context.Context) ([]domain.FailureAgeTrendPoint, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplaceFailureFeatureFacts(ctx context.Context, rows []domain.FailureFeatureFact) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListFailureFeatureFacts(ctx context.Context) ([]domain.FailureFeatureFact, error) { _ = ctx; return nil, nil }
func (fakeDatasetRepo) ReplaceStorageTopServerRates(ctx context.Context, rows []domain.StorageTopServerRate) error { _ = ctx; _ = rows; return nil }
func (fakeDatasetRepo) ListStorageTopServerRates(ctx context.Context) ([]domain.StorageTopServerRate, error) { _ = ctx; return nil, nil }

func TestLegacyReader_ListCandidates(t *testing.T) {
	r := NewLegacyReader(fakeServerRepo{}, fakeDatasetRepo{})
	rows, err := r.ListCandidates(context.Background())
	if err != nil {
		t.Fatalf("ListCandidates err=%v", err)
	}
	if len(rows) != 1 || rows[0].AssetID != "S1" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
	if rows[0].AnnualFailureRate <= 0 {
		t.Fatalf("expected positive AFR: %+v", rows[0])
	}
}
