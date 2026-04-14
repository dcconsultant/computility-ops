package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type RowError struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type ImportResult struct {
	Total   int        `json:"total"`
	Success int        `json:"success"`
	Failed  int        `json:"failed"`
	Errors  []RowError `json:"errors"`
}

type ImportService struct {
	serverRepo  repository.ServerRepo
	datasetRepo repository.DatasetRepo
}

func NewImportService(serverRepo repository.ServerRepo, datasetRepo repository.DatasetRepo) *ImportService {
	return &ImportService{serverRepo: serverRepo, datasetRepo: datasetRepo}
}

func NormalizeHeaderName(raw string) string {
	n := strings.TrimSpace(strings.ToLower(raw))
	n = strings.ReplaceAll(n, " ", "")
	n = strings.ReplaceAll(n, "-", "")
	n = strings.ReplaceAll(n, "_", "")
	return n
}

func normalizeByMap(raw string, m map[string]string) string {
	n := NormalizeHeaderName(raw)
	if v, ok := m[n]; ok {
		return v
	}
	return n
}

func applyResult(total int, errs []RowError) ImportResult {
	s := total - len(errs)
	if s < 0 {
		s = 0
	}
	return ImportResult{Total: total, Success: s, Failed: len(errs), Errors: errs}
}

var serverHeaderMap = map[string]string{
	"sn":              "sn",
	"序列号":             "sn",
	"制造商":             "manufacturer",
	"厂商":              "manufacturer",
	"manufacturer":    "manufacturer",
	"型号":              "model",
	"服务器型号":           "model",
	"model":           "model",
	"psa":             "psa",
	"机房":              "idc",
	"idc":             "idc",
	"环境":              "environment",
	"env":             "environment",
	"environment":     "environment",
	"配置类型":            "config_type",
	"套餐":              "config_type",
	"configtype":      "config_type",
	"保修结束日期":          "warranty_end_date",
	"保修截止日期":          "warranty_end_date",
	"warrantyenddate": "warranty_end_date",
	"投产日期":            "launch_date",
	"launchdate":      "launch_date",
}

func (s *ImportService) ValidateAndReplaceServers(ctx context.Context, rows []map[string]string) (ImportResult, error) {
	errRows := make([]RowError, 0)
	out := make([]domain.Server, 0, len(rows))
	for i, raw := range rows {
		rowNo := i + 2
		server, err := validateServerRow(raw)
		if err != nil {
			errRows = append(errRows, RowError{Row: rowNo, Reason: err.Error()})
			continue
		}
		out = append(out, server)
	}
	res := applyResult(len(rows), errRows)
	if len(out) > 0 {
		if err := s.serverRepo.ReplaceAll(ctx, out); err != nil {
			return res, err
		}
	}
	return res, nil
}

func (s *ImportService) ListServers(ctx context.Context) ([]domain.Server, error) {
	return s.serverRepo.List(ctx)
}

func validateServerRow(raw map[string]string) (domain.Server, error) {
	get := func(k string) string { return strings.TrimSpace(raw[k]) }
	sn := get("sn")
	if sn == "" {
		return domain.Server{}, fmt.Errorf("SN 不能为空")
	}
	psa := get("psa")
	cfg := get("config_type")
	if cfg == "" {
		return domain.Server{}, fmt.Errorf("配置类型 不能为空")
	}
	return domain.Server{
		SN:              sn,
		Manufacturer:    get("manufacturer"),
		Model:           get("model"),
		PSA:             psa,
		IDC:             get("idc"),
		Environment:     get("environment"),
		ConfigType:      cfg,
		WarrantyEndDate: get("warranty_end_date"),
		LaunchDate:      get("launch_date"),
	}, nil
}

var hostPackageHeaderMap = map[string]string{
	"配置类型":                   "config_type",
	"套餐":                     "config_type",
	"configtype":             "config_type",
	"场景大类":                   "scene_category",
	"scenecategory":          "scene_category",
	"cpu逻辑核数":                "cpu_logical_cores",
	"cpulogicalcores":        "cpu_logical_cores",
	"gpu卡数":                  "gpu_card_count",
	"卡数":                     "gpu_card_count",
	"gpu_card_count":         "gpu_card_count",
	"gpucardcount":           "gpu_card_count",
	"数据盘类型":                  "data_disk_type",
	"数据盘种类":                  "data_disk_type",
	"datadisktype":           "data_disk_type",
	"磁盘类型":                   "data_disk_type",
	"disktype":               "data_disk_type",
	"数据盘数量":                  "data_disk_count",
	"datadiskcount":          "data_disk_count",
	"存储容量(tb)":               "storage_capacity_tb",
	"存储容量":                   "storage_capacity_tb",
	"storagecapacitytb":      "storage_capacity_tb",
	"服务器价值分":                 "server_value_score",
	"价值分":                    "server_value_score",
	"servervaluescore":       "server_value_score",
	"架构标准化系数":                "arch_standardized_factor",
	"archstandardizedfactor": "arch_standardized_factor",
}

func (s *ImportService) ValidateAndReplaceHostPackages(ctx context.Context, rows []map[string]string) (ImportResult, error) {
	errRows := make([]RowError, 0)
	out := make([]domain.HostPackageConfig, 0, len(rows))
	for i, raw := range rows {
		rowNo := i + 2
		item, err := validateHostPackageRow(raw)
		if err != nil {
			errRows = append(errRows, RowError{Row: rowNo, Reason: err.Error()})
			continue
		}
		out = append(out, item)
	}
	res := applyResult(len(rows), errRows)
	if len(out) > 0 {
		if err := s.datasetRepo.ReplaceHostPackages(ctx, out); err != nil {
			return res, err
		}
	}
	return res, nil
}
func (s *ImportService) ListHostPackages(ctx context.Context) ([]domain.HostPackageConfig, error) {
	return s.datasetRepo.ListHostPackages(ctx)
}

