package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

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
	params, err := s.repo.GetValueScoreCostParams(ctx)
	if err != nil {
		return domain.ValueScoreCabinetBaseline{}, err
	}
	if params.CabinetUtilization <= 0 {
		params.CabinetUtilization = 1
	}
	status := "ok"
	note := "机柜基线已切换为全局参数（不再依赖机柜配置管理表）"
	if params.RatedPowerKW <= 0 || params.MonthlyRentCNY <= 0 {
		status = "warning"
		note = "全局参数缺少有效机柜额定功率或机柜月租"
	}
	return domain.ValueScoreCabinetBaseline{
		Status:             status,
		IDC:                "GLOBAL",
		CabinetUtilization: params.CabinetUtilization,
		MinRatedPowerKW:    params.RatedPowerKW,
		MonthlyRentCNY:     params.MonthlyRentCNY,
		Formula:            "机柜月租 * (功率(W)/1000) / (额定功率(KW) * 机柜利用率)",
		SourceCount:        1,
		Note:               note,
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
	if params.NetworkDeviceShareCNY < 0 {
		params.NetworkDeviceShareCNY = 0
	}
	if params.ServerRenewalFeeCNY < 0 {
		params.ServerRenewalFeeCNY = 0
	}
	if params.CabinetUtilization <= 0 {
		params.CabinetUtilization = 1
	}
	if params.RatedPowerKW < 0 {
		params.RatedPowerKW = 0
	}
	if params.MonthlyRentCNY < 0 {
		params.MonthlyRentCNY = 0
	}
	return params, nil
}

func (s *ValueScoreSetupService) UpdateCostParams(ctx context.Context, params domain.ValueScoreCostParams) (domain.ValueScoreCostParams, error) {
	// 020301口径：折旧月数固定60，不允许编辑
	params.DepreciationMonths = 60
	if params.NetworkDeviceShareCNY < 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("网络设备分摊成本(CNY) 必须大于等于0")
	}
	if params.ServerRenewalFeeCNY < 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("服务器续保费(CNY) 必须大于等于0")
	}
	if params.CabinetUtilization <= 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("机柜利用率必须大于0")
	}
	if params.RatedPowerKW <= 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("额定功率(KW) 必须大于0")
	}
	if params.MonthlyRentCNY <= 0 {
		return domain.ValueScoreCostParams{}, fmt.Errorf("机柜月租(CNY) 必须大于0")
	}
	if err := s.repo.SetValueScoreCostParams(ctx, params); err != nil {
		return domain.ValueScoreCostParams{}, err
	}
	return params, nil
}

func (s *ValueScoreSetupService) ReplaceOriginalValues(ctx context.Context, rows []domain.ValueScoreOriginalValue) error {
	for i := range rows {
		rows[i].ConfigType = strings.TrimSpace(rows[i].ConfigType)
		if rows[i].ConfigType == "" {
			return fmt.Errorf("第%d行配置类型不能为空", i+1)
		}
		if rows[i].ServerOriginalCNY < 0 {
			return fmt.Errorf("第%d行原值(CNY) 必须大于等于0", i+1)
		}
	}
	return s.repo.ReplaceValueScoreOriginalValues(ctx, rows)
}

func (s *ValueScoreSetupService) ListOriginalValues(ctx context.Context) ([]domain.ValueScoreOriginalValue, error) {
	return s.repo.ListValueScoreOriginalValues(ctx)
}

