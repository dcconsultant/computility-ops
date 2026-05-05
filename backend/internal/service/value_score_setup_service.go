package service

import (
	"context"
	"fmt"
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