func validateHostPackageRow(raw map[string]string) (domain.HostPackageConfig, error) {
	get := func(k string) string { return strings.TrimSpace(raw[k]) }
	cfg := get("config_type")
	if cfg == "" {
		return domain.HostPackageConfig{}, fmt.Errorf("配置类型 不能为空")
	}
	cores, err := strconv.Atoi(get("cpu_logical_cores"))
	if err != nil || cores <= 0 {
		return domain.HostPackageConfig{}, fmt.Errorf("CPU逻辑核数 必须是大于0的整数")
	}
	coef := 1.0
	if v := get("arch_standardized_factor"); v != "" {
		coef, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.HostPackageConfig{}, fmt.Errorf("架构标准化系数 必须是数字")
		}
	}
	serverValueScore := 0.0
	if v := get("server_value_score"); v != "" {
		serverValueScore, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.HostPackageConfig{}, fmt.Errorf("服务器价值分 必须是数字")
		}
	}
	storage := 0.0
	if v := get("storage_capacity_tb"); v != "" {
		storage, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.HostPackageConfig{}, fmt.Errorf("存储容量(TB) 必须是数字")
		}
	}
	gpuCardCount := 0
	if v := get("gpu_card_count"); v != "" {
		gpuCardCount, err = strconv.Atoi(v)
		if err != nil || gpuCardCount < 0 {
			return domain.HostPackageConfig{}, fmt.Errorf("GPU卡数 必须是大于等于0的整数")
		}
	}
	dataDiskCount := 0
	if v := get("data_disk_count"); v != "" {
		dataDiskCount, err = strconv.Atoi(v)
		if err != nil || dataDiskCount < 0 {
			return domain.HostPackageConfig{}, fmt.Errorf("数据盘数量 必须是大于等于0的整数")
		}
	}
	return domain.HostPackageConfig{ConfigType: cfg, SceneCategory: get("scene_category"), CPULogicalCores: cores, GPUCardCount: gpuCardCount, DataDiskType: get("data_disk_type"), DataDiskCount: dataDiskCount, StorageCapacityTB: storage, ServerValueScore: serverValueScore, ArchStandardizedFactor: coef}, nil
}

var specialHeaderMap = map[string]string{
	"sn":           "sn",
	"序列号":          "sn",
	"制造商":          "manufacturer",
	"厂商":           "manufacturer",
	"manufacturer": "manufacturer",
	"型号":           "model",
	"model":        "model",
	"psa":          "psa",
	"机房":           "idc",
	"idc":          "idc",
	"套餐":           "package_type",
	"配置类型":         "package_type",
	"保修结束日期":       "warranty_end_date",
	"投产日期":         "launch_date",
	"策略":           "policy",
	"标签":           "policy",
	"黑白":           "policy",
}

func (s *ImportService) ValidateAndReplaceSpecialRules(ctx context.Context, rows []map[string]string) (ImportResult, error) {
	servers, err := s.serverRepo.List(ctx)
	if err != nil {
		return ImportResult{}, err
	}
	serverBySN := make(map[string]domain.Server, len(servers))
	for _, srv := range servers {
		sn := strings.TrimSpace(srv.SN)
		if sn == "" {
			continue
		}
		serverBySN[sn] = srv
	}

	errRows := make([]RowError, 0)
	out := make([]domain.SpecialRule, 0, len(rows))
	for i, raw := range rows {
		rowNo := i + 2
		item, err := validateSpecialRuleRow(raw)
		if err != nil {
			errRows = append(errRows, RowError{Row: rowNo, Reason: err.Error()})
			continue
		}

		srv, ok := serverBySN[item.SN]
		if !ok {
			errRows = append(errRows, RowError{Row: rowNo, Reason: "SN 不存在于服务器管理表"})
			continue
		}

		item.Manufacturer = strings.TrimSpace(srv.Manufacturer)
		item.Model = strings.TrimSpace(srv.Model)
		item.PSA = strings.TrimSpace(srv.PSA)
		item.IDC = strings.TrimSpace(srv.IDC)
		item.PackageType = strings.TrimSpace(srv.ConfigType)
		item.WarrantyEndDate = strings.TrimSpace(srv.WarrantyEndDate)
		item.LaunchDate = strings.TrimSpace(srv.LaunchDate)
		out = append(out, item)
	}
	res := applyResult(len(rows), errRows)
	if len(out) > 0 {
		if err := s.datasetRepo.ReplaceSpecialRules(ctx, out); err != nil {
			return res, err
		}
	}
	return res, nil
}
func (s *ImportService) ListSpecialRules(ctx context.Context) ([]domain.SpecialRule, error) {
	return s.datasetRepo.ListSpecialRules(ctx)
}

func validateSpecialRuleRow(raw map[string]string) (domain.SpecialRule, error) {
	get := func(k string) string { return strings.TrimSpace(raw[k]) }
	sn := get("sn")
	if sn == "" {
		return domain.SpecialRule{}, fmt.Errorf("SN 不能为空")
	}
	policy := normalizeSpecialPolicy(get("policy"))
	if policy == "" {
		return domain.SpecialRule{}, fmt.Errorf("策略必须是加白/加黑(whitelist/blacklist)")
	}
	return domain.SpecialRule{SN: sn, Policy: policy}, nil
}

func normalizeSpecialPolicy(v string) string {
	n := strings.ToLower(strings.TrimSpace(v))
	switch n {
	case "whitelist", "white", "加白", "白名单", "1", "renew":
		return "whitelist"
	case "blacklist", "black", "加黑", "黑名单", "-1", "drop", "norenew":
		return "blacklist"
	default:
		return ""
	}
}

var modelFailureHeaderMap = map[string]string{
	"厂商":                      "manufacturer",
	"制造商":                     "manufacturer",
	"manufacturer":            "manufacturer",
	"型号":                      "model",
	"服务器型号":                   "model",
	"model":                   "model",
	"故障率":                     "failure_rate",
	"failurerate":             "failure_rate",
	"过保故障率":                   "over_warranty_failure_rate",
	"overwarrantyfailurerate": "over_warranty_failure_rate",
}

func (s *ImportService) ValidateAndReplaceModelFailureRates(ctx context.Context, rows []map[string]string) (ImportResult, error) {
	out := make([]domain.ModelFailureRate, 0, len(rows))
	errRows := make([]RowError, 0)
	for i, raw := range rows {
		rowNo := i + 2
		v, err := validateModelFailureRow(raw)
		if err != nil {
			errRows = append(errRows, RowError{Row: rowNo, Reason: err.Error()})
			continue
		}
		out = append(out, v)
	}
	res := applyResult(len(rows), errRows)
	if len(out) > 0 {
		if err := s.datasetRepo.ReplaceModelFailureRates(ctx, out); err != nil {
			return res, err
		}
	}
	return res, nil
}
func (s *ImportService) ListModelFailureRates(ctx context.Context) ([]domain.ModelFailureRate, error) {
	return s.datasetRepo.ListModelFailureRates(ctx)
}

