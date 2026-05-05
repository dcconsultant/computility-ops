package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type CabinetService struct {
	repo repository.DatasetRepo
}

func NewCabinetService(repo repository.DatasetRepo) *CabinetService { return &CabinetService{repo: repo} }

func (s *CabinetService) GetUtilization(ctx context.Context) (domain.CabinetUtilizationSetting, error) {
	v, err := s.repo.GetCabinetUtilization(ctx)
	if err != nil {
		return domain.CabinetUtilizationSetting{}, err
	}
	if v.Utilization <= 0 {
		return domain.CabinetUtilizationSetting{Utilization: 1}, nil
	}
	return v, nil
}

func (s *CabinetService) UpdateUtilization(ctx context.Context, utilization float64) (domain.CabinetUtilizationSetting, error) {
	if utilization < 0.0001 || utilization > 2 {
		return domain.CabinetUtilizationSetting{}, fmt.Errorf("机柜利用率需在0.0001~2.0000之间")
	}
	setting := domain.CabinetUtilizationSetting{Utilization: utilization}
	if err := s.repo.SetCabinetUtilization(ctx, setting); err != nil {
		return domain.CabinetUtilizationSetting{}, err
	}
	return setting, nil
}

func (s *CabinetService) ListCabinetConfigs(ctx context.Context) ([]domain.CabinetConfig, error) {
	rows, err := s.repo.ListCabinetConfigs(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].IDC == rows[j].IDC {
			return rows[i].RatedPowerKW < rows[j].RatedPowerKW
		}
		return rows[i].IDC < rows[j].IDC
	})
	return rows, nil
}

func (s *CabinetService) CreateCabinetConfig(ctx context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error) {
	if err := validateCabinetConfig(row); err != nil {
		return domain.CabinetConfig{}, err
	}
	return s.repo.CreateCabinetConfig(ctx, row)
}

func (s *CabinetService) UpdateCabinetConfig(ctx context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error) {
	if row.ID <= 0 {
		return domain.CabinetConfig{}, fmt.Errorf("id 不能为空")
	}
	if err := validateCabinetConfig(row); err != nil {
		return domain.CabinetConfig{}, err
	}
	return s.repo.UpdateCabinetConfig(ctx, row)
}

func (s *CabinetService) DeleteCabinetConfig(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("id 无效")
	}
	return s.repo.DeleteCabinetConfig(ctx, id)
}

func validateCabinetConfig(row domain.CabinetConfig) error {
	row.IDC = strings.TrimSpace(row.IDC)
	if row.IDC == "" {
		return fmt.Errorf("机房不能为空")
	}
	if row.RatedPowerKW <= 0 {
		return fmt.Errorf("额定功率(KW) 必须大于0")
	}
	if row.MonthlyRent <= 0 {
		return fmt.Errorf("机柜月租 必须大于0")
	}
	return nil
}
