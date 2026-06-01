package service

import (
	"context"
	"math"
	"testing"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/storage/memory"
)

func TestValueScoreSetupService_CalculateTableItems_MatchesFrontendRows(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDatasetRepo()
	svc := NewValueScoreSetupService(repo, nil)

	if err := repo.SetValueScoreCostParams(ctx, domain.ValueScoreCostParams{
		DepreciationMonths:    60,
		NetworkDeviceShareCNY: 10,
		ServerRenewalFeeCNY:   20,
		CabinetUtilization:    1,
		RatedPowerKW:          10,
		MonthlyRentCNY:        1000,
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{
		{ConfigType: "compute-a", SceneCategory: "计算", CPULogicalCores: 8, MemoryCapacityGB: 64, PowerWatts: 1000, ReleaseYear: 0, ArchStandardizedFactor: 1},
		{ConfigType: "gpu-a", SceneCategory: "GPU", CPULogicalCores: 16, MemoryCapacityGB: 128, GPUCardCount: 4, PowerWatts: 1000, ReleaseYear: 0, ArchStandardizedFactor: 1},
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceValueScoreOriginalValues(ctx, []domain.ValueScoreOriginalValue{
		{ConfigType: "compute-a", ServerOriginalCNY: 60000},
		{ConfigType: "gpu-a", ServerOriginalCNY: 60000},
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceValueScorePerformanceParams(ctx, []domain.ValueScorePerformanceParam{
		{ConfigType: "compute-a", PerformanceScore: 2277},
		{ConfigType: "gpu-a", PerformanceScore: 2277},
	}); err != nil {
		t.Fatal(err)
	}

	items, err := svc.CalculateTableItems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items)=%d, want 2", len(items))
	}

	compute := items[0]
	if compute.ConfigType != "compute-a" {
		t.Fatalf("first config_type=%s, want compute-a", compute.ConfigType)
	}
	if compute.SceneType != "compute" {
		t.Fatalf("compute SceneType=%s, want compute", compute.SceneType)
	}
	if !closeEnough(compute.UnitTCO, 137.5) {
		t.Fatalf("compute UnitTCO=%f, want 137.5", compute.UnitTCO)
	}
	if !closeEnough(compute.ValueScoreV1, 4.5833) {
		t.Fatalf("compute ValueScoreV1=%f, want 4.5833", compute.ValueScoreV1)
	}

	gpu := items[1]
	if gpu.SceneType != "gpu" {
		t.Fatalf("gpu SceneType=%s, want gpu", gpu.SceneType)
	}
	if !closeEnough(gpu.UnitTCO, 275) {
		t.Fatalf("gpu UnitTCO=%f, want 275", gpu.UnitTCO)
	}
	if !closeEnough(gpu.ValueScoreV1, 275) {
		t.Fatalf("gpu ValueScoreV1=%f, want 275", gpu.ValueScoreV1)
	}
}

func closeEnough(got, want float64) bool {
	return math.Abs(got-want) < 0.0001
}
