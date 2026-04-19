package service

import (
	"context"
	"testing"

	"computility-ops/backend/internal/domain"
	mem "computility-ops/backend/internal/storage/memory"
)

func TestRenewalService_CreatePlan_SelectsWhitelistAndSortsByScore(t *testing.T) {
	ctx := context.Background()

	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	_ = serverRepo.ReplaceAll(ctx, []domain.Server{
		{SN: "A", ConfigType: "c1", PSA: "10", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "B", ConfigType: "c1", PSA: "20", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "C", ConfigType: "c1", PSA: "15", WarrantyEndDate: "2025-01-01", Environment: "生产"},
	})
	_ = datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{ConfigType: "c1", SceneCategory: "计算型", CPULogicalCores: 8, ArchStandardizedFactor: 1}})
	_ = datasetRepo.ReplaceSpecialRules(ctx, []domain.SpecialRule{{SN: "A", Policy: "whitelist"}, {SN: "C", Policy: "blacklist"}})

	svc := NewRenewalService(serverRepo, datasetRepo, renewalRepo)
	plan, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", ExcludedEnvironments: []string{"开发", "测试"}, TargetCores: 16, WarmTargetStorageTB: 0, HotTargetStorageTB: 0})
	if err != nil {
		t.Fatalf("CreatePlan() error = %v", err)
	}

	if plan.SelectedCount != 2 {
		t.Fatalf("SelectedCount=%d, want 2", plan.SelectedCount)
	}
	if plan.SelectedCores != 16 {
		t.Fatalf("SelectedCores=%d, want 16", plan.SelectedCores)
	}

	if len(plan.Items) != 2 {
		t.Fatalf("len(Items)=%d, want 2", len(plan.Items))
	}
	if plan.Items[0].SN != "A" || plan.Items[0].SpecialPolicy != "whitelist" {
		t.Fatalf("first item = %+v, want SN A as whitelist", plan.Items[0])
	}
	if plan.Items[1].SN != "B" {
		t.Fatalf("second item SN=%s, want B", plan.Items[1].SN)
	}
}

func TestRenewalService_CreatePlan_TargetIncludesUnexpiredButOutputExcludesUnexpired(t *testing.T) {
	ctx := context.Background()
	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	_ = serverRepo.ReplaceAll(ctx, []domain.Server{
		{SN: "U1", ConfigType: "compute-a", PSA: "10", WarrantyEndDate: "2026-12-31", Environment: "生产"}, // 未过保，计入覆盖
		{SN: "E1", ConfigType: "compute-a", PSA: "9", WarrantyEndDate: "2025-01-01", Environment: "生产"},  // 过保，候选
		{SN: "E2", ConfigType: "compute-a", PSA: "8", WarrantyEndDate: "2025-01-01", Environment: "生产"},  // 过保，候选
	})
	_ = datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{ConfigType: "compute-a", SceneCategory: "计算型", CPULogicalCores: 8, ArchStandardizedFactor: 1}})

	svc := NewRenewalService(serverRepo, datasetRepo, renewalRepo)
	plan, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 16, WarmTargetStorageTB: 0, HotTargetStorageTB: 0})
	if err != nil {
		t.Fatalf("CreatePlan() error = %v", err)
	}

	if plan.CoveredComputeCores != 8 {
		t.Fatalf("CoveredComputeCores=%d, want 8", plan.CoveredComputeCores)
	}
	if plan.RequiredComputeCores != 8 {
		t.Fatalf("RequiredComputeCores=%d, want 8", plan.RequiredComputeCores)
	}
	if len(plan.Items) != 1 || plan.Items[0].SN != "E1" {
		t.Fatalf("selected items=%+v, want only E1", plan.Items)
	}
}

func TestRenewalService_CreatePlan_ExcludePSA(t *testing.T) {
	ctx := context.Background()
	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	_ = serverRepo.ReplaceAll(ctx, []domain.Server{
		{SN: "A", ConfigType: "c1", PSA: "P0", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "B", ConfigType: "c1", PSA: "20", WarrantyEndDate: "2025-01-01", Environment: "生产"},
	})
	_ = datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{ConfigType: "c1", SceneCategory: "计算型", CPULogicalCores: 8, ArchStandardizedFactor: 1}})

	svc := NewRenewalService(serverRepo, datasetRepo, renewalRepo)
	plan, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 8, ExcludedPSAs: []string{"P0"}})
	if err != nil {
		t.Fatalf("CreatePlan() error = %v", err)
	}
	if len(plan.Items) != 1 || plan.Items[0].SN != "B" {
		t.Fatalf("exclude psa failed, items=%+v", plan.Items)
	}
}

