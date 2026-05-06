package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type ValueScoreSetupService struct {
	repo       repository.DatasetRepo
	serverRepo repository.ServerRepo
}

func NewValueScoreSetupService(repo repository.DatasetRepo, serverRepo repository.ServerRepo) *ValueScoreSetupService {
	return &ValueScoreSetupService{repo: repo, serverRepo: serverRepo}
}

func (s *ValueScoreSetupService) GetCabinetBaseline(ctx context.Context) (domain.ValueScoreCabinetBaseline, error) {
	util, err := s.repo.GetCabinetUtilization(ctx)
	if err != nil {
		return domain.ValueScoreCabinetBaseline{}, err
	}
	if util.Utilization <= 0 {
		util.Utilization = 1
	}
	cabs, err := s.repo.ListCabinetConfigs(ctx)
	if err != nil {
		return domain.ValueScoreCabinetBaseline{}, err
	}
	const targetIDC = "CN-N01-TJ01-ZJ01"
	minPower := 0.0
	minRent := 0.0
	count := 0
	for _, c := range cabs {
		if strings.EqualFold(strings.TrimSpace(c.IDC), targetIDC) {
			count++
			if minPower == 0 || c.RatedPowerKW < minPower || (c.RatedPowerKW == minPower && c.MonthlyRent < minRent) {
				minPower = c.RatedPowerKW
				minRent = c.MonthlyRent
			}
		}
	}
	if count == 0 || minPower <= 0 || minRent <= 0 {
		return domain.ValueScoreCabinetBaseline{Status: "warning", IDC: targetIDC, CabinetUtilization: util.Utilization, Formula: "机柜月租 * (功率(W)/1000) / (额定功率(KW) * 机柜利用率)", SourceCount: 0, Note: "目标机房没有可用机柜配置"}, nil
	}
	return domain.ValueScoreCabinetBaseline{
		Status:             "ok",
		IDC:                targetIDC,
		CabinetUtilization: util.Utilization,
		MinRatedPowerKW:    minPower,
		MonthlyRentCNY:     minRent,
		Formula:            "机柜月租 * (功率(W)/1000) / (额定功率(KW) * 机柜利用率)",
		SourceCount:        count,
		Note:               fmt.Sprintf("套餐展示按目标机房 %s 的最低额定功率机柜统一取参", targetIDC),
	}, nil
}

func (s *ValueScoreSetupService) GetCostParams(ctx context.Context) (domain.ValueScoreCostParams, error) {
	params, err := s.repo.GetValueScoreCostParams(ctx)
	if err != nil {
		return domain.ValueScoreCostParams{}, err
	}
	if params.DepreciationMonths <= 0 {
		params.DepreciationMonths = 60
	}
	if params.NetworkCabinetShareCNY < 0 {
		params.NetworkCabinetShareCNY = 0
	}
	if params.OtherFixedCostCNY < 0 {
		params.OtherFixedCostCNY = 0
	}
	return params, nil
}

func (s *ValueScoreSetupService) UpdateCostParams(ctx context.Context, params domain.ValueScoreCostParams) (domain.ValueScoreCostParams, error) {
	if params.DepreciationMonths <= 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("折旧月数必须大于0")
	}
	if params.NetworkCabinetShareCNY < 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("网络机柜分摊(CNY) 必须大于等于0")
	}
	if params.OtherFixedCostCNY < 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("其他固定成本(CNY) 必须大于等于0")
	}
	if err := s.repo.SetValueScoreCostParams(ctx, params); err != nil {
		return domain.ValueScoreCostParams{}, err
	}
	return params, nil
}

func (s *ValueScoreSetupService) CalculateMonthlyTCO(ctx context.Context, req domain.ValueScoreTCOCalculateRequest) (domain.ValueScoreTCOCalculateResult, error) {
	baseline, err := s.GetCabinetBaseline(ctx)
	if err != nil {
		return domain.ValueScoreTCOCalculateResult{}, err
	}
	params, err := s.GetCostParams(ctx)
	if err != nil {
		return domain.ValueScoreTCOCalculateResult{}, err
	}
	pkgs, err := s.repo.ListHostPackages(ctx)
	if err != nil {
		return domain.ValueScoreTCOCalculateResult{}, err
	}

	filter := make(map[string]struct{})
	for _, c := range req.ConfigTypes {
		k := strings.TrimSpace(c)
		if k != "" {
			filter[k] = struct{}{}
		}
	}

	items := make([]domain.ValueScoreTCOItem, 0, len(pkgs))
	for _, p := range pkgs {
		if len(filter) > 0 {
			if _, ok := filter[p.ConfigType]; !ok {
				continue
			}
		}
		powerW := p.PowerWatts
		powerKW := round4(powerW / 1000)
		cabCost := 0.0
		if baseline.MinRatedPowerKW > 0 && baseline.CabinetUtilization > 0 {
			cabCost = baseline.MonthlyRentCNY * powerKW / (baseline.MinRatedPowerKW * baseline.CabinetUtilization)
		}
		cabCost = round4(cabCost)
		depreciation := round4(p.MonthlyDepreciationCNY)
		networkShare := round4(params.NetworkCabinetShareCNY)
		otherFixed := round4(params.OtherFixedCostCNY)
		totalTCO := round4(cabCost + depreciation + networkShare + otherFixed)
		item := domain.ValueScoreTCOItem{
			ConfigType:            p.ConfigType,
			PowerWatts:            round4(powerW),
			PowerKW:               powerKW,
			CabinetCostMonthly:    cabCost,
			DepreciationMonthly:   depreciation,
			NetworkCabinetMonthly: networkShare,
			OtherFixedCostMonthly: otherFixed,
			TotalTCOMonthly:       totalTCO,
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool { return items[i].ConfigType < items[j].ConfigType })

	res := domain.ValueScoreTCOCalculateResult{
		Status:             baseline.Status,
		IDC:                baseline.IDC,
		CabinetUtilization: baseline.CabinetUtilization,
		MinRatedPowerKW:    baseline.MinRatedPowerKW,
		MonthlyRentCNY:     baseline.MonthlyRentCNY,
		DepreciationMonths: params.DepreciationMonths,
		Formula:            baseline.Formula,
		Items:              items,
		Note:               fmt.Sprintf("月TCO口径：机柜费 + 折旧 + 网络机柜等分摊 + 其他固定成本；折旧月数=%d", params.DepreciationMonths),
	}
	if baseline.Status != "ok" && baseline.Note != "" {
		res.Note = baseline.Note
	}
	return res, nil
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
