package tests

import (
	"context"
	"testing"

	masterapp "computility-ops/backend/internal/modules/master-data/application"
	masterdomain "computility-ops/backend/internal/modules/master-data/domain"
	masterinfra "computility-ops/backend/internal/modules/master-data/infrastructure"
)

func TestListAssets_NormalizesAssetID(t *testing.T) {
	repo := masterinfra.NewMemoryRepository([]masterdomain.Asset{{SN: "SN-1"}})
	svc := masterapp.NewService(repo)

	assets, err := svc.ListAssets(context.Background())
	if err != nil {
		t.Fatalf("ListAssets error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	if assets[0].AssetID != "SN-1" {
		t.Fatalf("expected AssetID normalized to SN-1, got %q", assets[0].AssetID)
	}
}
