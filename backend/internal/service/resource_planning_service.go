package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type ResourcePlanningService struct {
	serverRepo  repository.ServerRepo
	datasetRepo repository.DatasetRepo
	renewalRepo repository.RenewalPlanRepo
}

func NewResourcePlanningService(serverRepo repository.ServerRepo, datasetRepo repository.DatasetRepo, renewalRepo repository.RenewalPlanRepo) *ResourcePlanningService {
	return &ResourcePlanningService{serverRepo: serverRepo, datasetRepo: datasetRepo, renewalRepo: renewalRepo}
}

type ResourcePlanningRequest struct {
	AdmitValueScore           float64  `json:"admit_value_score"`
	ComputeDemandCores        int      `json:"compute_demand_cores"`
	WarmStorageDemandTB       float64  `json:"warm_storage_demand_tb"`
	HotStorageDemandTB        float64  `json:"hot_storage_demand_tb"`
	CabinetAndOtherCostCNY    float64  `json:"cabinet_and_other_cost_cny"`
	AnnualDepreciationCNY     float64  `json:"annual_depreciation_cny"`
	DisposalPSAs              string   `json:"disposal_psas"`
	NonBusinessPSAs           string   `json:"non_business_psas"`
	ReconfigDoneServerCount   *int     `json:"reconfig_done_server_count"`
	ReconfigDoneLogicalCores  *int     `json:"reconfig_done_logical_cores"`
	ReconfigDoneCostCNY       *float64 `json:"reconfig_done_cost_cny"`
	QuasiPurchaseServerCount  int      `json:"quasi_purchase_server_count"`
	QuasiPurchaseLogicalCores int      `json:"quasi_purchase_logical_cores"`
	QuasiPurchaseCostCNY      float64  `json:"quasi_purchase_cost_cny"`
}

type ResourcePlanningResponse struct {
	GeneratedAt       time.Time                         `json:"generated_at"`
	Config            ResourcePlanningRequest           `json:"config"`
	ReconfigPlan      ResourcePlanningReconfigPlan      `json:"reconfig_plan"`
	QuasiPurchasePlan ResourcePlanningQuasiPurchasePlan `json:"quasi_purchase_plan"`
	NewPurchasePlan   ResourcePlanningNewPurchasePlan   `json:"new_purchase_plan"`
	RenewalPlan       ResourcePlanningRenewalPlan       `json:"renewal_plan"`
	SelfRepairPlan    ResourcePlanningSelfRepairPlan    `json:"self_repair_plan"`
	DisposalPlan      ResourcePlanningDisposalPlan      `json:"disposal_plan"`
}

type ResourcePlanningConfigState struct {
	SavedAt time.Time               `json:"saved_at"`
	Config  ResourcePlanningRequest `json:"config"`
}

type ResourcePlanningReconfigPlan struct {
	SourcePlanID string  `json:"source_plan_id"`
	ServerCount  int     `json:"server_count"`
	LogicalCores int     `json:"logical_cores"`
	CostCNY      float64 `json:"cost_cny"`
}

type ResourcePlanningQuasiPurchasePlan struct {
	ServerCount  int     `json:"server_count"`
	LogicalCores int     `json:"logical_cores"`
	CostCNY      float64 `json:"cost_cny"`
}

type ResourcePlanningNewPurchasePlan struct {
	PackageConfigType       string  `json:"package_config_type"`
	PackageReleaseYear      int     `json:"package_release_year"`
	ServerCount             int     `json:"server_count"`
	CoveredLogicalCores     int     `json:"covered_logical_cores"`
	BaseDemandCores         int     `json:"base_demand_cores"`
	RoutineReplacementCores int     `json:"routine_replacement_cores"`
	ExtraReplacementCores   int     `json:"extra_replacement_cores"`
	TotalReplacementCores   int     `json:"total_replacement_cores"`
	PurchaseAmountCNY       float64 `json:"purchase_amount_cny"`
	AnnualCostCNY           float64 `json:"annual_cost_cny"`
	AnnualBudgetCNY         float64 `json:"annual_budget_cny"`
	ValueScore              float64 `json:"value_score"`
}

