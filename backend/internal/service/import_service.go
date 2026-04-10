package service

import (
	"context"
	"fmt"
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
	errRows := make([]RowError, 0)
	out := make([]domain.SpecialRule, 0, len(rows))
	for i, raw := range rows {
		rowNo := i + 2
		item, err := validateSpecialRuleRow(raw)
		if err != nil {
			errRows = append(errRows, RowError{Row: rowNo, Reason: err.Error()})
			continue
		}
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
	return domain.SpecialRule{SN: sn, Manufacturer: get("manufacturer"), Model: get("model"), PSA: get("psa"), IDC: get("idc"), PackageType: get("package_type"), WarrantyEndDate: get("warranty_end_date"), LaunchDate: get("launch_date"), Policy: policy}, nil
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

type FailureRateSummary struct {
	Segment              string  `json:"segment"`
	FullCycleFailureRate float64 `json:"full_cycle_failure_rate"`
	OverWarrantyRate     float64 `json:"over_warranty_failure_rate"`
	FaultCount           int     `json:"fault_count"`
	OverWarrantyFaults   int     `json:"over_warranty_fault_count"`
	ServerYears          float64 `json:"server_years"`
	OverWarrantyYears    float64 `json:"over_warranty_years"`
}

type FaultAnalysisResult struct {
	TotalFaultRows             int                  `json:"total_fault_rows"`
	MatchedFaultRows           int                  `json:"matched_fault_rows"`
	GeneratedModelRates        int                  `json:"generated_model_rates"`
	GeneratedPackageRates      int                  `json:"generated_package_rates"`
	GeneratedPackageModelRates int                  `json:"generated_package_model_rates"`
	OverallRates               []FailureRateSummary `json:"overall_rates,omitempty"`
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

func (s *ImportService) AnalyzeFaultRates(ctx context.Context, rows []map[string]string) (FaultAnalysisResult, error) {
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

	type faultEvent struct {
		createdAt *time.Time
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
	pkgOverNum := map[string]float64{}
	pkgOverDen := map[string]float64{}

	pkgModelNum := map[string]float64{}
	pkgModelDen := map[string]float64{}
	pkgModelOverNum := map[string]float64{}
	pkgModelOverDen := map[string]float64{}

	overallFaultNum := map[string]float64{"storage": 0, "non_storage": 0}
	overallFaultOverNum := map[string]float64{"storage": 0, "non_storage": 0}
	overallDenYears := map[string]float64{"storage": 0, "non_storage": 0}
	overallOverDenYears := map[string]float64{"storage": 0, "non_storage": 0}

	now := time.Now()
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

		years := yearsBetween(launchAt, now)
		if years <= 0 {
			continue
		}
		overStart := launchAt.AddDate(5, 0, 0)
		overYears := yearsBetween(overStart, now)

		modelKey := strings.Join([]string{strings.TrimSpace(srv.Manufacturer), strings.TrimSpace(srv.Model)}, "|")
		pkgKey := strings.TrimSpace(srv.ConfigType)
		pkgModelKey := strings.Join([]string{pkgKey, strings.TrimSpace(srv.Manufacturer), strings.TrimSpace(srv.Model)}, "|")

		weightedYears := weight * years
		modelDen[modelKey] += weightedYears
		pkgDen[pkgKey] += weightedYears
		pkgModelDen[pkgModelKey] += weightedYears
		overallDenYears[segment] += years

		if overYears > 0 {
			weightedOverYears := weight * overYears
			modelOverDen[modelKey] += weightedOverYears
			pkgOverDen[pkgKey] += weightedOverYears
			pkgModelOverDen[pkgModelKey] += weightedOverYears
			overallOverDenYears[segment] += overYears
		}

		events := faultEventsBySN[srv.SN]
		totalFaultN := float64(len(events))
		overFaultN := 0.0
		if overYears > 0 {
			for _, ev := range events {
				if ev.createdAt == nil {
					continue
				}
				if !ev.createdAt.Before(overStart) && !ev.createdAt.After(now) {
					overFaultN += 1
				}
			}
		}

		modelNum[modelKey] += totalFaultN
		pkgNum[pkgKey] += totalFaultN
		pkgModelNum[pkgModelKey] += totalFaultN

		modelOverNum[modelKey] += overFaultN
		pkgOverNum[pkgKey] += overFaultN
		pkgModelOverNum[pkgModelKey] += overFaultN

		overallFaultNum[segment] += totalFaultN
		overallFaultOverNum[segment] += overFaultN
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

	packageRates := make([]domain.PackageFailureRate, 0, len(pkgDen))
	for k, den := range pkgDen {
		if den <= 0 {
			continue
		}
		overRate := 0.0
		if pkgOverDen[k] > 0 {
			overRate = pkgOverNum[k] / pkgOverDen[k]
		}
		packageRates = append(packageRates, domain.PackageFailureRate{
			ConfigType:              k,
			FailureRate:             pkgNum[k] / den,
			OverWarrantyFailureRate: overRate,
		})
	}

	packageModelRates := make([]domain.PackageModelFailureRate, 0, len(pkgModelDen))
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
			ConfigType:              parts[0],
			Manufacturer:            parts[1],
			Model:                   parts[2],
			FailureRate:             pkgModelNum[k] / den,
			OverWarrantyFailureRate: overRate,
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

	buildSummary := func(segment string) FailureRateSummary {
		fullRate := 0.0
		if overallDenYears[segment] > 0 {
			fullRate = overallFaultNum[segment] / overallDenYears[segment]
		}
		overRate := 0.0
		if overallOverDenYears[segment] > 0 {
			overRate = overallFaultOverNum[segment] / overallOverDenYears[segment]
		}
		return FailureRateSummary{
			Segment:              segment,
			FullCycleFailureRate: fullRate,
			OverWarrantyRate:     overRate,
			FaultCount:           int(overallFaultNum[segment]),
			OverWarrantyFaults:   int(overallFaultOverNum[segment]),
			ServerYears:          overallDenYears[segment],
			OverWarrantyYears:    overallOverDenYears[segment],
		}
	}

	return FaultAnalysisResult{
		TotalFaultRows:             totalFaultRows,
		MatchedFaultRows:           matchedFaultRows,
		GeneratedModelRates:        len(modelRates),
		GeneratedPackageRates:      len(packageRates),
		GeneratedPackageModelRates: len(packageModelRates),
		OverallRates: []FailureRateSummary{
			buildSummary("storage"),
			buildSummary("non_storage"),
		},
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
