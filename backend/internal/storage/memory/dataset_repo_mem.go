package memory

import (
	"context"
	"sync"

	"computility-ops/backend/internal/domain"
)

type DatasetRepo struct {
	mu sync.RWMutex

	hostPackages      []domain.HostPackageConfig
	specialRules      []domain.SpecialRule
	modelFailureRates []domain.ModelFailureRate
	pkgFailureRates   []domain.PackageFailureRate
	pkgModelRates     []domain.PackageModelFailureRate
	overallRates      []domain.FailureRateSummary
}

func NewDatasetRepo() *DatasetRepo { return &DatasetRepo{} }

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