type ResourcePlanningRenewalPlan struct {
	SourcePlanID         string  `json:"source_plan_id"`
	DeviceCount          int     `json:"device_count"`
	CoveredComputeCores  int     `json:"covered_compute_cores"`
	CoveredWarmStorageTB float64 `json:"covered_warm_storage_tb"`
	CoveredHotStorageTB  float64 `json:"covered_hot_storage_tb"`
	CoveredGPUCards      int     `json:"covered_gpu_cards"`
	BudgetCNY            float64 `json:"budget_cny"`
}

type ResourcePlanningSelfRepairPlan struct {
	DeviceCount  int `json:"device_count"`
	CoveredCores int `json:"covered_cores"`
}

type ResourcePlanningDisposalPlan struct {
	DeviceCount           int      `json:"device_count"`
	CoveredComputeCores   int      `json:"covered_compute_cores"`
	CoveredWarmStorageTB  float64  `json:"covered_warm_storage_tb"`
	CoveredHotStorageTB   float64  `json:"covered_hot_storage_tb"`
	CoveredGPUCards       int      `json:"covered_gpu_cards"`
	UnmatchedPackageCount int      `json:"unmatched_package_count"`
	MatchedPSAServerCount int      `json:"matched_psa_server_count"`
	NormalizedPSAs        []string `json:"normalized_psas,omitempty"`
}

type reconfigSnapshotLite struct {
	PlanID             string    `json:"plan_id"`
	CreatedAt          time.Time `json:"created_at"`
	SuccessServerCount int       `json:"success_server_count"`
	SuccessCoreCount   float64   `json:"success_core_count"`
	ReconfigFee        float64   `json:"reconfig_fee"`
}

func (s *ResourcePlanningService) Calculate(ctx context.Context, req ResourcePlanningRequest) (ResourcePlanningResponse, error) {
	if req.QuasiPurchaseServerCount < 0 || req.QuasiPurchaseLogicalCores < 0 || req.QuasiPurchaseCostCNY < 0 {
		return ResourcePlanningResponse{}, fmt.Errorf("准系统采购利旧数据不能为空且必须>=0")
	}
	if req.ReconfigDoneServerCount != nil && *req.ReconfigDoneServerCount < 0 {
		return ResourcePlanningResponse{}, fmt.Errorf("已改配成功服务器数必须>=0")
	}
	if req.ReconfigDoneLogicalCores != nil && *req.ReconfigDoneLogicalCores < 0 {
		return ResourcePlanningResponse{}, fmt.Errorf("已改配成功逻辑核必须>=0")
	}
	if req.ReconfigDoneCostCNY != nil && *req.ReconfigDoneCostCNY < 0 {
		return ResourcePlanningResponse{}, fmt.Errorf("已改配费用必须>=0")
	}

	servers, err := s.serverRepo.List(ctx)
	if err != nil {
		return ResourcePlanningResponse{}, err
	}
	packages, err := s.datasetRepo.ListHostPackages(ctx)
	if err != nil {
		return ResourcePlanningResponse{}, err
	}
	if len(packages) == 0 {
		return ResourcePlanningResponse{}, fmt.Errorf("套餐模型为空，无法规划")
	}
	originalValues, err := s.datasetRepo.ListValueScoreOriginalValues(ctx)
	if err != nil {
		return ResourcePlanningResponse{}, err
	}
	if len(originalValues) == 0 {
		return ResourcePlanningResponse{}, fmt.Errorf("价值分缺失，请先导入价值分原值")
	}

	reconfigPlan, err := s.calcReconfigPlan(req)
	if err != nil {
		return ResourcePlanningResponse{}, err
	}

	quasi := ResourcePlanningQuasiPurchasePlan{
		ServerCount:  req.QuasiPurchaseServerCount,
		LogicalCores: req.QuasiPurchaseLogicalCores,
		CostCNY:      round2RP(req.QuasiPurchaseCostCNY),
	}

	renewalPlan, err := s.calcRenewalPlan(ctx)
	if err != nil {
		return ResourcePlanningResponse{}, err
	}

	nonBusinessPSAs := splitCSV(req.NonBusinessPSAs)
	disposalPSAs := splitCSV(req.DisposalPSAs)
	byConfig := mapPackageByConfigType(packages)
	eligibleCore := 0
	for _, srv := range servers {
		if isPSAExcluded(srv.PSA, nonBusinessPSAs) {
			continue
		}
		pkg, ok := byConfig[srv.ConfigType]
		if !ok {
			continue
		}
		if normalizeScene(pkg.SceneCategory) != "compute" {
			continue
		}
		eligibleCore += pkg.CPULogicalCores
	}

	baseDemand := req.ComputeDemandCores - reconfigPlan.LogicalCores - quasi.LogicalCores - eligibleCore
	if baseDemand < 0 {
		baseDemand = 0
	}

	newPlan, err := s.calcNewPurchasePlan(req, baseDemand, packages, originalValues, servers, nonBusinessPSAs)
	if err != nil {
		return ResourcePlanningResponse{}, err
	}

	selfRepairPlan := calcSelfRepairPlan(servers, byConfig)
	disposalPlan := calcDisposalPlan(servers, byConfig, disposalPSAs)

	out := ResourcePlanningResponse{
		GeneratedAt:       time.Now(),
		Config:            req,
		ReconfigPlan:      reconfigPlan,
		QuasiPurchasePlan: quasi,
		NewPurchasePlan:   newPlan,
		RenewalPlan:       renewalPlan,
		SelfRepairPlan:    selfRepairPlan,
		DisposalPlan:      disposalPlan,
	}
	_ = s.SaveConfig(ctx, req)
	return out, nil
}

