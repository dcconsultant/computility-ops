package memory

import (
	"context"
	"fmt"
	"sync"

	"computility-ops/backend/internal/domain"
)

type DatasetRepo struct {
	mu sync.RWMutex

	hostPackages        []domain.HostPackageConfig
	cabinetUtilization  domain.CabinetUtilizationSetting
	cabinetConfigs      []domain.CabinetConfig
	cabinetAutoID       int64
	specialRules        []domain.SpecialRule
	modelFailureRates   []domain.ModelFailureRate
	pkgFailureRates     []domain.PackageFailureRate
	pkgModelRates       []domain.PackageModelFailureRate
	overallRates        []domain.FailureRateSummary
	overviewCards       []domain.FailureOverviewCard
	ageTrendPoints      []domain.FailureAgeTrendPoint
	featureFacts        []domain.FailureFeatureFact
	storageTopRates     []domain.StorageTopServerRate
}

func NewDatasetRepo() *DatasetRepo {
	return &DatasetRepo{cabinetUtilization: domain.CabinetUtilizationSetting{Utilization: 1}}
}

func (r *DatasetRepo) ReplaceHostPackages(_ context.Context, rows []domain.HostPackageConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hostPackages = append([]domain.HostPackageConfig(nil), rows...)
	return nil
}
func (r *DatasetRepo) ListHostPackages(_ context.Context) ([]domain.HostPackageConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.HostPackageConfig(nil), r.hostPackages...), nil
}

func (r *DatasetRepo) GetCabinetUtilization(_ context.Context) (domain.CabinetUtilizationSetting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cabinetUtilization, nil
}

func (r *DatasetRepo) SetCabinetUtilization(_ context.Context, setting domain.CabinetUtilizationSetting) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cabinetUtilization = setting
	return nil
}

func (r *DatasetRepo) CreateCabinetConfig(_ context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.cabinetConfigs {
		if item.IDC == row.IDC && item.RatedPowerKW == row.RatedPowerKW {
			return domain.CabinetConfig{}, fmt.Errorf("duplicate key")
		}
	}
	r.cabinetAutoID++
	row.ID = r.cabinetAutoID
	r.cabinetConfigs = append(r.cabinetConfigs, row)
	return row, nil
}

func (r *DatasetRepo) UpdateCabinetConfig(_ context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.cabinetConfigs {
		if item.ID != row.ID && item.IDC == row.IDC && item.RatedPowerKW == row.RatedPowerKW {
			return domain.CabinetConfig{}, fmt.Errorf("duplicate key")
		}
	}
	for i := range r.cabinetConfigs {
		if r.cabinetConfigs[i].ID == row.ID {
			r.cabinetConfigs[i].IDC = row.IDC
			r.cabinetConfigs[i].RatedPowerKW = row.RatedPowerKW
			r.cabinetConfigs[i].MonthlyRent = row.MonthlyRent
			return r.cabinetConfigs[i], nil
		}
	}
	return domain.CabinetConfig{}, fmt.Errorf("not found")
}

func (r *DatasetRepo) DeleteCabinetConfig(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.CabinetConfig, 0, len(r.cabinetConfigs))
	for _, item := range r.cabinetConfigs {
		if item.ID != id {
			out = append(out, item)
		}
	}
	r.cabinetConfigs = out
	return nil
}

func (r *DatasetRepo) ListCabinetConfigs(_ context.Context) ([]domain.CabinetConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.CabinetConfig(nil), r.cabinetConfigs...), nil
}

func (r *DatasetRepo) ReplaceSpecialRules(_ context.Context, rows []domain.SpecialRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.specialRules = append([]domain.SpecialRule(nil), rows...)
	return nil
}
func (r *DatasetRepo) ListSpecialRules(_ context.Context) ([]domain.SpecialRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.SpecialRule(nil), r.specialRules...), nil
}

func (r *DatasetRepo) ReplaceModelFailureRates(_ context.Context, rows []domain.ModelFailureRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modelFailureRates = append([]domain.ModelFailureRate(nil), rows...)
	return nil
}
func (r *DatasetRepo) ListModelFailureRates(_ context.Context) ([]domain.ModelFailureRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.ModelFailureRate(nil), r.modelFailureRates...), nil
}

func (r *DatasetRepo) ReplacePackageFailureRates(_ context.Context, rows []domain.PackageFailureRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pkgFailureRates = append([]domain.PackageFailureRate(nil), rows...)
	return nil
}
func (r *DatasetRepo) ListPackageFailureRates(_ context.Context) ([]domain.PackageFailureRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.PackageFailureRate(nil), r.pkgFailureRates...), nil
}

func (r *DatasetRepo) ReplacePackageModelFailureRates(_ context.Context, rows []domain.PackageModelFailureRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pkgModelRates = append([]domain.PackageModelFailureRate(nil), rows...)
	return nil
}
func (r *DatasetRepo) ListPackageModelFailureRates(_ context.Context) ([]domain.PackageModelFailureRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.PackageModelFailureRate(nil), r.pkgModelRates...), nil
}

func (r *DatasetRepo) ReplaceOverallFailureRates(_ context.Context, rows []domain.FailureRateSummary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.overallRates = append([]domain.FailureRateSummary(nil), rows...)
	return nil
}

func (r *DatasetRepo) ListOverallFailureRates(_ context.Context) ([]domain.FailureRateSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.FailureRateSummary(nil), r.overallRates...), nil
}

func (r *DatasetRepo) ReplaceFailureOverviewCards(_ context.Context, rows []domain.FailureOverviewCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.overviewCards = append([]domain.FailureOverviewCard(nil), rows...)
	return nil
}

func (r *DatasetRepo) ListFailureOverviewCards(_ context.Context) ([]domain.FailureOverviewCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.FailureOverviewCard(nil), r.overviewCards...), nil
}

func (r *DatasetRepo) ReplaceFailureAgeTrendPoints(_ context.Context, rows []domain.FailureAgeTrendPoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ageTrendPoints = append([]domain.FailureAgeTrendPoint(nil), rows...)
	return nil
}

func (r *DatasetRepo) ListFailureAgeTrendPoints(_ context.Context) ([]domain.FailureAgeTrendPoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.FailureAgeTrendPoint(nil), r.ageTrendPoints...), nil
}

func (r *DatasetRepo) ReplaceFailureFeatureFacts(_ context.Context, rows []domain.FailureFeatureFact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.featureFacts = append([]domain.FailureFeatureFact(nil), rows...)
	return nil
}

func (r *DatasetRepo) ListFailureFeatureFacts(_ context.Context) ([]domain.FailureFeatureFact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.FailureFeatureFact(nil), r.featureFacts...), nil
}

func (r *DatasetRepo) ReplaceStorageTopServerRates(_ context.Context, rows []domain.StorageTopServerRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storageTopRates = append([]domain.StorageTopServerRate(nil), rows...)
	return nil
}

func (r *DatasetRepo) ListStorageTopServerRates(_ context.Context) ([]domain.StorageTopServerRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.StorageTopServerRate(nil), r.storageTopRates...), nil
}