func validateModelFailureRow(raw map[string]string) (domain.ModelFailureRate, error) {
	get := func(k string) string { return strings.TrimSpace(raw[k]) }
	m := get("manufacturer")
	model := get("model")
	if m == "" || model == "" {
		return domain.ModelFailureRate{}, fmt.Errorf("厂商和型号不能为空")
	}
	rate, err := strconv.ParseFloat(get("failure_rate"), 64)
	if err != nil {
		return domain.ModelFailureRate{}, fmt.Errorf("故障率必须是数字")
	}
	overRate := 0.0
	if v := get("over_warranty_failure_rate"); v != "" {
		overRate, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.ModelFailureRate{}, fmt.Errorf("过保故障率必须是数字")
		}
	}
	return domain.ModelFailureRate{Manufacturer: m, Model: model, FailureRate: rate, OverWarrantyFailureRate: overRate}, nil
}

var packageFailureHeaderMap = map[string]string{
	"配置类型":                    "config_type",
	"套餐":                      "config_type",
	"configtype":              "config_type",
	"故障率":                     "failure_rate",
	"failurerate":             "failure_rate",
	"过保故障率":                   "over_warranty_failure_rate",
	"overwarrantyfailurerate": "over_warranty_failure_rate",
}

func (s *ImportService) ValidateAndReplacePackageFailureRates(ctx context.Context, rows []map[string]string) (ImportResult, error) {
	out := make([]domain.PackageFailureRate, 0, len(rows))
	errRows := make([]RowError, 0)
	for i, raw := range rows {
		rowNo := i + 2
		v, err := validatePackageFailureRow(raw)
		if err != nil {
			errRows = append(errRows, RowError{Row: rowNo, Reason: err.Error()})
			continue
		}
		out = append(out, v)
	}
	res := applyResult(len(rows), errRows)
	if len(out) > 0 {
		if err := s.datasetRepo.ReplacePackageFailureRates(ctx, out); err != nil {
			return res, err
		}
	}
	return res, nil
}
func (s *ImportService) ListPackageFailureRates(ctx context.Context) ([]domain.PackageFailureRate, error) {
	return s.datasetRepo.ListPackageFailureRates(ctx)
}

func validatePackageFailureRow(raw map[string]string) (domain.PackageFailureRate, error) {
	get := func(k string) string { return strings.TrimSpace(raw[k]) }
	cfg := get("config_type")
	if cfg == "" {
		return domain.PackageFailureRate{}, fmt.Errorf("配置类型不能为空")
	}
	rate, err := strconv.ParseFloat(get("failure_rate"), 64)
	if err != nil {
		return domain.PackageFailureRate{}, fmt.Errorf("故障率必须是数字")
	}
	overRate := 0.0
	if v := get("over_warranty_failure_rate"); v != "" {
		overRate, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.PackageFailureRate{}, fmt.Errorf("过保故障率必须是数字")
		}
	}
	return domain.PackageFailureRate{ConfigType: cfg, FailureRate: rate, OverWarrantyFailureRate: overRate}, nil
}

var packageModelFailureHeaderMap = map[string]string{
	"套餐":                      "config_type",
	"配置类型":                    "config_type",
	"configtype":              "config_type",
	"厂商":                      "manufacturer",
	"制造商":                     "manufacturer",
	"manufacturer":            "manufacturer",
	"型号":                      "model",
	"服务器型号":                   "model",
	"model":                   "model",
	"故障率":                     "failure_rate",
	"failurerate":             "failure_rate",
	"过保故障率":                   "over_warranty_failure_rate",
	"overwarrantyfailurerate": "over_warranty_failure_rate",
}

func (s *ImportService) ValidateAndReplacePackageModelFailureRates(ctx context.Context, rows []map[string]string) (ImportResult, error) {
	out := make([]domain.PackageModelFailureRate, 0, len(rows))
	errRows := make([]RowError, 0)
	for i, raw := range rows {
		rowNo := i + 2
		v, err := validatePackageModelFailureRow(raw)
		if err != nil {
			errRows = append(errRows, RowError{Row: rowNo, Reason: err.Error()})
			continue
		}
		out = append(out, v)
	}
	res := applyResult(len(rows), errRows)
	if len(out) > 0 {
		if err := s.datasetRepo.ReplacePackageModelFailureRates(ctx, out); err != nil {
			return res, err
		}
	}
	return res, nil
}
func (s *ImportService) ListPackageModelFailureRates(ctx context.Context) ([]domain.PackageModelFailureRate, error) {
	return s.datasetRepo.ListPackageModelFailureRates(ctx)
}

func (s *ImportService) ListOverallFailureRates(ctx context.Context) ([]domain.FailureRateSummary, error) {
	return s.datasetRepo.ListOverallFailureRates(ctx)
}

func (s *ImportService) ListFailureOverviewCards(ctx context.Context) ([]domain.FailureOverviewCard, error) {
	return s.datasetRepo.ListFailureOverviewCards(ctx)
}