func (s *ResourcePlanningService) calcReconfigPlan(req ResourcePlanningRequest) (ResourcePlanningReconfigPlan, error) {
	snap, err := loadLatestReconfigSnapshot(filepath.Join("backend", "logs", "reconfig-plans"))
	if err != nil {
		return ResourcePlanningReconfigPlan{}, err
	}
	plan := ResourcePlanningReconfigPlan{
		SourcePlanID: snap.PlanID,
		ServerCount:  snap.SuccessServerCount,
		LogicalCores: int(math.Round(snap.SuccessCoreCount)),
		CostCNY:      round2(snap.ReconfigFee),
	}
	if req.ReconfigDoneServerCount != nil || req.ReconfigDoneLogicalCores != nil || req.ReconfigDoneCostCNY != nil {
		if req.ReconfigDoneServerCount == nil || req.ReconfigDoneLogicalCores == nil || req.ReconfigDoneCostCNY == nil {
			return ResourcePlanningReconfigPlan{}, fmt.Errorf("已改配输入需同时提供成功服务器、成功逻辑核、费用")
		}
		plan.ServerCount += *req.ReconfigDoneServerCount
		plan.LogicalCores += *req.ReconfigDoneLogicalCores
		plan.CostCNY = round2RP(plan.CostCNY + *req.ReconfigDoneCostCNY)
	}
	return plan, nil
}

func (s *ResourcePlanningService) calcRenewalPlan(ctx context.Context) (ResourcePlanningRenewalPlan, error) {
	plans, err := s.renewalRepo.ListPlans(ctx)
	if err != nil {
		return ResourcePlanningRenewalPlan{}, err
	}
	if len(plans) == 0 {
		return ResourcePlanningRenewalPlan{}, fmt.Errorf("续保方案为空，请先生成续保方案")
	}
	latest := plans[0]
	for _, p := range plans[1:] {
		if strings.TrimSpace(p.PlanID) > strings.TrimSpace(latest.PlanID) {
			latest = p
		}
	}
	budget := 0.0
	if latest.DomesticBudget > 0 || latest.IndiaBudget > 0 {
		budget = latest.DomesticBudget + latest.IndiaBudget
	}
	coveredWarm := 0.0
	coveredHot := 0.0
	coveredGPU := 0
	for _, sec := range latest.Sections {
		bucket := normalizeScene(sec.Bucket)
		switch bucket {
		case "warm_storage":
			coveredWarm += sec.SelectedStorageTB
		case "hot_storage":
			coveredHot += sec.SelectedStorageTB
		case "gpu":
			for _, it := range sec.Items {
				coveredGPU += it.GPUCardCount
			}
		}
	}
	if coveredWarm == 0 || coveredHot == 0 || coveredGPU == 0 {
		for _, it := range latest.Items {
			bucket := normalizeScene(it.Bucket)
			if bucket == "" {
				bucket = normalizeScene(it.SceneCategory)
			}
			switch bucket {
			case "warm_storage":
				coveredWarm += it.StorageCapacityTB
			case "hot_storage":
				coveredHot += it.StorageCapacityTB
			case "gpu":
				coveredGPU += it.GPUCardCount
			}
		}
	}
	return ResourcePlanningRenewalPlan{
		SourcePlanID:         latest.PlanID,
		DeviceCount:          latest.SelectedCount,
		CoveredComputeCores:  latest.SelectedCores,
		CoveredWarmStorageTB: round2RP(coveredWarm),
		CoveredHotStorageTB:  round2RP(coveredHot),
		CoveredGPUCards:      coveredGPU,
		BudgetCNY:            round2RP(budget),
	}, nil
}

