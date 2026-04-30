package application

import (
	"context"

	legacy "computility-ops/backend/internal/domain"
	fadomain "computility-ops/backend/internal/modules/failure-analytics/domain"
)

type DatasetReader interface {
	ListOverallFailureRates(ctx context.Context) ([]legacy.FailureRateSummary, error)
	ListFailureOverviewCards(ctx context.Context) ([]legacy.FailureOverviewCard, error)
	ListFailureAgeTrendPoints(ctx context.Context) ([]legacy.FailureAgeTrendPoint, error)
	ListFailureFeatureFacts(ctx context.Context) ([]legacy.FailureFeatureFact, error)
	ListStorageTopServerRates(ctx context.Context) ([]legacy.StorageTopServerRate, error)
}

type Service struct {
	datasets DatasetReader
}

func NewService(datasets DatasetReader) *Service {
	return &Service{datasets: datasets}
}

func (s *Service) ListOverallFailureRates(ctx context.Context) ([]fadomain.FailureRateSummary, error) {
	rows, err := s.datasets.ListOverallFailureRates(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fadomain.FailureRateSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, fadomain.FailureRateSummary(row))
	}
	return out, nil
}

func (s *Service) ListFailureOverviewCards(ctx context.Context) ([]fadomain.FailureOverviewCard, error) {
	rows, err := s.datasets.ListFailureOverviewCards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fadomain.FailureOverviewCard, 0, len(rows))
	for _, row := range rows {
		out = append(out, fadomain.FailureOverviewCard(row))
	}
	return out, nil
}

func (s *Service) ListFailureAgeTrendPoints(ctx context.Context) ([]fadomain.FailureAgeTrendPoint, error) {
	rows, err := s.datasets.ListFailureAgeTrendPoints(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fadomain.FailureAgeTrendPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, fadomain.FailureAgeTrendPoint(row))
	}
	return out, nil
}

func (s *Service) ListFailureFeatureFacts(ctx context.Context) ([]fadomain.FailureFeatureFact, error) {
	rows, err := s.datasets.ListFailureFeatureFacts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fadomain.FailureFeatureFact, 0, len(rows))
	for _, row := range rows {
		out = append(out, fadomain.FailureFeatureFact(row))
	}
	return out, nil
}

func (s *Service) ListStorageTopServerRates(ctx context.Context) ([]fadomain.StorageTopServerRate, error) {
	rows, err := s.datasets.ListStorageTopServerRates(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fadomain.StorageTopServerRate, 0, len(rows))
	for _, row := range rows {
		out = append(out, fadomain.StorageTopServerRate(row))
	}
	return out, nil
}
