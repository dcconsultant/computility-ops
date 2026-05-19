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
	SourcePlanID string  `json:"source_plan_id"`
	DeviceCount  int     `json:"device_count"`
	CoveredCores int     `json:"covered_cores"`
	BudgetCNY    float64 `json:"budget_cny"`
}

type ResourcePlanningSelfRepairPlan struct {
	DeviceCount  int `json:"device_count"`
	CoveredCores int `json:"covered_cores"`
}

type ResourcePlanningDisposalPlan struct {
	DeviceCount  int `json:"device_count"`
	CoveredCores int `json:"covered_cores"`
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

	return ResourcePlanningResponse{
		GeneratedAt:       time.Now(),
		Config:            req,
		ReconfigPlan:      reconfigPlan,
		QuasiPurchasePlan: quasi,
		NewPurchasePlan:   newPlan,
		RenewalPlan:       renewalPlan,
		SelfRepairPlan:    selfRepairPlan,
		DisposalPlan:      disposalPlan,
	}, nil
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
	return ResourcePlanningRenewalPlan{
		SourcePlanID: latest.PlanID,
		DeviceCount:  latest.SelectedCount,
		CoveredCores: latest.SelectedCores,
		BudgetCNY:    round2RP(budget),
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

	routine := int(math.Floor(float64(baseDemand) / 10.0))
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
	cores := 0
	for _, srv := range servers {
		if !isPSAExcluded(srv.PSA, disposalPSAs) {
			continue
		}
		pkg, ok := pkgByConfig[srv.ConfigType]
		if !ok {
			continue
		}
		count++
		cores += pkg.CPULogicalCores
	}
	return ResourcePlanningDisposalPlan{DeviceCount: count, CoveredCores: cores}
}

func estimateMonthlyTCO(pkg domain.HostPackageConfig, original float64) float64 {
	dep := original / 60.0
	power := (pkg.PowerWatts / 1000.0) * 24 * 30 * 0.8
	other := original * 0.01
	return dep + power + other
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func isPSAExcluded(psa string, rules []string) bool {
	psa = strings.TrimSpace(psa)
	if psa == "" || len(rules) == 0 {
		return false
	}
	for _, r := range rules {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if psa == r || strings.HasPrefix(psa, strings.TrimRight(r, "/")+"/") {
			return true
		}
	}
	return false
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

func round2RP(v float64) float64 {
	return math.Round(v*100) / 100
}