func (s *ResourcePlanningService) calcNewPurchasePlan(req ResourcePlanningRequest, baseDemand int, packages []domain.HostPackageConfig, originals []domain.ValueScoreOriginalValue, servers []domain.Server, nonBusinessPSAs []string) (ResourcePlanningNewPurchasePlan, error) {
	candidates := make([]domain.HostPackageConfig, 0)
	maxYear := -1
	for _, p := range packages {
		if normalizeScene(p.SceneCategory) != "compute" {
			continue
		}
		if p.ReleaseYear > maxYear {
			maxYear = p.ReleaseYear
			candidates = []domain.HostPackageConfig{p}
		} else if p.ReleaseYear == maxYear {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return ResourcePlanningNewPurchasePlan{}, fmt.Errorf("不存在场景大类=计算的套餐")
	}
	if len(candidates) > 1 {
		return ResourcePlanningNewPurchasePlan{}, fmt.Errorf("套餐发布年份并列，请处理后再规划")
	}
	selectedPkg := candidates[0]
	if selectedPkg.CPULogicalCores <= 0 {
		return ResourcePlanningNewPurchasePlan{}, fmt.Errorf("新机套餐逻辑核无效")
	}

	originByConfig := make(map[string]float64, len(originals))
	for _, o := range originals {
		originByConfig[strings.TrimSpace(o.ConfigType)] = o.ServerOriginalCNY
	}
	selectedOriginal, ok := originByConfig[selectedPkg.ConfigType]
	if !ok || selectedOriginal <= 0 {
		return ResourcePlanningNewPurchasePlan{}, fmt.Errorf("新机套餐缺少价值分原值")
	}
	selectedScore := 1.0 / selectedOriginal

	routine := int(math.Floor(float64(req.ComputeDemandCores) / 10.0))
	if routine < 0 {
		routine = 0
	}
	maxReplace := int(math.Floor(float64(baseDemand) / 4.0))
	if maxReplace < routine {
		maxReplace = routine
	}

	eligibleScores := make([]float64, 0)
	pkgByConfig := mapPackageByConfigType(packages)
	for _, srv := range servers {
		if isPSAExcluded(srv.PSA, nonBusinessPSAs) {
			continue
		}
		pkg, exists := pkgByConfig[srv.ConfigType]
		if !exists || normalizeScene(pkg.SceneCategory) != "compute" {
			continue
		}
		origin := originByConfig[srv.ConfigType]
		if origin <= 0 {
			continue
		}
		eligibleScores = append(eligibleScores, 1.0/origin)
	}
	sort.Float64s(eligibleScores)
	remaining := routine
	extra := 0
	for _, sc := range eligibleScores {
		if remaining >= maxReplace {
			break
		}
		if selectedScore > sc {
			remaining++
			extra++
		}
	}
	if remaining > maxReplace {
		remaining = maxReplace
	}
	if remaining < 0 {
		remaining = 0
	}
	if routine > 0 && remaining < routine {
		remaining = routine
	}
	serverCount := int(math.Ceil(float64(remaining) / float64(selectedPkg.CPULogicalCores)))
	coveredCores := serverCount * selectedPkg.CPULogicalCores
	monthlyTCO := estimateMonthlyTCO(selectedPkg, selectedOriginal)
	month := int(time.Now().Month())
	remainingMonths := 12 - month
	if remainingMonths < 0 {
		remainingMonths = 0
	}
	if remainingMonths > 0 {
		remainingMonths -= 1
	}
	return ResourcePlanningNewPurchasePlan{
		PackageConfigType:       selectedPkg.ConfigType,
		PackageReleaseYear:      selectedPkg.ReleaseYear,
		ServerCount:             serverCount,
		CoveredLogicalCores:     coveredCores,
		BaseDemandCores:         baseDemand,
		RoutineReplacementCores: routine,
		ExtraReplacementCores:   extra,
		TotalReplacementCores:   remaining,
		PurchaseAmountCNY:       round2RP(selectedOriginal * float64(serverCount)),
		AnnualCostCNY:           round2RP(monthlyTCO * 12 * float64(serverCount)),
		AnnualBudgetCNY:         round2RP(monthlyTCO * float64(remainingMonths) * float64(serverCount)),
		ValueScore:              selectedScore,
	}, nil
}

func calcSelfRepairPlan(servers []domain.Server, pkgByConfig map[string]domain.HostPackageConfig) ResourcePlanningSelfRepairPlan {
	count := 0
	cores := 0
	for _, srv := range servers {
		env := strings.ToLower(strings.TrimSpace(srv.Environment))
		if env != "test" && env != "dev" {
			continue
		}
		pkg, ok := pkgByConfig[srv.ConfigType]
		if !ok {
			continue
		}
		count++
		cores += pkg.CPULogicalCores
	}
	return ResourcePlanningSelfRepairPlan{DeviceCount: count, CoveredCores: cores}
}

func calcDisposalPlan(servers []domain.Server, pkgByConfig map[string]domain.HostPackageConfig, disposalPSAs []string) ResourcePlanningDisposalPlan {
	count := 0
	compute := 0
	warm := 0.0
	hot := 0.0
	gpu := 0
	unmatched := 0
	matchedPSA := 0
	for _, srv := range servers {
		if !isPSAExcluded(srv.PSA, disposalPSAs) {
			continue
		}
		matchedPSA++
		count++
		pkg, ok := pkgByConfig[srv.ConfigType]
		if !ok {
			unmatched++
			continue
		}
		scene := normalizeScene(pkg.SceneCategory)
		switch scene {
		case "compute":
			compute += pkg.CPULogicalCores
		case "warm_storage":
			warm += pkg.StorageCapacityTB
		case "hot_storage":
			hot += pkg.StorageCapacityTB
		case "gpu":
			gpu += pkg.GPUCardCount
		default:
			compute += pkg.CPULogicalCores
		}
	}
	return ResourcePlanningDisposalPlan{
		DeviceCount:           count,
		CoveredComputeCores:   compute,
		CoveredWarmStorageTB:  round2RP(warm),
		CoveredHotStorageTB:   round2RP(hot),
		CoveredGPUCards:       gpu,
		UnmatchedPackageCount: unmatched,
		MatchedPSAServerCount: matchedPSA,
		NormalizedPSAs:        disposalPSAs,
	}
}

func estimateMonthlyTCO(pkg domain.HostPackageConfig, original float64) float64 {
	dep := original / 60.0
	power := (pkg.PowerWatts / 1000.0) * 24 * 30 * 0.8
	other := original * 0.01
	return dep + power + other
}

func splitCSV(raw string) []string {
	replacer := strings.NewReplacer("，", ",", "；", ",", ";", ",", "\n", ",", "\t", ",")
	norm := replacer.Replace(raw)
	parts := strings.Split(norm, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		t := normalizePSAPath(p)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func isPSAExcluded(psa string, rules []string) bool {
	targets := splitPSATokensForRP(psa)
	if len(targets) == 0 || len(rules) == 0 {
		return false
	}

	ruleSet := map[string]struct{}{}
	for _, rawRule := range rules {
		for _, r := range splitPSATokensForRP(rawRule) {
			ruleSet[r] = struct{}{}
		}
	}
	if len(ruleSet) == 0 {
		return false
	}

	for _, target := range targets {
		for rule := range ruleSet {
			if target == rule || strings.HasPrefix(target, rule+"/") {
				return true
			}
		}
	}
	return false
}

func splitPSATokensForRP(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	// 支持 JSON 数组格式，例如：["/a/b","/c/d"]
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			out := make([]string, 0, len(arr))
			for _, x := range arr {
				if n := normalizePSAPath(x); n != "" {
					out = append(out, n)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '，', ';', '；':
			return true
		default:
			return false
		}
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if n := normalizePSAPath(p); n != "" {
			out = append(out, n)
		}
	}
	if len(out) > 0 {
		return out
	}

	if n := normalizePSAPath(raw); n != "" {
		return []string{n}
	}
	return nil
}

func normalizePSAPath(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// 去掉常见包裹字符："/a/b" 或 '/a/b'
	s = strings.Trim(s, "\"'")
	s = strings.ReplaceAll(s, "\\", "/")
	for strings.Contains(s, "//") {
		s = strings.ReplaceAll(s, "//", "/")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if len(s) > 1 {
		s = strings.TrimRight(s, "/")
	}
	return s
}

func normalizeScene(scene string) string {
	s := strings.ToLower(strings.TrimSpace(scene))
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "场景", "")
	s = strings.ReplaceAll(s, "大类", "")
	s = strings.ReplaceAll(s, "存储", "storage")
	s = strings.ReplaceAll(s, "温", "warm")
	s = strings.ReplaceAll(s, "热", "hot")
	s = strings.ReplaceAll(s, "计算", "compute")
	s = strings.ReplaceAll(s, "通用计算", "compute")
	s = strings.ReplaceAll(s, "暖", "warm")
	s = strings.ReplaceAll(s, "cold", "warm")
	s = strings.ReplaceAll(s, "hotstorage", "hotstorage")
	if strings.Contains(s, "gpu") {
		return "gpu"
	}
	if strings.Contains(s, "compute") {
		return "compute"
	}
	if strings.Contains(s, "warm") {
		return "warm_storage"
	}
	if strings.Contains(s, "hot") {
		return "hot_storage"
	}
	return s
}

func mapPackageByConfigType(rows []domain.HostPackageConfig) map[string]domain.HostPackageConfig {
	m := make(map[string]domain.HostPackageConfig, len(rows))
	for _, r := range rows {
		m[strings.TrimSpace(r.ConfigType)] = r
	}
	return m
}

func loadLatestReconfigSnapshot(base string) (reconfigSnapshotLite, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return reconfigSnapshotLite{}, fmt.Errorf("改配历史方案为空，请先生成改配方案")
		}
		return reconfigSnapshotLite{}, fmt.Errorf("读取改配历史方案失败: %w", err)
	}
	latest := reconfigSnapshotLite{}
	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		payload, readErr := os.ReadFile(filepath.Join(base, e.Name()))
		if readErr != nil {
			continue
		}
		var row reconfigSnapshotLite
		if jsonErr := json.Unmarshal(payload, &row); jsonErr != nil {
			continue
		}
		if !found || row.CreatedAt.After(latest.CreatedAt) {
			latest = row
			found = true
		}
	}
	if !found {
		return reconfigSnapshotLite{}, fmt.Errorf("改配历史方案为空，请先生成改配方案")
	}
	if latest.PlanID == "" {
		latest.PlanID = strconv.FormatInt(latest.CreatedAt.Unix(), 10)
	}
	return latest, nil
}

func (s *ResourcePlanningService) SaveConfig(ctx context.Context, req ResourcePlanningRequest) error {
	state := ResourcePlanningConfigState{SavedAt: time.Now(), Config: req}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join("backend", "logs", "resource-planning"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join("backend", "logs", "resource-planning", "config.latest.json"), payload, 0o644)
}

func (s *ResourcePlanningService) GetConfig(ctx context.Context) (ResourcePlanningConfigState, bool, error) {
	path := filepath.Join("backend", "logs", "resource-planning", "config.latest.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ResourcePlanningConfigState{}, false, nil
		}
		return ResourcePlanningConfigState{}, false, err
	}
	var state ResourcePlanningConfigState
	if err := json.Unmarshal(payload, &state); err != nil {
		return ResourcePlanningConfigState{}, false, err
	}
	return state, true, nil
}

func round2RP(v float64) float64 {
	return math.Round(v*100) / 100
}
