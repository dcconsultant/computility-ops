package service

import (
	"context"
	"fmt"
	"math"
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

func (s *ValueScoreSetupService) GetCostSettings(ctx context.Context) (domain.ValueScoreCostSettings, error) {
	v, err := s.repo.GetValueScoreCostSettings(ctx)
	if err != nil {
		return domain.ValueScoreCostSettings{}, err
	}
	if v.CabinetUtilization <= 0 {
		v.CabinetUtilization = 1
	}
	if v.DepreciationMonths <= 0 {
		v.DepreciationMonths = 60
	}
	if v.ElectricityPriceCNYPerKWh < 0 {
		v.ElectricityPriceCNYPerKWh = 0
	}
	return v, nil
}

func (s *ValueScoreSetupService) UpdateCostSettings(ctx context.Context, in domain.ValueScoreCostSettings) (domain.ValueScoreCostSettings, error) {
	if in.ElectricityPriceCNYPerKWh < 0 {
		return domain.ValueScoreCostSettings{}, fmt.Errorf("电费单价不能为负数")
	}
	if in.DepreciationMonths <= 0 {
		return domain.ValueScoreCostSettings{}, fmt.Errorf("折旧月数必须大于0")
	}
	if in.CabinetUtilization < 0.0001 || in.CabinetUtilization > 2 {
		return domain.ValueScoreCostSettings{}, fmt.Errorf("机柜利用率需在0.0001~2.0000之间")
	}
	if err := s.repo.SetValueScoreCostSettings(ctx, in); err != nil {
		return domain.ValueScoreCostSettings{}, err
	}
	return in, nil
}

func (s *ValueScoreSetupService) CheckPackageCabinetMapping(ctx context.Context) ([]domain.PackageCabinetCheckItem, error) {
	pkgs, err := s.repo.ListHostPackages(ctx)
	if err != nil {
		return nil, err
	}
	servers, err := s.serverRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	cabs, err := s.repo.ListCabinetConfigs(ctx)
	if err != nil {
		return nil, err
	}
	cabIdx := make(map[string]struct{}, len(cabs))
	for _, c := range cabs {
		cabIdx[cabKey(c.IDC, c.RatedPowerKW)] = struct{}{}
	}

	pkgIDC := make(map[string]string)
	for _, sv := range servers {
		if strings.TrimSpace(sv.ConfigType) == "" || strings.TrimSpace(sv.IDC) == "" {
			continue
		}
		if _, ok := pkgIDC[sv.ConfigType]; !ok {
			pkgIDC[sv.ConfigType] = strings.TrimSpace(sv.IDC)
		}
	}

	out := make([]domain.PackageCabinetCheckItem, 0, len(pkgs))
	for _, p := range pkgs {
		idc := strings.TrimSpace(pkgIDC[p.ConfigType])
		powerKW := round4(p.PowerWatts / 1000)
		item := domain.PackageCabinetCheckItem{
			ConfigType: p.ConfigType,
			IDC:        idc,
			PowerWatts: p.PowerWatts,
			PowerKW:    powerKW,
		}
		if idc == "" {
			item.Matched = false
			item.Reason = "缺少机房映射（服务器数据中该配置类型未出现IDC）"
			out = append(out, item)
			continue
		}
		if p.PowerWatts <= 0 {
			item.Matched = false
			item.Reason = "套餐功率为空或<=0"
			out = append(out, item)
			continue
		}
		if _, ok := cabIdx[cabKey(idc, powerKW)]; !ok {
			item.Matched = false
			item.Reason = "未找到机柜配置（机房+额定功率）"
			out = append(out, item)
			continue
		}
		item.Matched = true
		out = append(out, item)
	}
	return out, nil
}

func cabKey(idc string, powerKW float64) string {
	return strings.ToUpper(strings.TrimSpace(idc)) + "#" + fmt.Sprintf("%.4f", powerKW)
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
