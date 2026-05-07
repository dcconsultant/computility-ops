package repository

import (
	"context"

	"computility-ops/backend/internal/domain"
)

type ServerRepo interface {
	ReplaceAll(ctx context.Context, servers []domain.Server) error
	List(ctx context.Context) ([]domain.Server, error)
	Clear(ctx context.Context) error
}

type DatasetRepo interface {
	ReplaceHostPackages(ctx context.Context, rows []domain.HostPackageConfig) error
	ListHostPackages(ctx context.Context) ([]domain.HostPackageConfig, error)

	GetValueScoreCabinetBaseline(ctx context.Context) (domain.ValueScoreCabinetBaseline, error)

	GetCabinetUtilization(ctx context.Context) (domain.CabinetUtilizationSetting, error)
	SetCabinetUtilization(ctx context.Context, setting domain.CabinetUtilizationSetting) error
	GetValueScoreCostParams(ctx context.Context) (domain.ValueScoreCostParams, error)
	SetValueScoreCostParams(ctx context.Context, params domain.ValueScoreCostParams) error
	ReplaceValueScoreOriginalValues(ctx context.Context, rows []domain.ValueScoreOriginalValue) error
	ListValueScoreOriginalValues(ctx context.Context) ([]domain.ValueScoreOriginalValue, error)
	ReplaceValueScorePerformanceParams(ctx context.Context, rows []domain.ValueScorePerformanceParam) error
	ListValueScorePerformanceParams(ctx context.Context) ([]domain.ValueScorePerformanceParam, error)
	CreateCabinetConfig(ctx context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error)
	UpdateCabinetConfig(ctx context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error)
	DeleteCabinetConfig(ctx context.Context, id int64) error
	ListCabinetConfigs(ctx context.Context) ([]domain.CabinetConfig, error)

	ReplaceSpecialRules(ctx context.Context, rows []domain.SpecialRule) error
	ListSpecialRules(ctx context.Context) ([]domain.SpecialRule, error)

	ReplaceModelFailureRates(ctx context.Context, rows []domain.ModelFailureRate) error
	ListModelFailureRates(ctx context.Context) ([]domain.ModelFailureRate, error)

	ReplacePackageFailureRates(ctx context.Context, rows []domain.PackageFailureRate) error
	ListPackageFailureRates(ctx context.Context) ([]domain.PackageFailureRate, error)

	ReplacePackageModelFailureRates(ctx context.Context, rows []domain.PackageModelFailureRate) error
	ListPackageModelFailureRates(ctx context.Context) ([]domain.PackageModelFailureRate, error)

	ReplaceOverallFailureRates(ctx context.Context, rows []domain.FailureRateSummary) error
	ListOverallFailureRates(ctx context.Context) ([]domain.FailureRateSummary, error)

	ReplaceFailureOverviewCards(ctx context.Context, rows []domain.FailureOverviewCard) error
	ListFailureOverviewCards(ctx context.Context) ([]domain.FailureOverviewCard, error)

	ReplaceFailureAgeTrendPoints(ctx context.Context, rows []domain.FailureAgeTrendPoint) error
	ListFailureAgeTrendPoints(ctx context.Context) ([]domain.FailureAgeTrendPoint, error)

	ReplaceFailureFeatureFacts(ctx context.Context, rows []domain.FailureFeatureFact) error
	ListFailureFeatureFacts(ctx context.Context) ([]domain.FailureFeatureFact, error)

	ReplaceStorageTopServerRates(ctx context.Context, rows []domain.StorageTopServerRate) error
	ListStorageTopServerRates(ctx context.Context) ([]domain.StorageTopServerRate, error)
}