func TestRenewalService_CreatePlan_ExcludePSA_PathSegmentPrefix(t *testing.T) {
	ctx := context.Background()
	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	_ = serverRepo.ReplaceAll(ctx, []domain.Server{
		{SN: "A", ConfigType: "c1", PSA: "/aa/ss/as", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "B", ConfigType: "c1", PSA: "/aa/ss/as,/aa/ss/at", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "C", ConfigType: "c1", PSA: "/bb/xx", WarrantyEndDate: "2025-01-01", Environment: "生产"},
	})
	_ = datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{ConfigType: "c1", SceneCategory: "计算型", CPULogicalCores: 8, ArchStandardizedFactor: 1}})

	svc := NewRenewalService(serverRepo, datasetRepo, renewalRepo)

	planPrefix, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 8, ExcludedPSAs: []string{"/aa"}})
	if err != nil {
		t.Fatalf("CreatePlan(prefix) error = %v", err)
	}
	if len(planPrefix.Items) != 1 || planPrefix.Items[0].SN != "C" {
		t.Fatalf("prefix exclude failed, items=%+v", planPrefix.Items)
	}

	planExact, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 16, ExcludedPSAs: []string{"/aa/ss/at"}})
	if err != nil {
		t.Fatalf("CreatePlan(exact) error = %v", err)
	}
	if len(planExact.Items) != 2 || planExact.Items[0].SN != "A" || planExact.Items[1].SN != "C" {
		t.Fatalf("exact exclude failed, items=%+v", planExact.Items)
	}
}

func TestRenewalService_CreatePlan_SkipUnmatchedConfigType(t *testing.T) {
	ctx := context.Background()
	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	_ = serverRepo.ReplaceAll(ctx, []domain.Server{
		{SN: "A", ConfigType: "known", PSA: "10", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "B", ConfigType: "unknown", PSA: "10", WarrantyEndDate: "2025-01-01", Environment: "生产"},
	})
	_ = datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{ConfigType: "known", SceneCategory: "计算型", CPULogicalCores: 8, ArchStandardizedFactor: 1}})

	svc := NewRenewalService(serverRepo, datasetRepo, renewalRepo)
	plan, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 8})
	if err != nil {
		t.Fatalf("CreatePlan() error = %v", err)
	}
	if plan.UnmatchedConfigCount != 1 {
		t.Fatalf("UnmatchedConfigCount=%d, want 1", plan.UnmatchedConfigCount)
	}
	if len(plan.UnmatchedConfigTypes) != 1 || plan.UnmatchedConfigTypes[0] != "unknown" {
		t.Fatalf("UnmatchedConfigTypes=%v, want [unknown]", plan.UnmatchedConfigTypes)
	}
	if len(plan.Items) != 1 || plan.Items[0].SN != "A" {
		t.Fatalf("items=%+v, want only SN A", plan.Items)
	}
}

func TestRenewalService_CreatePlan_ExcludePSA_SegmentBoundary(t *testing.T) {
	ctx := context.Background()
	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	_ = serverRepo.ReplaceAll(ctx, []domain.Server{
		{SN: "AA", ConfigType: "c1", PSA: "/aa", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "AB", ConfigType: "c1", PSA: "/ab", WarrantyEndDate: "2025-01-01", Environment: "生产"},
		{SN: "SSA", ConfigType: "c1", PSA: "/ss/st/a", WarrantyEndDate: "2025-01-01", Environment: "生产"},
	})
	_ = datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{ConfigType: "c1", SceneCategory: "计算型", CPULogicalCores: 8, ArchStandardizedFactor: 1}})

	svc := NewRenewalService(serverRepo, datasetRepo, renewalRepo)

	planA, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 24, ExcludedPSAs: []string{"/a"}})
	if err != nil {
		t.Fatalf("CreatePlan(/a) error = %v", err)
	}
	if len(planA.Items) != 3 {
		t.Fatalf("/a should not match /aa,/ab,/ss/st/a, items=%+v", planA.Items)
	}

	planSS, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 16, ExcludedPSAs: []string{"/ss"}})
	if err != nil {
		t.Fatalf("CreatePlan(/ss) error = %v", err)
	}
	if len(planSS.Items) != 2 || planSS.Items[0].SN != "AA" || planSS.Items[1].SN != "AB" {
		t.Fatalf("/ss should match /ss/st/a only, items=%+v", planSS.Items)
	}

	planSSST, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 16, ExcludedPSAs: []string{"/ss/st"}})
	if err != nil {
		t.Fatalf("CreatePlan(/ss/st) error = %v", err)
	}
	if len(planSSST.Items) != 2 || planSSST.Items[0].SN != "AA" || planSSST.Items[1].SN != "AB" {
		t.Fatalf("/ss/st should match /ss/st/a, items=%+v", planSSST.Items)
	}

	planExact, err := svc.CreatePlan(ctx, CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 16, ExcludedPSAs: []string{"/ss/st/a"}})
	if err != nil {
		t.Fatalf("CreatePlan(/ss/st/a) error = %v", err)
	}
	if len(planExact.Items) != 2 || planExact.Items[0].SN != "AA" || planExact.Items[1].SN != "AB" {
		t.Fatalf("/ss/st/a should match itself, items=%+v", planExact.Items)
	}
}

func TestRenewalService_CreatePlan_InvalidTarget(t *testing.T) {
	svc := NewRenewalService(mem.NewServerRepo(), mem.NewDatasetRepo(), mem.NewRenewalRepo())
	_, err := svc.CreatePlan(context.Background(), CreatePlanInput{TargetDate: "2026-01-01", TargetCores: 0})
	if err == nil {
		t.Fatal("expected error for target_cores <= 0")
	}
}
