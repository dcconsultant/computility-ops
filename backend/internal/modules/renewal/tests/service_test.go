package tests

import (
	"context"
	"testing"

	legacy "computility-ops/backend/internal/domain"
	renewalapp "computility-ops/backend/internal/modules/renewal/application"
)

type fakeRepo struct{}

func (fakeRepo) ListPlans(ctx context.Context) ([]legacy.RenewalPlan, error) {
	_ = ctx
	return []legacy.RenewalPlan{{PlanID: "P1", TargetDate: "2026-12-31", TargetCores: 100}}, nil
}

func (fakeRepo) GetPlan(ctx context.Context, planID string) (legacy.RenewalPlan, error) {
	_ = ctx
	return legacy.RenewalPlan{PlanID: planID, TargetDate: "2026-12-31"}, nil
}

func (fakeRepo) GetSettings(ctx context.Context) (legacy.RenewalPlanSettings, bool, error) {
	_ = ctx
	return legacy.RenewalPlanSettings{TargetDate: "2026-12-31"}, true, nil
}

func (fakeRepo) ListUnitPrices(ctx context.Context) ([]legacy.RenewalUnitPrice, error) {
	_ = ctx
	return []legacy.RenewalUnitPrice{{Country: "国内", SceneCategory: "compute", UnitPrice: 100}}, nil
}

func TestListPlans(t *testing.T) {
	svc := renewalapp.NewService(fakeRepo{})
	rows, err := svc.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans err=%v", err)
	}
	if len(rows) != 1 || rows[0].PlanID != "P1" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestGetPlan(t *testing.T) {
	svc := renewalapp.NewService(fakeRepo{})
	row, err := svc.GetPlan(context.Background(), " P2 ")
	if err != nil {
		t.Fatalf("GetPlan err=%v", err)
	}
	if row.PlanID != "P2" {
		t.Fatalf("expected trimmed plan id P2, got %s", row.PlanID)
	}
}
