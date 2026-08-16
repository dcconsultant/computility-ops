package service

import (
	"context"
	"math"
	"testing"

	"computility-ops/backend/internal/domain"
)

func TestDeliveryDecisionService_Calculate_DefaultChina(t *testing.T) {
	svc := NewDeliveryDecisionService()
	defaults := svc.Defaults(context.Background(), "China")
	res, err := svc.Calculate(context.Background(), domain.DeliveryDecisionCalculateRequest{
		Input: defaults.Input,
	})
	if err != nil {
		t.Fatalf("calculate error: %v", err)
	}

	if !deliveryDecisionCloseEnough(res.Formula.CloudDailyNet, 1.5877641509433962) {
		t.Fatalf("cloud_daily_net=%f", res.Formula.CloudDailyNet)
	}
	if !deliveryDecisionCloseEnough(res.Formula.SelfDailyTCO, 0.9564102101782064) {
		t.Fatalf("self_daily_tco=%f", res.Formula.SelfDailyTCO)
	}
	if !deliveryDecisionCloseEnough(res.Formula.SelfDailyTCO3Y, 1.9058398668754106) {
		t.Fatalf("self_daily_tco_3y=%f", res.Formula.SelfDailyTCO3Y)
	}
	if res.Formula.CloudHedgeLost {
		t.Fatal("expected hedge not lost for default China case")
	}
	if len(res.SensitivityPoints) != 11 {
		t.Fatalf("sensitivity points=%d, want 11", len(res.SensitivityPoints))
	}
}

func TestDeliveryDecisionService_Calculate_HedgeLost(t *testing.T) {
	svc := NewDeliveryDecisionService()
	res, err := svc.Calculate(context.Background(), domain.DeliveryDecisionCalculateRequest{
		Input: domain.DeliveryDecisionInput{
			HardwareTotalCNY:       277000,
			HardwareCores:          128,
			HardwareTaxRate:        0.13,
			IDCRentMonthly:         4639,
			IDCRackKW:              3.52,
			IDCFillRate:            1.20,
			IDCServerPowerW:        570,
			IDCNetworkDepreciation: 3.62,
			CloudMemoryRatio:       8,
			CloudDiskRatio:         0,
			CloudCPUDailyPrice:     1.5,
			CloudMemoryDailyPrice:  0.2,
			CloudDiskDailyPrice:    0.05,
			CloudTaxRate:           0.06,
			DepreciationYears:      7,
			WACCRate:               0.03,
			ResidualRate:           0.05,
			Country:                "China",
			Currency:               "CNY",
			CloudCurrentDiscount:   0.25,
		},
	})
	if err != nil {
		t.Fatalf("calculate error: %v", err)
	}

	if !res.Formula.CloudHedgeLost {
		t.Fatal("expected hedge lost when cloud price is very low")
	}
	if !deliveryDecisionCloseEnough(res.Formula.FinalSelfShare, 1) || !deliveryDecisionCloseEnough(res.Formula.FinalCloudShare, 0) {
		t.Fatalf("unexpected final shares: %+v", res.Formula)
	}
}

func deliveryDecisionCloseEnough(got, want float64) bool {
	return math.Abs(got-want) < 0.0001
}
