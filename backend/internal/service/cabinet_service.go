package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
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

type CabinetImportRowError struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type CabinetImportResult struct {
	Total   int                    `json:"total"`
	Success int                    `json:"success"`
	Failed  int                    `json:"failed"`
	Errors  []CabinetImportRowError `json:"errors"`
}

func (s *CabinetService) ImportCabinetConfigs(ctx context.Context, rows []map[string]string) (CabinetImportResult, error) {
	errRows := make([]CabinetImportRowError, 0)
	parsed := make([]domain.CabinetConfig, 0, len(rows))
	for i, raw := range rows {
		rowNo := i + 2
		item, err := parseCabinetImportRow(raw)
		if err != nil {
			errRows = append(errRows, CabinetImportRowError{Row: rowNo, Reason: err.Error()})
			continue
		}
		parsed = append(parsed, item)
	}
	result := CabinetImportResult{Total: len(rows), Failed: len(errRows), Errors: errRows}
	result.Success = result.Total - result.Failed
	if len(parsed) == 0 {
		return result, nil
	}

	existing, err := s.repo.ListCabinetConfigs(ctx)
	if err != nil {
		return result, err
	}
	idx := make(map[string]domain.CabinetConfig, len(existing))
	for _, e := range existing {
		idx[cabinetKey(e.IDC, e.RatedPowerKW)] = e
	}
	for _, row := range parsed {
		if ex, ok := idx[cabinetKey(row.IDC, row.RatedPowerKW)]; ok {
			row.ID = ex.ID
			if _, err := s.repo.UpdateCabinetConfig(ctx, row); err != nil {
				return result, err
			}
			continue
		}
		if _, err := s.repo.CreateCabinetConfig(ctx, row); err != nil {
			return result, err
		}
	}
	return result, nil
}

func cabinetKey(idc string, power float64) string {
	return strings.TrimSpace(strings.ToUpper(idc)) + "#" + strconv.FormatFloat(power, 'f', 8, 64)
}

func parseCabinetImportRow(raw map[string]string) (domain.CabinetConfig, error) {
	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k]; ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}
	idc := get("idc", "机房")
	if idc == "" {
		return domain.CabinetConfig{}, fmt.Errorf("机房不能为空")
	}
	powerStr := get("rated_power_kw", "额定功率(kw)", "额定功率（kw）", "额定功率")
	power, err := strconv.ParseFloat(powerStr, 64)
	if err != nil || power <= 0 {
		return domain.CabinetConfig{}, fmt.Errorf("额定功率(KW) 必须大于0")
	}
	rentStr := get("monthly_rent", "机柜月租(cny)", "机柜月租（cny）", "机柜月租")
	rent, err := strconv.ParseFloat(rentStr, 64)
	if err != nil || rent <= 0 {
		return domain.CabinetConfig{}, fmt.Errorf("机柜月租(CNY) 必须大于0")
	}
	return domain.CabinetConfig{IDC: idc, RatedPowerKW: power, MonthlyRent: rent}, nil
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
		return fmt.Errorf("机柜月租(CNY) 必须大于0")
	}
	return nil
}
