package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
)

const (
	deliveryDecisionFormulaVersion = "weighted-product-v5.0"
	daysPerMonth                   = 30.42
	agileHedgeYears                = 3.0
	equalityTolerance              = 0.000001
)

type DeliveryDecisionService struct{}

func NewDeliveryDecisionService() *DeliveryDecisionService {
	return &DeliveryDecisionService{}
}

func (s *DeliveryDecisionService) Defaults(_ context.Context, country string) domain.DeliveryDecisionDefaults {
	input := defaultDeliveryDecisionInput(country)
	return domain.DeliveryDecisionDefaults{
		Country:  input.Country,
		Currency: input.Currency,
		Input:    input,
	}
}

func (s *DeliveryDecisionService) Calculate(_ context.Context, req domain.DeliveryDecisionCalculateRequest) (domain.DeliveryDecisionResult, error) {
	input := normalizeDeliveryDecisionInput(req.Input)
	if err := validateDeliveryDecisionInput(input); err != nil {
		return domain.DeliveryDecisionResult{}, err
	}

	formula := calculateDeliveryDecisionFormula(input)
	points := buildDeliveryDecisionSensitivity(input)
	return domain.DeliveryDecisionResult{
		Input:             input,
		Formula:           formula,
		SensitivityPoints: points,
		Snapshot: domain.DeliveryDecisionSnapshot{
			Operator:       "sys",
			FormulaVersion: deliveryDecisionFormulaVersion,
			CalculatedAt:   time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}

func defaultDeliveryDecisionInput(country string) domain.DeliveryDecisionInput {
	country = deliveryDecisionNormalizeCountry(country)
	input := domain.DeliveryDecisionInput{
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
		CloudCPUDailyPrice:     0.78543,
		CloudMemoryDailyPrice:  0.11220,
		CloudDiskDailyPrice:    0.00732,
		CloudTaxRate:           0.06,
		DepreciationYears:      7,
		WACCRate:               0.03,
		ResidualRate:           0.05,
		Country:                country,
		Currency:               "CNY",
		CloudCurrentDiscount:   0.25,
	}
	if country == "India" {
		input.HardwareTaxRate = 0.18
		input.CloudTaxRate = 0.18
	}
	return input
}

func normalizeDeliveryDecisionInput(input domain.DeliveryDecisionInput) domain.DeliveryDecisionInput {
	if strings.TrimSpace(input.Country) == "" {
		input.Country = "China"
	}
	input.Country = deliveryDecisionNormalizeCountry(input.Country)
	if strings.TrimSpace(input.Currency) == "" {
		input.Currency = "CNY"
	}
	input.Currency = "CNY"
	if input.CloudCurrentDiscount == 0 {
		input.CloudCurrentDiscount = 0.25
	}
	return input
}

func deliveryDecisionNormalizeCountry(country string) string {
	switch strings.ToLower(strings.TrimSpace(country)) {
	case "india", "印度":
		return "India"
	default:
		return "China"
	}
}

func validateDeliveryDecisionInput(input domain.DeliveryDecisionInput) error {
	if input.HardwareCores <= 0 {
		return fmt.Errorf("整机物理核心数必须大于 0")
	}
	if input.IDCRackKW <= 0 {
		return fmt.Errorf("机柜功率必须大于 0")
	}
	if input.DepreciationYears <= 0 {
		return fmt.Errorf("财务折旧年限必须大于 0")
	}
	if input.DepreciationYears < 1 || input.DepreciationYears > 10 {
		return fmt.Errorf("财务折旧年限必须在 1~10 年范围内")
	}
	if input.HardwareTaxRate < 0 || input.HardwareTaxRate > 0.18 {
		return fmt.Errorf("硬件增值税率必须在 0%%~18%% 范围内")
	}
	if input.CloudTaxRate < 0 || input.CloudTaxRate > 0.17 {
		return fmt.Errorf("云服务进项税率必须在 0%%~17%% 范围内")
	}
	if input.WACCRate < 0 || input.WACCRate > 1 {
		return fmt.Errorf("WACC 年资金成本率必须在 0%%~100%% 范围内")
	}
	if input.ResidualRate < 0 || input.ResidualRate > 1 {
		return fmt.Errorf("设备残值率必须在 0%%~100%% 范围内")
	}
	if input.IDCFillRate < 0.5 || input.IDCFillRate > 1.5 {
		return fmt.Errorf("满柜率必须在 50%%~150%% 范围内")
	}
	if input.CloudCurrentDiscount <= 0 {
		return fmt.Errorf("当前公有云采购折扣必须大于 0")
	}
	amounts := []struct {
		name  string
		value float64
	}{
		{"整机采购总价", input.HardwareTotalCNY},
		{"机柜月租", input.IDCRentMonthly},
		{"单机功率", input.IDCServerPowerW},
		{"网络及机柜月折旧", input.IDCNetworkDepreciation},
		{"内存比", input.CloudMemoryRatio},
		{"磁盘比", input.CloudDiskRatio},
		{"CPU 单核日价", input.CloudCPUDailyPrice},
		{"内存日价", input.CloudMemoryDailyPrice},
		{"磁盘日价", input.CloudDiskDailyPrice},
	}
	for _, item := range amounts {
		if item.value < 0 || math.IsNaN(item.value) || math.IsInf(item.value, 0) {
			return fmt.Errorf("%s不得为负数或非法数值", item.name)
		}
	}
	return nil
}

func calculateDeliveryDecisionFormula(input domain.DeliveryDecisionInput) domain.DeliveryDecisionFormulaTrace {
	cloudGross := input.CloudCPUDailyPrice + input.CloudMemoryRatio*input.CloudMemoryDailyPrice + input.CloudDiskRatio*input.CloudDiskDailyPrice
	cloudDailyNet := cloudGross / (1 + input.CloudTaxRate)
	hardwareNetPerCore := input.HardwareTotalCNY / (1 + input.HardwareTaxRate) / input.HardwareCores
	serverKW := input.IDCServerPowerW / 1000
	physicalMonthlyNet := (input.IDCRentMonthly*serverKW/(input.IDCRackKW*input.IDCFillRate) + input.IDCNetworkDepreciation) / input.HardwareCores
	physicalDailyNet := physicalMonthlyNet / daysPerMonth
	dailyDepreciation := hardwareNetPerCore * (1 - input.ResidualRate) / (float64(input.DepreciationYears) * 365)
	dailyWACC := hardwareNetPerCore * input.WACCRate * (1 + input.ResidualRate) * 0.5 / 365
	selfDailyTCO := dailyDepreciation + dailyWACC + physicalDailyNet
	dailyDepreciation3Y := hardwareNetPerCore * (1 - input.ResidualRate) / (agileHedgeYears * 365)
	selfDailyTCO3Y := dailyDepreciation3Y + dailyWACC + physicalDailyNet
	r := cloudDailyNet / selfDailyTCO
	selfWeight := agileHedgeYears * r
	cloudWeight := float64(input.DepreciationYears)
	formulaSelfShare := selfWeight / (selfWeight + cloudWeight)
	formulaCloudShare := 1 - formulaSelfShare
	cloudHedgeLost := selfDailyTCO3Y <= cloudDailyNet+equalityTolerance
	finalSelfShare := formulaSelfShare
	finalCloudShare := formulaCloudShare
	if cloudHedgeLost {
		finalSelfShare = 1
		finalCloudShare = 0
	}
	dailyMargin := cloudDailyNet - physicalDailyNet
	var breakEvenYears *float64
	if dailyMargin > 0 {
		value := hardwareNetPerCore * (1 - input.ResidualRate) / (dailyMargin * 365)
		breakEvenYears = &value
	}

	return domain.DeliveryDecisionFormulaTrace{
		CloudGross:          cloudGross,
		CloudDailyNet:       cloudDailyNet,
		HardwareNetPerCore:  hardwareNetPerCore,
		ServerKW:            serverKW,
		PhysicalMonthlyNet:  physicalMonthlyNet,
		PhysicalDailyNet:    physicalDailyNet,
		DailyDepreciation:   dailyDepreciation,
		DailyWACC:           dailyWACC,
		DailyDepreciation3Y: dailyDepreciation3Y,
		SelfDailyTCO:        selfDailyTCO,
		SelfDailyTCO3Y:      selfDailyTCO3Y,
		PremiumRatioR:       r,
		SelfWeight:          selfWeight,
		CloudWeight:         cloudWeight,
		FormulaSelfShare:    formulaSelfShare,
		FormulaCloudShare:   formulaCloudShare,
		FinalSelfShare:      finalSelfShare,
		FinalCloudShare:     finalCloudShare,
		DailyMargin:         dailyMargin,
		BreakEvenYears:      breakEvenYears,
		CloudHedgeLost:      cloudHedgeLost,
		EqualityTolerance:   equalityTolerance,
	}
}

func buildDeliveryDecisionSensitivity(input domain.DeliveryDecisionInput) []domain.DeliveryDecisionSensitivityPoint {
	points := make([]domain.DeliveryDecisionSensitivityPoint, 0, 11)
	hardwareSamples := []struct {
		label string
		ratio float64
	}{
		{"0.25x", 0.25},
		{"0.5x", 0.5},
		{"当前", 1},
		{"2x", 2},
		{"4x", 4},
	}
	for _, sample := range hardwareSamples {
		next := input
		next.HardwareTotalCNY = input.HardwareTotalCNY * sample.ratio
		formula := calculateDeliveryDecisionFormula(next)
		points = append(points, domain.DeliveryDecisionSensitivityPoint{
			Curve:            "hardware_price",
			Label:            sample.label,
			XValue:           next.HardwareTotalCNY,
			CloudDailyNet:    formula.CloudDailyNet,
			SelfDailyTCO:     formula.SelfDailyTCO,
			SelfDailyTCO3Y:   formula.SelfDailyTCO3Y,
			FormulaSelfShare: formula.FormulaSelfShare,
			FinalSelfShare:   formula.FinalSelfShare,
			CloudHedgeLost:   formula.CloudHedgeLost,
		})
	}

	discountSamples := []float64{0.10, 0.15, 0.18, 0.25, 0.30, 0.35}
	for _, discount := range discountSamples {
		next := input
		ratio := discount / input.CloudCurrentDiscount
		next.CloudCPUDailyPrice = input.CloudCPUDailyPrice * ratio
		next.CloudMemoryDailyPrice = input.CloudMemoryDailyPrice * ratio
		next.CloudDiskDailyPrice = input.CloudDiskDailyPrice * ratio
		formula := calculateDeliveryDecisionFormula(next)
		points = append(points, domain.DeliveryDecisionSensitivityPoint{
			Curve:            "cloud_discount",
			Label:            fmt.Sprintf("%.0f折", discount*100),
			XValue:           discount,
			CloudDailyNet:    formula.CloudDailyNet,
			SelfDailyTCO:     formula.SelfDailyTCO,
			SelfDailyTCO3Y:   formula.SelfDailyTCO3Y,
			FormulaSelfShare: formula.FormulaSelfShare,
			FinalSelfShare:   formula.FinalSelfShare,
			CloudHedgeLost:   formula.CloudHedgeLost,
		})
	}
	return points
}
