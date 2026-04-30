package infrastructure

import (
	"context"

	masterapp "computility-ops/backend/internal/modules/master-data/application"
	masterdomain "computility-ops/backend/internal/modules/master-data/domain"
)

var _ masterapp.Repository = (*MemoryRepository)(nil)

type MemoryRepository struct {
	assets []masterdomain.Asset
}

func NewMemoryRepository(seed []masterdomain.Asset) *MemoryRepository {
	copied := make([]masterdomain.Asset, len(seed))
	copy(copied, seed)
	return &MemoryRepository{assets: copied}
}

func (r *MemoryRepository) ListAssets(ctx context.Context) ([]masterdomain.Asset, error) {
	_ = ctx
	out := make([]masterdomain.Asset, len(r.assets))
	copy(out, r.assets)
	return out, nil
}
