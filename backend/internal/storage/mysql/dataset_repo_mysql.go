package mysql

import (
	"context"
	"errors"

	"computility-ops/backend/internal/domain"
)

type DatasetRepo struct{}

func NewDatasetRepo(_ string) *DatasetRepo { return &DatasetRepo{} }

func (r *DatasetRepo) ReplaceHostPackages(_ context.Context, _ []domain.HostPackageConfig) error {
	return errors.New("mysql repo not implemented in phase 1")
}
func (r *DatasetRepo) ListHostPackages(_ context.Context) ([]domain.HostPackageConfig, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}

func (r *DatasetRepo) ReplaceSpecialRules(_ context.Context, _ []domain.SpecialRule) error {
	return errors.New("mysql repo not implemented in phase 1")
}
func (r *DatasetRepo) ListSpecialRules(_ context.Context) ([]domain.SpecialRule, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}

func (r *DatasetRepo) ReplaceModelFailureRates(_ context.Context, _ []domain.ModelFailureRate) error {
	return errors.New("mysql repo not implemented in phase 1")
}
func (r *DatasetRepo) ListModelFailureRates(_ context.Context) ([]domain.ModelFailureRate, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}

func (r *DatasetRepo) ReplacePackageFailureRates(_ context.Context, _ []domain.PackageFailureRate) error {
	return errors.New("mysql repo not implemented in phase 1")
}
func (r *DatasetRepo) ListPackageFailureRates(_ context.Context) ([]domain.PackageFailureRate, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}

func (r *DatasetRepo) ReplacePackageModelFailureRates(_ context.Context, _ []domain.PackageModelFailureRate) error {
	return errors.New("mysql repo not implemented in phase 1")
}
func (r *DatasetRepo) ListPackageModelFailureRates(_ context.Context) ([]domain.PackageModelFailureRate, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}

func (r *DatasetRepo) ReplaceOverallFailureRates(_ context.Context, _ []domain.FailureRateSummary) error {
	return errors.New("mysql repo not implemented in phase 1")
}

func (r *DatasetRepo) ListOverallFailureRates(_ context.Context) ([]domain.FailureRateSummary, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}
