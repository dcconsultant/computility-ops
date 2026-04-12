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
}