func (s *ImportService) ListFailureAgeTrendPoints(ctx context.Context) ([]domain.FailureAgeTrendPoint, error) {
	rows, err := s.datasetRepo.ListFailureAgeTrendPoints(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *ImportService) ListFailureFeatureFacts(ctx context.Context) ([]domain.FailureFeatureFact, error) {
	rows, err := s.datasetRepo.ListFailureFeatureFacts(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *ImportService) ListStorageTopServerRates(ctx context.Context) ([]domain.StorageTopServerRate, error) {
	rows, err := s.datasetRepo.ListStorageTopServerRates(ctx)
	if err != nil {
		return nil, err
	}
	rows, err = s.enrichStorageTopRatesWithWarranty(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(rows) > 100 {
		rows = rows[:100]
	}
	return rows, nil
}

func (s *ImportService) ListWarmStorageServerRates(ctx context.Context) ([]domain.StorageTopServerRate, error) {
	rows, err := s.datasetRepo.ListStorageTopServerRates(ctx)
	if err != nil {
		return nil, err
	}
	return s.enrichStorageTopRatesWithWarranty(ctx, rows)
}

func (s *ImportService) enrichStorageTopRatesWithWarranty(ctx context.Context, rows []domain.StorageTopServerRate) ([]domain.StorageTopServerRate, error) {
	if len(rows) == 0 {
		return rows, nil
	}
	servers, err := s.serverRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	warrantyBySN := make(map[string]string, len(servers))
	for _, srv := range servers {
		warrantyBySN[strings.TrimSpace(srv.SN)] = strings.TrimSpace(srv.WarrantyEndDate)
	}
	for i := range rows {
		if strings.TrimSpace(rows[i].WarrantyEndDate) != "" {
			continue
		}
		rows[i].WarrantyEndDate = warrantyBySN[strings.TrimSpace(rows[i].SN)]
	}
	return rows, nil
}

func validatePackageModelFailureRow(raw map[string]string) (domain.PackageModelFailureRate, error) {
	get := func(k string) string { return strings.TrimSpace(raw[k]) }
	cfg, m, model := get("config_type"), get("manufacturer"), get("model")
	if cfg == "" || m == "" || model == "" {
		return domain.PackageModelFailureRate{}, fmt.Errorf("套餐、厂商、型号不能为空")
	}
	rate, err := strconv.ParseFloat(get("failure_rate"), 64)
	if err != nil {
		return domain.PackageModelFailureRate{}, fmt.Errorf("故障率必须是数字")
	}
	overRate := 0.0
	if v := get("over_warranty_failure_rate"); v != "" {
		overRate, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.PackageModelFailureRate{}, fmt.Errorf("过保故障率必须是数字")
		}
	}
	return domain.PackageModelFailureRate{ConfigType: cfg, Manufacturer: m, Model: model, FailureRate: rate, OverWarrantyFailureRate: overRate}, nil
}

type FaultAnalysisResult struct {
	TotalFaultRows             int                           `json:"total_fault_rows"`
	MatchedFaultRows           int                           `json:"matched_fault_rows"`
	GeneratedModelRates        int                           `json:"generated_model_rates"`
	GeneratedPackageRates      int                           `json:"generated_package_rates"`
	GeneratedPackageModelRates int                           `json:"generated_package_model_rates"`
	OverallRates               []domain.FailureRateSummary   `json:"overall_rates,omitempty"`
	OverviewCards              []domain.FailureOverviewCard  `json:"overview_cards,omitempty"`
	AgeTrendPoints             []domain.FailureAgeTrendPoint `json:"age_trend_points,omitempty"`
	FailureFeatureFacts        []domain.FailureFeatureFact    `json:"failure_feature_facts,omitempty"`
	StorageTopServerRates      []domain.StorageTopServerRate  `json:"storage_top_server_rates,omitempty"`
}

type faultEvent struct {
	createdAt *time.Time
}

type ageMetricAccumulator struct {
	numerator   float64
	denominator float64
}

func buildFailureAgeMetrics(
	servers []domain.Server,
	faultEventsBySN map[string][]faultEvent,
	pkgMap map[string]domain.HostPackageConfig,
	now time.Time,
	excludeOverWarranty bool,
) ([]domain.FailureAgeTrendPoint, []domain.FailureOverviewCard) {
	const maxAgeBucket = 10

	observationStart := now
	for _, events := range faultEventsBySN {
		for _, ev := range events {
			if ev.createdAt == nil {
				continue
			}
			if ev.createdAt.Before(observationStart) {
				observationStart = *ev.createdAt
			}
		}
	}
	if observationStart.After(now) {
		observationStart = now
	}
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	history := map[string]map[int]*ageMetricAccumulator{
		"storage":     {},
		"non_storage": {},
	}
	yearly := map[string]*ageMetricAccumulator{
		"storage":     &ageMetricAccumulator{},
		"non_storage": &ageMetricAccumulator{},
	}

	for _, srv := range servers {
		pkg, ok := pkgMap[strings.TrimSpace(srv.ConfigType)]
		if !ok {
			continue
		}
		purchaseDate, ok := parseFlexibleDate(srv.LaunchDate)
		if !ok {
			continue
		}

		segment := "non_storage"
		weight := 1.0
		bucket := normalizeBucket(pkg.SceneCategory)
		if bucket == "warm_storage" || bucket == "hot_storage" {
			segment = "storage"
			weight = 1 + float64(pkg.DataDiskCount)
		}

		warrantyEndAt, hasWarrantyEnd := parseFlexibleDate(srv.WarrantyEndDate)
		isOverWarranty := hasWarrantyEnd && warrantyEndAt.Before(now)
		if excludeOverWarranty && isOverWarranty {
			continue
		}
		analysisEnd := now
		if !analysisEnd.After(purchaseDate) {
			continue
		}

		for age := 1; age <= maxAgeBucket; age++ {
			start, end := ageBucketRange(purchaseDate, age)
			if intervalsOverlap(start, end, observationStart, analysisEnd) {
				if history[segment][age] == nil {
					history[segment][age] = &ageMetricAccumulator{}
				}
				history[segment][age].denominator += weight
			}
			if intervalsOverlap(start, end, yearStart, analysisEnd) {
				yearly[segment].denominator += weight
			}
		}

		events := faultEventsBySN[srv.SN]
		for _, ev := range events {
			if ev.createdAt == nil {
				continue
			}
			ts := *ev.createdAt
			if ts.Before(observationStart) || ts.After(now) {
				continue
			}
			age := ageBucketForEvent(purchaseDate, ts)
			if age >= 1 && age <= maxAgeBucket {
				if history[segment][age] == nil {
					history[segment][age] = &ageMetricAccumulator{}
				}
				history[segment][age].numerator += 1
			}
			if !ts.Before(yearStart) && !ts.After(analysisEnd) {
				yearly[segment].numerator += 1
			}
		}
	}

	trend := make([]domain.FailureAgeTrendPoint, 0, maxAgeBucket*2)
	cards := make([]domain.FailureOverviewCard, 0, 2)
	for _, segment := range []string{"storage", "non_storage"} {
		historyNumerator := 0.0
		historyDenominator := 0.0
		for age := 1; age <= maxAgeBucket; age++ {
			acc := history[segment][age]
			if acc == nil {
				acc = &ageMetricAccumulator{}
			}
			rate := 0.0
			if acc.denominator > 0 {
				rate = acc.numerator / acc.denominator
			}
			trend = append(trend, domain.FailureAgeTrendPoint{
				Segment:             segment,
				AgeBucket:           age,
				NumeratorFaultCount: int(acc.numerator),
				DenominatorExposure: acc.denominator,
				FaultRate:           rate,
			})
			historyNumerator += acc.numerator
			historyDenominator += acc.denominator
		}
		yearlyRate := 0.0
		if yearly[segment].denominator > 0 {
			yearlyRate = yearly[segment].numerator / yearly[segment].denominator
		}
		historyRate := 0.0
		if historyDenominator > 0 {
			historyRate = historyNumerator / historyDenominator
		}
		cards = append(cards, domain.FailureOverviewCard{
			Segment:                segment,
			Year:                   now.Year(),
			CurrentYearFaultRate:   yearlyRate,
			HistoryAvgFaultRate:    historyRate,
			CurrentYearFaultCount:  int(yearly[segment].numerator),
			CurrentYearDenominator: yearly[segment].denominator,
			HistoryFaultCount:      int(historyNumerator),
			HistoryDenominator:     historyDenominator,
		})
	}

	return trend, cards
}

type failureFeatureAgg struct {
	denominator float64
	faultCount  float64
}

func buildStorageTopServerRates(
	servers []domain.Server,
	faultEventsBySN map[string][]faultEvent,
	pkgMap map[string]domain.HostPackageConfig,
	now time.Time,
) []domain.StorageTopServerRate {
	windowStart := now.AddDate(-1, 0, 0)
	out := make([]domain.StorageTopServerRate, 0)
	for _, srv := range servers {
		pkg, ok := pkgMap[strings.TrimSpace(srv.ConfigType)]
		if !ok {
			continue
		}
		bucket := normalizeBucket(pkg.SceneCategory)
		if bucket != "warm_storage" {
			continue
		}
		totalCapacityTB := pkg.StorageCapacityTB
		singleDiskCapacityTB := 0.0
		if pkg.DataDiskCount > 0 {
			singleDiskCapacityTB = totalCapacityTB / float64(pkg.DataDiskCount)
		}
		den := 1 + float64(pkg.DataDiskCount)
		if den <= 0 {
			continue
		}
		faultCount := 0
		for _, ev := range faultEventsBySN[srv.SN] {
			if ev.createdAt == nil {
				continue
			}
			ts := *ev.createdAt
			if ts.Before(windowStart) || ts.After(now) {
				continue
			}
			faultCount++
		}
		rate := float64(faultCount) / den
		out = append(out, domain.StorageTopServerRate{
			SN:                   srv.SN,
			Manufacturer:         srv.Manufacturer,
			Model:                srv.Model,
			ConfigType:           srv.ConfigType,
			Environment:          srv.Environment,
			IDC:                  srv.IDC,
			DataDiskCount:        pkg.DataDiskCount,
			SingleDiskCapacityTB: singleDiskCapacityTB,
			TotalCapacityTB:      totalCapacityTB,
			FaultCount:           faultCount,
			Denominator:          den,
			FaultRate:            rate,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FaultRate != out[j].FaultRate {
			return out[i].FaultRate > out[j].FaultRate
		}
		if out[i].FaultCount != out[j].FaultCount {
			return out[i].FaultCount > out[j].FaultCount
		}
		return out[i].SN < out[j].SN
	})
	return out
}

func buildFailureFeatureFacts(
	servers []domain.Server,
	faultEventsBySN map[string][]faultEvent,
	pkgMap map[string]domain.HostPackageConfig,
	now time.Time,
) []domain.FailureFeatureFact {
	anchor := time.Date(2021, 4, 7, 0, 0, 0, 0, now.Location())
	if now.Before(anchor) {
		return nil
	}
	type recordYearWindow struct {
		index int
		start time.Time
		end   time.Time
	}
	windows := make([]recordYearWindow, 0)
	for i := 1; ; i++ {
		start := anchor.AddDate(i-1, 0, 0)
		if start.After(now) {
			break
		}
		end := anchor.AddDate(i, 0, 0).Add(-time.Nanosecond)
		windows = append(windows, recordYearWindow{index: i, start: start, end: end})
	}
	if len(windows) == 0 {
		return nil
	}

	type key struct {
		yearIndex int
		yearStart string
		yearEnd   string
		scope     string
		scene     string
		ageBucket int
	}
	agg := map[key]*failureFeatureAgg{}
	addDen := func(k key, v float64) {
		if agg[k] == nil {
			agg[k] = &failureFeatureAgg{}
		}
		agg[k].denominator += v
	}
	addFault := func(k key, v float64) {
		if agg[k] == nil {
			agg[k] = &failureFeatureAgg{}
		}
		agg[k].faultCount += v
	}

	rangeSceneGroups := func(bucket string) []string {
		out := []string{bucket}
		if bucket == "warm_storage" || bucket == "hot_storage" {
			out = append(out, "storage")
		} else {
			out = append(out, "non_storage")
		}
		return out
	}
	ageBucket := func(age int) int {
		if age <= 0 {
			return 0
		}
		if age >= 9 {
			return 9
		}
		return age
	}

	for _, srv := range servers {
		pkg, ok := pkgMap[strings.TrimSpace(srv.ConfigType)]
		if !ok {
			continue
		}
		launchAt, ok := parseFlexibleDate(srv.LaunchDate)
		if !ok {
			continue
		}
		warrantyEndAt, ok := parseFlexibleDate(srv.WarrantyEndDate)
		if !ok {
			continue
		}

		scope := classifyScopeByEnv(srv.Environment)
		scopes := []string{"all"}
		if scope != "all" {
			scopes = append(scopes, scope)
		}
		bucket := normalizeBucket(pkg.SceneCategory)
		sceneGroups := rangeSceneGroups(bucket)
		weight := 1.0
		if bucket == "warm_storage" || bucket == "hot_storage" {
			weight = 1 + float64(pkg.DataDiskCount)
		}

		for _, w := range windows {
			if warrantyEndAt.Before(w.start) {
				continue
			}
			age := int(w.start.Sub(launchAt).Hours()/24/365.2425) + 1
			ab := ageBucket(age)
			if ab == 0 {
				continue
			}
			yStart := w.start.Format("2006-01-02")
			yEnd := w.end.Format("2006-01-02")
			for _, sc := range scopes {
				for _, sg := range sceneGroups {
					addDen(key{yearIndex: w.index, yearStart: yStart, yearEnd: yEnd, scope: sc, scene: sg, ageBucket: ab}, weight)
				}
			}
		}

		events := faultEventsBySN[srv.SN]
		for _, ev := range events {
			if ev.createdAt == nil {
				continue
			}
			ts := *ev.createdAt
			if ts.Before(anchor) || ts.After(now) {
				continue
			}
			var matched *recordYearWindow
			for i := range windows {
				if !ts.Before(windows[i].start) && !ts.After(windows[i].end) {
					matched = &windows[i]
					break
				}
			}
			if matched == nil {
				continue
			}
			if warrantyEndAt.Before(matched.start) {
				continue
			}
			ab := ageBucket(int(matched.start.Sub(launchAt).Hours()/24/365.2425) + 1)
			if ab == 0 {
				continue
			}
			yStart := matched.start.Format("2006-01-02")
			yEnd := matched.end.Format("2006-01-02")
			for _, sc := range scopes {
				for _, sg := range sceneGroups {
					addFault(key{yearIndex: matched.index, yearStart: yStart, yearEnd: yEnd, scope: sc, scene: sg, ageBucket: ab}, 1)
				}
			}
		}
	}

	out := make([]domain.FailureFeatureFact, 0, len(agg))
	for k, v := range agg {
		rate := 0.0
		if v.denominator > 0 {
			rate = v.faultCount / v.denominator
		}
		out = append(out, domain.FailureFeatureFact{
			RecordYearIndex:     k.yearIndex,
			RecordYearStart:     k.yearStart,
			RecordYearEnd:       k.yearEnd,
			Scope:               k.scope,
			SceneGroup:          k.scene,
			AgeBucket:           k.ageBucket,
			DenominatorWeighted: v.denominator,
			FaultCount:          int(v.faultCount),
			FaultRate:           rate,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RecordYearIndex != out[j].RecordYearIndex {
			return out[i].RecordYearIndex < out[j].RecordYearIndex
		}
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		if out[i].SceneGroup != out[j].SceneGroup {
			return out[i].SceneGroup < out[j].SceneGroup
		}
		return out[i].AgeBucket < out[j].AgeBucket
	})
	return out
}

func ageBucketRange(purchaseDate time.Time, age int) (time.Time, time.Time) {
	start := purchaseDate.AddDate(age-1, 0, 0)
	end := purchaseDate.AddDate(age, 0, 0)
	return start, end
}

func intervalsOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

func ageBucketForEvent(purchaseDate, eventTime time.Time) int {
	if eventTime.Before(purchaseDate) {
		return 0
	}
	years := eventTime.Sub(purchaseDate).Hours() / 24 / 365.2425
	return int(years) + 1
}

var faultListHeaderMap = map[string]string{
	"类型":    "type",
	"主机名":   "hostname",
	"业务":    "business",
	"机房":    "idc",
	"机柜":    "rack",
	"厂商":    "manufacturer",
	"制造商":   "manufacturer",
	"型号":    "model",
	"sn":    "sn",
	"序列号":   "sn",
	"ip":    "ip",
	"ipmi":  "ipmi",
	"过保日期":  "warranty_end_date",
	"上报故障":  "reported_fault",
	"故障描述":  "fault_desc",
	"故障来源":  "fault_source",
	"业务对接人": "business_owner",
	"处理环节":  "process_stage",
	"工单状态":  "ticket_status",
	"真实故障":  "real_fault",
	"创建时间":  "created_at",
	"提单人":   "creator",
	"更新时间":  "updated_at",
	"结束时间":  "ended_at",
	"工单链接":  "ticket_link",
	"日志链接":  "log_link",
}

func (s *ImportService) AnalyzeFaultRates(ctx context.Context, rows []map[string]string, excludeOverWarranty bool) (FaultAnalysisResult, error) {
	servers, err := s.serverRepo.List(ctx)
	if err != nil {
		return FaultAnalysisResult{}, err
	}
	if len(servers) == 0 {
		return FaultAnalysisResult{}, fmt.Errorf("服务器管理表为空，无法分析")
	}
	packages, err := s.datasetRepo.ListHostPackages(ctx)
	if err != nil {
		return FaultAnalysisResult{}, err
	}
	if len(packages) == 0 {
		return FaultAnalysisResult{}, fmt.Errorf("主机套餐配置表为空，无法分析")
	}

	pkgMap := map[string]domain.HostPackageConfig{}
	for _, p := range packages {
		pkgMap[strings.TrimSpace(p.ConfigType)] = p
	}

	faultEventsBySN := map[string][]faultEvent{}
	totalFaultRows := 0
	matchedFaultRows := 0
	for _, raw := range rows {
		totalFaultRows++
		sn := strings.TrimSpace(raw["sn"])
		if sn == "" {
			continue
		}
		matchedFaultRows++
		var createdAtPtr *time.Time
		if ts, ok := parseFlexibleDateTime(raw["created_at"]); ok {
			createdAtPtr = &ts
		}
		faultEventsBySN[sn] = append(faultEventsBySN[sn], faultEvent{createdAt: createdAtPtr})
	}

	modelNum := map[string]float64{}
	modelDen := map[string]float64{}
	modelOverNum := map[string]float64{}
	modelOverDen := map[string]float64{}

	pkgNum := map[string]float64{}
	pkgDen := map[string]float64{}
	pkgBaseDen := map[string]float64{}
	pkgYearNum := map[string]float64{}
	pkgOverNum := map[string]float64{}
	pkgOverDen := map[string]float64{}

	pkgModelNum := map[string]float64{}
	pkgModelDen := map[string]float64{}
	pkgModelBaseDen := map[string]float64{}
	pkgModelYearNum := map[string]float64{}
	pkgModelOverNum := map[string]float64{}
	pkgModelOverDen := map[string]float64{}

	overallFaultNum := map[string]float64{}
	overallFaultOverNum := map[string]float64{}
	overallDenYears := map[string]float64{}
	overallOverDenYears := map[string]float64{}

	overallYearFaultNum := map[string]float64{}
	overallYearFaultOverNum := map[string]float64{}
	overallYearDenYears := map[string]float64{}
	overallYearOverDenYears := map[string]float64{}

	now := time.Now()
	minFaultYear := now.Year()
	overallBaseDen := map[string]float64{}
	overallFaultByYear := map[int]map[string]float64{}
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	for _, srv := range servers {
		pkg, ok := pkgMap[strings.TrimSpace(srv.ConfigType)]
		if !ok {
			continue
		}
		launchAt, ok := parseFlexibleDate(srv.LaunchDate)
		if !ok {
			continue
		}

		bucket := normalizeBucket(pkg.SceneCategory)
		segment := "non_storage"
		weight := 1.0
		if bucket == "warm_storage" || bucket == "hot_storage" {
			segment = "storage"
			weight = 1 + float64(pkg.DataDiskCount)
		}
		scope := classifyScopeByEnv(srv.Environment)
		k := scope + "|" + segment
		overallBaseDen[k] += weight
		overallBaseDen["all|"+segment] += weight

	warrantyEndAt, hasWarrantyEnd := parseFlexibleDate(srv.WarrantyEndDate)
		isOverWarranty := hasWarrantyEnd && warrantyEndAt.Before(now)
		if excludeOverWarranty && isOverWarranty {
			continue
		}
		analysisEnd := now
		years := yearsBetween(launchAt, analysisEnd)
		if years <= 0 {
			continue
		}
		overStart := launchAt.AddDate(5, 0, 0)
		overYears := yearsBetween(overStart, analysisEnd)
		yearServiceStart := maxTime(launchAt, yearStart)
		yearYears := yearsBetween(yearServiceStart, analysisEnd)
		yearOverStart := maxTime(overStart, yearStart)
		yearOverYears := yearsBetween(yearOverStart, analysisEnd)

		modelKey := strings.Join([]string{strings.TrimSpace(srv.Manufacturer), strings.TrimSpace(srv.Model)}, "|")
		pkgKey := strings.TrimSpace(srv.ConfigType)
		pkgModelKey := strings.Join([]string{pkgKey, strings.TrimSpace(srv.Manufacturer), strings.TrimSpace(srv.Model)}, "|")

		weightedYears := weight * years
		modelDen[modelKey] += weightedYears
		pkgDen[pkgKey] += weightedYears
		pkgBaseDen[pkgKey] += weight
		pkgModelDen[pkgModelKey] += weightedYears
		pkgModelBaseDen[pkgModelKey] += weight
		overallDenYears[k] += weightedYears
		overallDenYears["all|"+segment] += weightedYears
		if yearYears > 0 {
			weightedYearYears := weight * yearYears
			overallYearDenYears[k] += weightedYearYears
			overallYearDenYears["all|"+segment] += weightedYearYears
		}

		if overYears > 0 {
			weightedOverYears := weight * overYears
			modelOverDen[modelKey] += weightedOverYears
			pkgOverDen[pkgKey] += weightedOverYears
			pkgModelOverDen[pkgModelKey] += weightedOverYears
			overallOverDenYears[k] += weightedOverYears
			overallOverDenYears["all|"+segment] += weightedOverYears
		}
		if yearOverYears > 0 {
			weightedYearOverYears := weight * yearOverYears
			overallYearOverDenYears[k] += weightedYearOverYears
			overallYearOverDenYears["all|"+segment] += weightedYearOverYears
		}

		events := faultEventsBySN[srv.SN]
		totalFaultN := float64(len(events))
		overFaultN := 0.0
		yearFaultN := 0.0
		yearOverFaultN := 0.0
		for _, ev := range events {
			if ev.createdAt == nil {
				continue
			}
			if ev.createdAt.After(now) {
				continue
			}
			evYear := ev.createdAt.Year()
			if evYear <= now.Year() {
				if overallFaultByYear[evYear] == nil {
					overallFaultByYear[evYear] = map[string]float64{}
				}
				overallFaultByYear[evYear][k] += 1
				overallFaultByYear[evYear]["all|"+segment] += 1
				if evYear < minFaultYear {
					minFaultYear = evYear
				}
			}
			if !ev.createdAt.Before(yearStart) && !ev.createdAt.After(analysisEnd) {
				yearFaultN += 1
			}
			if overYears > 0 && !ev.createdAt.Before(overStart) && !ev.createdAt.After(analysisEnd) {
				overFaultN += 1
			}
			if yearOverYears > 0 && !ev.createdAt.Before(yearOverStart) && !ev.createdAt.After(analysisEnd) {
				yearOverFaultN += 1
			}
		}

		modelNum[modelKey] += totalFaultN
		pkgNum[pkgKey] += totalFaultN
		pkgYearNum[pkgKey] += yearFaultN
		pkgModelNum[pkgModelKey] += totalFaultN
		pkgModelYearNum[pkgModelKey] += yearFaultN

		modelOverNum[modelKey] += overFaultN
		pkgOverNum[pkgKey] += overFaultN
		pkgModelOverNum[pkgModelKey] += overFaultN

		overallFaultNum[k] += totalFaultN
		overallFaultOverNum[k] += overFaultN
		overallFaultNum["all|"+segment] += totalFaultN
		overallFaultOverNum["all|"+segment] += overFaultN

		overallYearFaultNum[k] += yearFaultN
		overallYearFaultOverNum[k] += yearOverFaultN
		overallYearFaultNum["all|"+segment] += yearFaultN
		overallYearFaultOverNum["all|"+segment] += yearOverFaultN
	}

	modelRates := make([]domain.ModelFailureRate, 0, len(modelDen))
	for k, den := range modelDen {
		if den <= 0 {
			continue
		}
		parts := strings.SplitN(k, "|", 2)
		if len(parts) != 2 {
			continue
		}
		overRate := 0.0
		if modelOverDen[k] > 0 {
			overRate = modelOverNum[k] / modelOverDen[k]
		}
		modelRates = append(modelRates, domain.ModelFailureRate{
			Manufacturer:            parts[0],
			Model:                   parts[1],
			FailureRate:             modelNum[k] / den,
			OverWarrantyFailureRate: overRate,
		})
	}

	packageRates := make([]domain.PackageFailureRate, 0, len(pkgDen)*2)
	for k, den := range pkgDen {
		if den <= 0 {
			continue
		}
		overRate := 0.0
		if pkgOverDen[k] > 0 {
			overRate = pkgOverNum[k] / pkgOverDen[k]
		}
		packageRates = append(packageRates, domain.PackageFailureRate{
			Period:                  "history",
			Year:                    0,
			ConfigType:              k,
			FailureRate:             pkgNum[k] / den,
			OverWarrantyFailureRate: overRate,
		})
		yearRate := safeDivide(pkgYearNum[k], pkgBaseDen[k]) * annualizationFactor(now.Year(), now)
		packageRates = append(packageRates, domain.PackageFailureRate{
			Period:                  "year",
			Year:                    now.Year(),
			ConfigType:              k,
			FailureRate:             yearRate,
			OverWarrantyFailureRate: 0,
		})
	}

	packageModelRates := make([]domain.PackageModelFailureRate, 0, len(pkgModelDen)*2)
	for k, den := range pkgModelDen {
		if den <= 0 {
			continue
		}
		parts := strings.SplitN(k, "|", 3)
		if len(parts) != 3 {
			continue
		}
		overRate := 0.0
		if pkgModelOverDen[k] > 0 {
			overRate = pkgModelOverNum[k] / pkgModelOverDen[k]
		}
		packageModelRates = append(packageModelRates, domain.PackageModelFailureRate{
			Period:                  "history",
			Year:                    0,
			ConfigType:              parts[0],
			Manufacturer:            parts[1],
			Model:                   parts[2],
			FailureRate:             pkgModelNum[k] / den,
			OverWarrantyFailureRate: overRate,
		})
		yearRate := safeDivide(pkgModelYearNum[k], pkgModelBaseDen[k]) * annualizationFactor(now.Year(), now)
		packageModelRates = append(packageModelRates, domain.PackageModelFailureRate{
			Period:                  "year",
			Year:                    now.Year(),
			ConfigType:              parts[0],
			Manufacturer:            parts[1],
			Model:                   parts[2],
			FailureRate:             yearRate,
			OverWarrantyFailureRate: 0,
		})
	}

	if err := s.datasetRepo.ReplaceModelFailureRates(ctx, modelRates); err != nil {
		return FaultAnalysisResult{}, err
	}
	if err := s.datasetRepo.ReplacePackageFailureRates(ctx, packageRates); err != nil {
		return FaultAnalysisResult{}, err
	}
	if err := s.datasetRepo.ReplacePackageModelFailureRates(ctx, packageModelRates); err != nil {
		return FaultAnalysisResult{}, err
	}

	overallRates := make([]domain.FailureRateSummary, 0, 64)
	getYearFault := func(year int, key string) float64 {
		if overallFaultByYear[year] == nil {
			return 0
		}
		return overallFaultByYear[year][key]
	}
	startYear := minFaultYear
	if len(overallFaultByYear) == 0 || startYear > now.Year() {
		startYear = now.Year()
	}
	yearFactor := annualizationFactor(now.Year(), now)
	for _, scope := range []string{"all", "product", "devtest"} {
		for _, segment := range []string{"storage", "non_storage"} {
			key := scope + "|" + segment
			den := overallBaseDen[key]
			yearFault := getYearFault(now.Year(), key)
			yearRate := safeDivide(yearFault, den) * yearFactor

			historyRateSum := 0.0
			historyFaultCount := 0.0
			yearCount := now.Year() - startYear + 1
			if yearCount <= 0 {
				yearCount = 1
			}
			for y := startYear; y <= now.Year(); y++ {
				fy := getYearFault(y, key)
				historyFaultCount += fy
				historyRateSum += safeDivide(fy, den) * annualizationFactor(y, now)
			}
			historyRate := historyRateSum / float64(yearCount)
			historyOverRate := safeDivide(overallFaultOverNum[key], overallOverDenYears[key])
			yearOverRate := safeDivide(overallYearFaultOverNum[key], overallYearOverDenYears[key])

			overallRates = append(overallRates,
				domain.FailureRateSummary{
					Period:               "history",
					Year:                 0,
					Scope:                scope,
					Segment:              segment,
					FullCycleFailureRate: historyRate,
					OverWarrantyRate:     historyOverRate,
					FaultCount:           int(historyFaultCount),
					OverWarrantyFaults:   int(overallFaultOverNum[key]),
					ServerYears:          den * float64(yearCount),
					OverWarrantyYears:    overallOverDenYears[key],
				},
				domain.FailureRateSummary{
					Period:               "year",
					Year:                 now.Year(),
					Scope:                scope,
					Segment:              segment,
					FullCycleFailureRate: yearRate,
					OverWarrantyRate:     yearOverRate,
					FaultCount:           int(yearFault),
					OverWarrantyFaults:   int(overallYearFaultOverNum[key]),
					ServerYears:          den,
					OverWarrantyYears:    overallYearOverDenYears[key],
				},
			)

			trendStartYear := startYear
			if trendStartYear < 2021 {
				trendStartYear = 2021
			}
			for y := trendStartYear; y <= now.Year(); y++ {
				fy := getYearFault(y, key)
				overallRates = append(overallRates, domain.FailureRateSummary{
					Period:               "year_trend",
					Year:                 y,
					Scope:                scope,
					Segment:              segment,
					FullCycleFailureRate: safeDivide(fy, den) * annualizationFactor(y, now),
					OverWarrantyRate:     0,
					FaultCount:           int(fy),
					OverWarrantyFaults:   0,
					ServerYears:          den,
					OverWarrantyYears:    0,
				})
			}
		}
	}
	if err := s.datasetRepo.ReplaceOverallFailureRates(ctx, overallRates); err != nil {
		return FaultAnalysisResult{}, err
	}

	ageTrendPoints, overviewCards := buildFailureAgeMetrics(servers, faultEventsBySN, pkgMap, now, excludeOverWarranty)
	if err := s.datasetRepo.ReplaceFailureAgeTrendPoints(ctx, ageTrendPoints); err != nil {
		return FaultAnalysisResult{}, err
	}
	if err := s.datasetRepo.ReplaceFailureOverviewCards(ctx, overviewCards); err != nil {
		return FaultAnalysisResult{}, err
	}

	featureFacts := buildFailureFeatureFacts(servers, faultEventsBySN, pkgMap, now)
	if err := s.datasetRepo.ReplaceFailureFeatureFacts(ctx, featureFacts); err != nil {
		return FaultAnalysisResult{}, err
	}
	storageTopRates := buildStorageTopServerRates(servers, faultEventsBySN, pkgMap, now)
	if err := s.datasetRepo.ReplaceStorageTopServerRates(ctx, storageTopRates); err != nil {
		return FaultAnalysisResult{}, err
	}

	return FaultAnalysisResult{
		TotalFaultRows:             totalFaultRows,
		MatchedFaultRows:           matchedFaultRows,
		GeneratedModelRates:        len(modelRates),
		GeneratedPackageRates:      len(packageRates),
		GeneratedPackageModelRates: len(packageModelRates),
		OverallRates:               overallRates,
		OverviewCards:              overviewCards,
		AgeTrendPoints:             ageTrendPoints,
		FailureFeatureFacts:        featureFacts,
		StorageTopServerRates:      storageTopRates,
	}, nil
}

func parseFlexibleDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{"2006-01-02", "2006/01/02", "2006/1/2", "2006-1-2", "20060102"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseFlexibleDateTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04",
		"2006-01-02",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func yearsBetween(start, end time.Time) float64 {
	if !end.After(start) {
		return 0
	}
	return end.Sub(start).Hours() / 24 / 365
}

func annualizationFactor(year int, now time.Time) float64 {
	if year != now.Year() {
		return 1
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
	elapsed := now.Sub(start).Hours() / 24
	if elapsed <= 0 {
		return 1
	}
	total := start.AddDate(1, 0, 0).Sub(start).Hours() / 24
	if total <= 0 {
		return 1
	}
	factor := total / elapsed
	if factor < 1 {
		return 1
	}
	return factor
}

func safeDivide(num, den float64) float64 {
	if den <= 0 {
		return 0
	}
	return num / den
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func classifyScopeByEnv(env string) string {
	n := normalizeText(env)
	if n == "product" {
		return "product"
	}
	return "devtest"
}

func MapHeaders(headers []string, headerMap map[string]string) []string {
	out := make([]string, len(headers))
	for i, h := range headers {
		out[i] = normalizeByMap(h, headerMap)
	}
	return out
}

func ValidateRequiredHeaders(headers []string, required ...string) error {
	m := map[string]bool{}
	for _, h := range headers {
		m[h] = true
	}
	miss := make([]string, 0)
	for _, r := range required {
		if !m[r] {
			miss = append(miss, r)
		}
	}
	if len(miss) > 0 {
		return fmt.Errorf("缺少必填列: %s", strings.Join(miss, ", "))
	}
	return nil
}
