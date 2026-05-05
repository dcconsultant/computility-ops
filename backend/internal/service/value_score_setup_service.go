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

func (s *ValueScoreSetupService) CalculateMonthlyTCO(ctx context.Context, req domain.ValueScoreTCOCalculateRequest) (domain.ValueScoreTCOCalculateResult, error) {
	baseline, err := s.GetCabinetBaseline(ctx)
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
		item := domain.ValueScoreTCOItem{
			ConfigType:          p.ConfigType,
			PowerWatts:          round4(powerW),
			PowerKW:             powerKW,
			CabinetCostMonthly:  cabCost,
			DepreciationMonthly: 0,
			TotalTCOMonthly:     cabCost,
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
		DepreciationMonths: 60,
		Formula:            baseline.Formula,
		Items:              items,
		Note:               "折旧月数固定 60（5*12）；当前折旧月值待后续接入真实折旧数据",
	}
	if baseline.Status != "ok" && baseline.Note != "" {
		res.Note = baseline.Note
	}
	return res, nil
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
