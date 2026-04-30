package application

import (
	"context"
	"strings"

	masterdomain "computility-ops/backend/internal/modules/master-data/domain"
)

// Repository defines the persistence contract for master-data assets.
// NOTE: Phase 1 keeps this inside the module and avoids cross-module DB reads.
type Repository interface {
	ListAssets(ctx context.Context) ([]masterdomain.Asset, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListAssets(ctx context.Context) ([]masterdomain.Asset, error) {
	assets, err := s.repo.ListAssets(ctx)
	if err != nil {
		return nil, err
	}
	for i := range assets {
		assets[i].AssetID = normalizeAssetID(assets[i].AssetID, assets[i].SN)
	}
	return assets, nil
}

func normalizeAssetID(assetID, sn string) string {
	assetID = strings.TrimSpace(assetID)
	if assetID != "" {
		return assetID
	}
	return strings.TrimSpace(sn)
}
