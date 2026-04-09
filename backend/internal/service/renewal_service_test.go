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
		{SN: "A", ConfigType: "c1", PSA: 10},
		{SN: "B", ConfigType: "c1", PSA: 20},
		{SN: "C", ConfigType: "c1", PSA: 15},
	})
	_ = datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{ConfigType: "c1", CPULogicalCores: 8, ArchStandardizedFactor: 1}})
	_ = datasetRepo.ReplaceSpecialRules(ctx, []domain.SpecialRule{{SN: "A", Policy: "whitelist"}, {SN: "C", Policy: "blacklist"}})

	svc := NewRenewalService(serverRepo, datasetRepo, renewalRepo)
	plan, err := svc.CreatePlan(ctx, 16)
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

func TestRenewalService_CreatePlan_InvalidTarget(t *testing.T) {
	svc := NewRenewalService(mem.NewServerRepo(), mem.NewDatasetRepo(), mem.NewRenewalRepo())
	_, err := svc.CreatePlan(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for target_cores <= 0")
	}
}
