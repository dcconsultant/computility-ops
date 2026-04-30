package tests

import (
	"context"
	"testing"

	"computility-ops/backend/internal/domain"
	assetapp "computility-ops/backend/internal/modules/asset-config/application"
)

type fakeServerRepo struct{}

type fakeDatasetRepo struct{}

func (fakeServerRepo) List(ctx context.Context) ([]domain.Server, error) {
	_ = ctx
	return []domain.Server{{SN: "S1", Manufacturer: "A", ConfigType: "C1"}}, nil
}

func (fakeDatasetRepo) ListHostPackages(ctx context.Context) ([]domain.HostPackageConfig, error) {
	_ = ctx
	return []domain.HostPackageConfig{{ConfigType: "C1", CPULogicalCores: 64, ArchStandardizedFactor: 1}}, nil
}

func (fakeDatasetRepo) ListSpecialRules(ctx context.Context) ([]domain.SpecialRule, error) {
	_ = ctx
	return []domain.SpecialRule{{SN: "S1", Policy: "whitelist"}}, nil
}

func TestListServers(t *testing.T) {
	svc := assetapp.NewService(fakeServerRepo{}, fakeDatasetRepo{})
	rows, err := svc.ListServers(context.Background())
	if err != nil {
		t.Fatalf("ListServers err=%v", err)
	}
	if len(rows) != 1 || rows[0].AssetID != "S1" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}
