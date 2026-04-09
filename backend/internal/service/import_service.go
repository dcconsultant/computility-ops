package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	psa, err := strconv.ParseFloat(get("psa"), 64)
	if err != nil {
		return domain.Server{}, fmt.Errorf("PSA 必须是数字")
	}
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
	"磁盘类型":                   "disk_type",
	"disktype":               "disk_type",
	"存储容量(tb)":               "storage_capacity_tb",
	"存储容量":                   "storage_capacity_tb",
	"storagecapacitytb":      "storage_capacity_tb",
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
	storage := 0.0
	if v := get("storage_capacity_tb"); v != "" {
		storage, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return domain.HostPackageConfig{}, fmt.Errorf("存储容量(TB) 必须是数字")
		}
	}
	return domain.HostPackageConfig{ConfigType: cfg, SceneCategory: get("scene_category"), CPULogicalCores: cores, DiskType: get("disk_type"), StorageCapacityTB: storage, ArchStandardizedFactor: coef}, nil
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
	"厂商":           "manufacturer",
	"制造商":          "manufacturer",
	"manufacturer": "manufacturer",
	"型号":           "model",
	"model":        "model",
	"故障率":          "failure_rate",
	"failurerate":  "failure_rate",
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
	return domain.ModelFailureRate{Manufacturer: m, Model: model, FailureRate: rate}, nil
}

var packageFailureHeaderMap = map[string]string{
	"配置类型":        "config_type",
	"套餐":          "config_type",
	"configtype":  "config_type",
	"故障率":         "failure_rate",
	"failurerate": "failure_rate",
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
	return domain.PackageFailureRate{ConfigType: cfg, FailureRate: rate}, nil
}

var packageModelFailureHeaderMap = map[string]string{
	"套餐":           "config_type",
	"配置类型":         "config_type",
	"configtype":   "config_type",
	"厂商":           "manufacturer",
	"制造商":          "manufacturer",
	"manufacturer": "manufacturer",
	"型号":           "model",
	"model":        "model",
	"故障率":          "failure_rate",
	"failurerate":  "failure_rate",
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
	return domain.PackageModelFailureRate{ConfigType: cfg, Manufacturer: m, Model: model, FailureRate: rate}, nil
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