type ValueScorePerformanceImportRowError struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type ValueScorePerformanceImportRow struct {
	Row        int    `json:"row"`
	ConfigType string `json:"config_type"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
}

type ValueScorePerformanceImportResult struct {
	Total        int                                `json:"total"`
	NewCount     int                                `json:"new_count"`
	UpdatedCount int                                `json:"updated_count"`
	Failed       int                                `json:"failed"`
	Errors       []ValueScorePerformanceImportRowError `json:"errors"`
	Rows         []ValueScorePerformanceImportRow    `json:"rows,omitempty"`
}

func (s *ValueScoreSetupService) ListPerformanceParams(ctx context.Context) ([]domain.ValueScorePerformanceParam, error) {
	return s.repo.ListValueScorePerformanceParams(ctx)
}

func (s *ValueScoreSetupService) PreviewPerformanceParams(ctx context.Context, rows []map[string]string) (ValueScorePerformanceImportResult, error) {
	res, _, err := s.preparePerformanceParamsWithRows(ctx, rows)
	return res, err
}

func (s *ValueScoreSetupService) ImportPerformanceParams(ctx context.Context, rows []map[string]string) (ValueScorePerformanceImportResult, error) {
	res, parsed, err := s.preparePerformanceParamsWithRows(ctx, rows)
	if err != nil {
		return res, err
	}
	if len(parsed) == 0 {
		return res, nil
	}
	if err := s.repo.ReplaceValueScorePerformanceParams(ctx, parsed); err != nil {
		return res, err
	}
	return res, nil
}

func (s *ValueScoreSetupService) ImportUnifiedConfigParams(ctx context.Context, rows []map[string]string) (ValueScorePerformanceImportResult, error) {
	res, perfParsed, err := s.preparePerformanceParamsWithRows(ctx, rows)
	if err != nil {
		return res, err
	}
	if len(perfParsed) == 0 {
		return res, nil
	}
	rawByConfig := make(map[string]map[string]string, len(rows))
	for _, raw := range rows {
		cfg := strings.TrimSpace(raw["config_type"])
		if cfg == "" {
			continue
		}
		rawByConfig[cfg] = raw
	}
	originals := make([]domain.ValueScoreOriginalValue, 0, len(perfParsed))
	for _, item := range perfParsed {
		raw := rawByConfig[item.ConfigType]
		v := strings.TrimSpace(raw["server_original_cny"])
		if v == "" {
			return res, fmt.Errorf("配置类型 %s 缺少原值(CNY)", item.ConfigType)
		}
		num, e := strconv.ParseFloat(v, 64)
		if e != nil || num < 0 {
			return res, fmt.Errorf("配置类型 %s 的原值(CNY)必须是大于等于0的数字", item.ConfigType)
		}
		originals = append(originals, domain.ValueScoreOriginalValue{ConfigType: item.ConfigType, ServerOriginalCNY: num})
	}
	if err := s.repo.ReplaceValueScoreConfigParams(ctx, originals, perfParsed); err != nil {
		return res, err
	}
	return res, nil
}

func (s *ValueScoreSetupService) preparePerformanceParamsWithRows(ctx context.Context, rows []map[string]string) (ValueScorePerformanceImportResult, []domain.ValueScorePerformanceParam, error) {
	existingRows, err := s.repo.ListValueScorePerformanceParams(ctx)
	if err != nil {
		return ValueScorePerformanceImportResult{}, nil, err
	}
	existing := make(map[string]struct{}, len(existingRows))
	for _, row := range existingRows {
		k := strings.TrimSpace(row.ConfigType)
		if k != "" {
			existing[k] = struct{}{}
		}
	}
	seen := make(map[string]struct{}, len(rows))
	out := make([]domain.ValueScorePerformanceParam, 0, len(rows))
	reportRows := make([]ValueScorePerformanceImportRow, 0, len(rows))
	errRows := make([]ValueScorePerformanceImportRowError, 0)
	newCount := 0
	updatedCount := 0
	for i, raw := range rows {
		rowNo := i + 2
		item, err := validatePerformanceParamRow(raw)
		if err != nil {
			errRows = append(errRows, ValueScorePerformanceImportRowError{Row: rowNo, Reason: err.Error()})
			reportRows = append(reportRows, ValueScorePerformanceImportRow{Row: rowNo, ConfigType: strings.TrimSpace(raw["config_type"]), Status: "失败", Reason: err.Error()})
			continue
		}
		if _, ok := seen[item.ConfigType]; ok {
			reason := "配置类型重复"
			errRows = append(errRows, ValueScorePerformanceImportRowError{Row: rowNo, Reason: reason})
			reportRows = append(reportRows, ValueScorePerformanceImportRow{Row: rowNo, ConfigType: item.ConfigType, Status: "失败", Reason: reason})
			continue
		}
		seen[item.ConfigType] = struct{}{}
		status := "新增"
		if _, ok := existing[item.ConfigType]; ok {
			status = "更新"
			updatedCount++
		} else {
			newCount++
		}
		out = append(out, item)
		reportRows = append(reportRows, ValueScorePerformanceImportRow{Row: rowNo, ConfigType: item.ConfigType, Status: status})
	}
	res := ValueScorePerformanceImportResult{
		Total:        len(rows),
		NewCount:     newCount,
		UpdatedCount: updatedCount,
		Failed:       len(errRows),
		Errors:       errRows,
		Rows:         reportRows,
	}
	return res, out, nil
}

func validatePerformanceParamRow(raw map[string]string) (domain.ValueScorePerformanceParam, error) {
	get := func(k string) string { return strings.TrimSpace(raw[k]) }
	cfg := get("config_type")
	if cfg == "" {
		return domain.ValueScorePerformanceParam{}, fmt.Errorf("配置类型不能为空")
	}
	unavailableCores := 0
	if v := get("unavailable_cores"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return domain.ValueScorePerformanceParam{}, fmt.Errorf("不可用核数必须是整数")
		}
		if n < 0 {
			return domain.ValueScorePerformanceParam{}, fmt.Errorf("不可用核数必须大于等于0")
		}
		unavailableCores = n
	}
	unavailableMemory := 0.0
	if v := get("unavailable_memory_gb"); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.ValueScorePerformanceParam{}, fmt.Errorf("不可用内存容量(GB)必须是数字")
		}
		if n < 0 {
			return domain.ValueScorePerformanceParam{}, fmt.Errorf("不可用内存容量(GB)必须大于等于0")
		}
		unavailableMemory = n
	}
	scoreStr := get("performance_score")
	if scoreStr == "" {
		return domain.ValueScorePerformanceParam{}, fmt.Errorf("性能跑分不能为空")
	}
	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return domain.ValueScorePerformanceParam{}, fmt.Errorf("性能跑分必须是数字")
	}
	return domain.ValueScorePerformanceParam{ConfigType: cfg, UnavailableCores: unavailableCores, UnavailableMemoryGB: unavailableMemory, PerformanceScore: score}, nil
}

func (s *ValueScoreSetupService) CalculatePerformance(ctx context.Context) (domain.ValueScorePerformanceCalcResult, error) {
	packages, err := s.repo.ListHostPackages(ctx)
	if err != nil {
		return domain.ValueScorePerformanceCalcResult{}, err
	}
	params, err := s.repo.ListValueScorePerformanceParams(ctx)
	if err != nil {
		return domain.ValueScorePerformanceCalcResult{}, err
	}
	pkgByType := make(map[string]domain.HostPackageConfig, len(packages))
	for _, p := range packages {
		k := strings.TrimSpace(p.ConfigType)
		if k != "" {
			pkgByType[k] = p
		}
	}
	items := make([]domain.ValueScorePerformanceCalcItem, 0, len(params))
	alertCount := 0
	for _, p := range params {
		item := domain.ValueScorePerformanceCalcItem{ConfigType: p.ConfigType, UnavailableCores: p.UnavailableCores, UnavailableMemoryGB: round4(p.UnavailableMemoryGB), PerformanceScore: round4(p.PerformanceScore)}
		if pkg, ok := pkgByType[p.ConfigType]; ok {
			item.CPULogicalCores = pkg.CPULogicalCores
			item.MemoryCapacityGB = round4(pkg.MemoryCapacityGB)
			item.AvailableCores = pkg.CPULogicalCores - p.UnavailableCores
			item.AvailableMemoryGB = round4(pkg.MemoryCapacityGB - p.UnavailableMemoryGB)
			if p.PerformanceScore > 0 {
				item.StandardScore = round4(math.Pow(2277.0/p.PerformanceScore, 0.4))
			} else {
				item.StandardScore = 0
			}
			item.CPUPerformanceFactor = round4(item.StandardScore*0.43 + 0.57)
			if item.AvailableCores > 0 {
				item.MemoryRatio = round4(item.AvailableMemoryGB / float64(item.AvailableCores))
			}
			switch {
			case item.MemoryRatio <= 3:
				item.MemoryRatioFactor = 1.5
			case item.MemoryRatio < 6:
				item.MemoryRatioFactor = 1.25
			default:
				item.MemoryRatioFactor = 1
			}
			item.OverallPerformanceRatio = round4(item.CPUPerformanceFactor * item.MemoryRatioFactor)
		} else {
			item.Alerts = append(item.Alerts, domain.ValueScorePerformanceAlert{
				ConfigType:   p.ConfigType,
				ErrorCode:    "PACKAGE_NOT_FOUND",
				Field:        "config_type",
				CurrentValue: p.ConfigType,
				Suggestion:   "先导入主机套餐配置",
			})
			alertCount++
		}
		if p.PerformanceScore <= 0 {
			item.Alerts = append(item.Alerts, domain.ValueScorePerformanceAlert{ConfigType: p.ConfigType, ErrorCode: "PERFORMANCE_SCORE_INVALID", Field: "performance_score", CurrentValue: fmt.Sprintf("%.4f", p.PerformanceScore), Suggestion: "性能跑分必须大于0"})
			alertCount++
		}
		if item.AvailableCores <= 0 {
			item.Alerts = append(item.Alerts, domain.ValueScorePerformanceAlert{ConfigType: p.ConfigType, ErrorCode: "AVAILABLE_CORES_INVALID", Field: "available_cores", CurrentValue: fmt.Sprintf("%d", item.AvailableCores), Suggestion: "降低不可用核数"})
			alertCount++
		}
		if item.AvailableMemoryGB <= 0 {
			item.Alerts = append(item.Alerts, domain.ValueScorePerformanceAlert{ConfigType: p.ConfigType, ErrorCode: "AVAILABLE_MEMORY_INVALID", Field: "available_memory_gb", CurrentValue: fmt.Sprintf("%.4f", item.AvailableMemoryGB), Suggestion: "降低不可用内存容量"})
			alertCount++
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ConfigType < items[j].ConfigType })
	return domain.ValueScorePerformanceCalcResult{Items: items, AlertCount: alertCount, Note: "按配置类型联算：可用核数=CPU逻辑核数-不可用核数；可用内存容量=内存容量-不可用内存容量；整体性能折算比=CPU性能折算系数*内存配比系数"}, nil
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
	originalRows, err := s.repo.ListValueScoreOriginalValues(ctx)
	if err != nil {
		return domain.ValueScoreTCOCalculateResult{}, err
	}
	originalByConfig := map[string]float64{}
	for _, r := range originalRows {
		k := strings.TrimSpace(r.ConfigType)
		if k == "" {
			continue
		}
		if r.ServerOriginalCNY < 0 {
			originalByConfig[k] = 0
			continue
		}
		originalByConfig[k] = r.ServerOriginalCNY
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
		original := round4(originalByConfig[p.ConfigType])
		depreciation := 0.0
		if params.DepreciationMonths > 0 {
			age := 0
			if p.ReleaseYear > 0 {
				age = getCurrentYear() - p.ReleaseYear
			}
			if age >= 5 {
				depreciation = 0
			} else {
				depreciation = original * 0.95 / float64(params.DepreciationMonths)
			}
		}
		depreciation = round4(depreciation)
		networkCabinetShare := round4(cabCost * 0.2)
		networkDevice := round4(params.NetworkDeviceShareCNY)
		serverRenewal := round4(params.ServerRenewalFeeCNY)
		otherFixed := round4(networkDevice + serverRenewal)
		totalTCO := round4(cabCost + depreciation + networkDevice + networkCabinetShare + serverRenewal)
		item := domain.ValueScoreTCOItem{
			ConfigType:          p.ConfigType,
			PowerWatts:          round4(powerW),
			PowerKW:             powerKW,
			CabinetCostMonthly:  cabCost,
			ServerOriginalCNY:   original,
			DepreciationMonthly: depreciation,
			NetworkDeviceMonthly:  networkDevice,
			NetworkCabinetMonthly: networkCabinetShare,
			ServerRenewalMonthly:  serverRenewal,
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
		Note:               fmt.Sprintf("月TCO口径：机柜费 + 折旧 + 网络设备分摊成本 + 网络机柜等分摊 + 服务器续保费；折旧月数固定=%d", params.DepreciationMonths),
	}
	if baseline.Status != "ok" && baseline.Note != "" {
		res.Note = baseline.Note
	}
	return res, nil
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func getCurrentYear() int {
	return time.Now().Year()
}
