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

type CreatePlanInput struct {
	TargetDate           string
	ExcludedEnvironments []string
	ExcludedPSAs         []string
	TargetCores          int
	WarmTargetStorageTB  float64
	HotTargetStorageTB   float64
}

type RenewalService struct {
	serverRepo  repository.ServerRepo
	datasetRepo repository.DatasetRepo
	renewalRepo repository.RenewalPlanRepo
}

type ListPlansFilter struct {
	PlanID              string
	TargetDateFrom      string
	TargetDateTo        string
	ExcludedPSA         string
	ExcludedEnvironment string
}

func NewRenewalService(serverRepo repository.ServerRepo, datasetRepo repository.DatasetRepo, renewalRepo repository.RenewalPlanRepo) *RenewalService {
	return &RenewalService{serverRepo: serverRepo, datasetRepo: datasetRepo, renewalRepo: renewalRepo}
}

func (s *RenewalService) CreatePlan(ctx context.Context, in CreatePlanInput) (domain.RenewalPlan, error) {
	if in.TargetCores <= 0 {
		return domain.RenewalPlan{}, fmt.Errorf("target_cores must be > 0")
	}
	if strings.TrimSpace(in.TargetDate) == "" {
		return domain.RenewalPlan{}, fmt.Errorf("target_date is required")
	}
	targetDate, err := parseDate(in.TargetDate)
	if err != nil {
		return domain.RenewalPlan{}, fmt.Errorf("invalid target_date: %v", err)
	}

	servers, err := s.serverRepo.List(ctx)
	if err != nil {
		return domain.RenewalPlan{}, err
	}
	if len(servers) == 0 {
		return domain.RenewalPlan{}, fmt.Errorf("no servers imported")
	}
	packages, err := s.datasetRepo.ListHostPackages(ctx)
	if err != nil {
		return domain.RenewalPlan{}, err
	}
	if len(packages) == 0 {
		return domain.RenewalPlan{}, fmt.Errorf("host package config is empty")
	}
	specialRules, err := s.datasetRepo.ListSpecialRules(ctx)
	if err != nil {
		return domain.RenewalPlan{}, err
	}
	packageRates, err := s.datasetRepo.ListPackageFailureRates(ctx)
	if err != nil {
		return domain.RenewalPlan{}, err
	}

	excludedSet := make(map[string]bool)
	excludedCanonical := make([]string, 0, len(in.ExcludedEnvironments))
	for _, env := range in.ExcludedEnvironments {
		n := normalizeEnv(env)
		if n == "" || excludedSet[n] {
			continue
		}
		excludedSet[n] = true
		excludedCanonical = append(excludedCanonical, strings.TrimSpace(env))
	}
	if len(excludedCanonical) == 0 {
		excludedCanonical = []string{"开发", "测试"}
		excludedSet[normalizeEnv("开发")] = true
		excludedSet[normalizeEnv("测试")] = true
	}

	excludedPSAMatcher := newPSAMatcher()
	excludedPSACanonical := make([]string, 0, len(in.ExcludedPSAs))
	for _, psa := range in.ExcludedPSAs {
		n := normalizeText(psa)
		if n == "" {
			continue
		}
		if !excludedPSAMatcher.AddNormalized(n) {
			continue
		}
		excludedPSACanonical = append(excludedPSACanonical, strings.TrimSpace(psa))
	}

	pkgMap := map[string]domain.HostPackageConfig{}
	for _, p := range packages {
		pkgMap[strings.TrimSpace(p.ConfigType)] = p
	}

	pkgAFRAvg := map[string]float64{}
	for _, p := range packageRates {
		pkgAFRAvg[strings.TrimSpace(p.ConfigType)] = p.FailureRate
	}

	type specialPolicyRule struct {
		Policy string
		Reason string
	}
	specialMap := map[string]specialPolicyRule{}
	for _, sp := range specialRules {
		if sp.SN == "" {
			continue
		}
		specialMap[sp.SN] = specialPolicyRule{Policy: sp.Policy, Reason: strings.TrimSpace(sp.Reason)}
	}

	bucketItems := map[string][]domain.RenewalItem{
		"compute":      {},
		"warm_storage": {},
		"hot_storage":  {},
		"gpu":          {},
	}
	unmatchedConfigSet := map[string]bool{}
	nonRenewalItems := make([]domain.NonRenewalItem, 0)
	coveredComputeCores := 0
	coveredComputeCount := 0
	coveredWarmStorage := 0.0
	coveredWarmCount := 0
	coveredHotStorage := 0.0
	coveredHotCount := 0
	gpuCurrentCards := 0
	gpuCurrentServers := 0
	gpuCoveredCards := 0
	gpuCoveredServers := 0

	for _, srv := range servers {
		if excludedSet[normalizeEnv(srv.Environment)] {
			continue
		}
		if excludedPSAMatcher.MatchRaw(srv.PSA) {
			nonRenewalItems = append(nonRenewalItems, domain.NonRenewalItem{
				SN:           srv.SN,
				Manufacturer: srv.Manufacturer,
				Model:        srv.Model,
				Environment:  srv.Environment,
				IDC:          srv.IDC,
				ConfigType:   srv.ConfigType,
				PSA:          domain.PSAString(strings.TrimSpace(srv.PSA)),
				ReasonCode:   "psa_exception",
				Reason:       "PSA例外",
				ReasonDetail: fmt.Sprintf("PSA=%s 命中排除条件", strings.TrimSpace(srv.PSA)),
			})
			continue
		}
		if strings.TrimSpace(srv.WarrantyEndDate) == "" {
			continue
		}
		wd, err := parseDate(srv.WarrantyEndDate)
		if err != nil {
			return domain.RenewalPlan{}, fmt.Errorf("invalid warranty_end_date for sn=%s: %v", srv.SN, err)
		}

		pkg, ok := pkgMap[srv.ConfigType]
		if !ok {
			unmatchedConfigSet[strings.TrimSpace(srv.ConfigType)] = true
			continue
		}
		cores := pkg.CPULogicalCores
		if cores <= 0 {
			return domain.RenewalPlan{}, fmt.Errorf("invalid cpu_logical_cores for config_type=%s", srv.ConfigType)
		}
		coef := pkg.ArchStandardizedFactor
		if coef == 0 {
			coef = 1
		}

		bucket := normalizeBucket(pkg.SceneCategory)
		gpuCards := pkg.GPUCardCount
		if bucket == "gpu" {
			gpuCurrentCards += gpuCards
			gpuCurrentServers++
		}
		if !wd.Before(targetDate) {
			// 未过保：计入目标覆盖基线，但不进入续保方案列表
			switch bucket {
			case "compute":
				coveredComputeCores += cores
				coveredComputeCount++
			case "warm_storage":
				coveredWarmStorage += pkg.StorageCapacityTB
				coveredWarmCount++
			case "hot_storage":
				coveredHotStorage += pkg.StorageCapacityTB
				coveredHotCount++
			case "gpu":
				gpuCoveredCards += gpuCards
				gpuCoveredServers++
			}
			continue
		}

		baseValue := pkg.ServerValueScore
		baseScore := baseValue * coef
		item := domain.RenewalItem{
			SN:                     srv.SN,
			Bucket:                 bucket,
			Manufacturer:           srv.Manufacturer,
			Model:                  srv.Model,
			Environment:            srv.Environment,
			IDC:                    srv.IDC,
			ConfigType:             srv.ConfigType,
			SceneCategory:          pkg.SceneCategory,
			CPULogicalCores:        cores,
			GPUCardCount:           gpuCards,
			StorageCapacityTB:      pkg.StorageCapacityTB,
			PSA:                    domain.PSAString(strings.TrimSpace(srv.PSA)),
			ArchStandardizedFactor: coef,
			BaseScore:              baseScore,
			FinalScore:             baseScore,
		}

		if bucket == "warm_storage" || bucket == "hot_storage" {
			afrAvg := pkgAFRAvg[strings.TrimSpace(srv.ConfigType)]
			if afrAvg > 0 {
				item.AFRAvg = afrAvg
				item.FailureAdjustFactor = 1 / afrAvg
				item.FinalScore = baseScore / afrAvg
			}
		}

		specialRule, hasSpecialRule := specialMap[srv.SN]
		specialPolicy := specialRule.Policy
		if specialPolicy == "blacklist" {
			reasonDetail := "命中特殊名单黑名单策略"
			if specialRule.Reason != "" {
				reasonDetail = specialRule.Reason
			}
			nonRenewalItems = append(nonRenewalItems, domain.NonRenewalItem{
				SN:           item.SN,
				Bucket:       item.Bucket,
				Manufacturer: item.Manufacturer,
				Model:        item.Model,
				Environment:  item.Environment,
				IDC:          item.IDC,
				ConfigType:   item.ConfigType,
				PSA:          item.PSA,
				FinalScore:   item.FinalScore,
				ReasonCode:   "blacklist",
				Reason:       "黑名单",
				ReasonDetail: reasonDetail,
			})
			continue
		}
		if hasSpecialRule && specialPolicy == "whitelist" {
			item.SpecialPolicy = "whitelist"
		}

		bucketItems[bucket] = append(bucketItems[bucket], item)
	}

	sortItems := func(items []domain.RenewalItem) {
		sort.Slice(items, func(i, j int) bool {
			if items[i].FinalScore != items[j].FinalScore {
				return items[i].FinalScore > items[j].FinalScore
			}
			if items[i].StorageCapacityTB != items[j].StorageCapacityTB {
				return items[i].StorageCapacityTB > items[j].StorageCapacityTB
			}
			if items[i].CPULogicalCores != items[j].CPULogicalCores {
				return items[i].CPULogicalCores > items[j].CPULogicalCores
			}
			return items[i].SN < items[j].SN
		})
	}
	for k := range bucketItems {
		sortItems(bucketItems[k])
	}

	selectByCores := func(items []domain.RenewalItem, target int) ([]domain.RenewalItem, int) {
		must, normal := splitByWhitelist(items)
		selected := make([]domain.RenewalItem, 0, len(items))
		pickedCores := 0
		for _, item := range must {
			selected = append(selected, item)
			pickedCores += item.CPULogicalCores
		}
		for _, item := range normal {
			if pickedCores >= target {
				break
			}
			selected = append(selected, item)
			pickedCores += item.CPULogicalCores
		}
		return selected, pickedCores
	}

	selectByStorage := func(items []domain.RenewalItem, target float64) ([]domain.RenewalItem, float64, int) {
		must, normal := splitByWhitelist(items)
		selected := make([]domain.RenewalItem, 0, len(items))
		pickedStorage := 0.0
		pickedCores := 0
		for _, item := range must {
			selected = append(selected, item)
			pickedStorage += item.StorageCapacityTB
			pickedCores += item.CPULogicalCores
		}
		for _, item := range normal {
			if pickedStorage >= target {
				break
			}
			selected = append(selected, item)
			pickedStorage += item.StorageCapacityTB
			pickedCores += item.CPULogicalCores
		}
		return selected, pickedStorage, pickedCores
	}

	requiredComputeCores := maxInt(0, in.TargetCores-coveredComputeCores)
	requiredWarmStorage := maxFloat(0, in.WarmTargetStorageTB-coveredWarmStorage)
	requiredHotStorage := maxFloat(0, in.HotTargetStorageTB-coveredHotStorage)

	computeItems, computeCores := selectByCores(bucketItems["compute"], requiredComputeCores)
	warmItems, warmStorage, warmCores := selectByStorage(bucketItems["warm_storage"], requiredWarmStorage)
	hotItems, hotStorage, hotCores := selectByStorage(bucketItems["hot_storage"], requiredHotStorage)
	gpuItems := bucketItems["gpu"] // 全部续保（已应用环境过滤、到期过滤、blacklist）

	appendRankingNonRenewals := func(bucket string, all []domain.RenewalItem, selected []domain.RenewalItem) {
		selectedSet := make(map[string]bool, len(selected))
		for _, item := range selected {
			selectedSet[item.SN] = true
		}
		for i, item := range all {
			if selectedSet[item.SN] {
				continue
			}
			nonRenewalItems = append(nonRenewalItems, domain.NonRenewalItem{
				SN:           item.SN,
				Bucket:       bucket,
				Manufacturer: item.Manufacturer,
				Model:        item.Model,
				Environment:  item.Environment,
				IDC:          item.IDC,
				ConfigType:   item.ConfigType,
				PSA:          item.PSA,
				FinalScore:   item.FinalScore,
				ReasonCode:   "value_rank",
				Reason:       "价值分排名未入选",
				ReasonDetail: fmt.Sprintf("在 %s 栏目排名第 %d，目标已被更高分机器满足", bucket, i+1),
				RankInBucket: i + 1,
			})
		}
	}
	appendRankingNonRenewals("compute", bucketItems["compute"], computeItems)
	appendRankingNonRenewals("warm_storage", bucketItems["warm_storage"], warmItems)
	appendRankingNonRenewals("hot_storage", bucketItems["hot_storage"], hotItems)
	gpuCores := 0
	gpuStorage := 0.0
	gpuRenewalCards := 0
	for _, item := range gpuItems {
		gpuCores += item.CPULogicalCores
		gpuStorage += item.StorageCapacityTB
		gpuRenewalCards += item.GPUCardCount
	}

	sort.SliceStable(nonRenewalItems, func(i, j int) bool {
		if nonRenewalItems[i].ReasonCode != nonRenewalItems[j].ReasonCode {
			return nonRenewalItems[i].ReasonCode < nonRenewalItems[j].ReasonCode
		}
		if nonRenewalItems[i].RankInBucket != nonRenewalItems[j].RankInBucket {
			return nonRenewalItems[i].RankInBucket < nonRenewalItems[j].RankInBucket
		}
		return nonRenewalItems[i].SN < nonRenewalItems[j].SN
	})

	unmatchedConfigTypes := make([]string, 0, len(unmatchedConfigSet))
	for cfg := range unmatchedConfigSet {
		if strings.TrimSpace(cfg) == "" {
			continue
		}
		unmatchedConfigTypes = append(unmatchedConfigTypes, cfg)
	}
	sort.Strings(unmatchedConfigTypes)

	plan := domain.RenewalPlan{
		PlanID:               strconv.FormatInt(time.Now().Unix(), 10),
		TargetDate:           targetDate.Format("2006-01-02"),
		ExcludedEnvironments: excludedCanonical,
		ExcludedPSAs:         excludedPSACanonical,
		TargetCores:          in.TargetCores,
		WarmTargetStorageTB:  in.WarmTargetStorageTB,
		HotTargetStorageTB:   in.HotTargetStorageTB,
		CoveredComputeCores:  coveredComputeCores,
		CoveredWarmStorageTB: coveredWarmStorage,
		CoveredHotStorageTB:  coveredHotStorage,
		RequiredComputeCores: requiredComputeCores,
		RequiredWarmStorage:  requiredWarmStorage,
		RequiredHotStorage:   requiredHotStorage,
		UnmatchedConfigCount: len(unmatchedConfigTypes),
		UnmatchedConfigTypes: unmatchedConfigTypes,
		GPUCurrentCards:      gpuCurrentCards,
		GPUCurrentServers:    gpuCurrentServers,
		GPUCoveredCards:      gpuCoveredCards,
		GPUCoveredServers:    gpuCoveredServers,
		GPURenewalCards:      gpuRenewalCards,
		GPURenewalServers:    len(gpuItems),
		NonRenewalItems:      nonRenewalItems,
		Sections: []domain.RenewalPlanSection{
			{
				Bucket:        "compute",
				TargetCores:   in.TargetCores,
				CoveredCores:  coveredComputeCores,
				RequiredCores: requiredComputeCores,
				CoveredCount:  coveredComputeCount,
				SelectedCores: computeCores,
				SelectedCount: len(computeItems),
				Items:         computeItems,
			},
			{
				Bucket:            "warm_storage",
				TargetStorageTB:   in.WarmTargetStorageTB,
				CoveredStorageTB:  coveredWarmStorage,
				RequiredStorageTB: requiredWarmStorage,
				CoveredCount:      coveredWarmCount,
				SelectedStorageTB: warmStorage,
				SelectedCores:     warmCores,
				SelectedCount:     len(warmItems),
				Items:             warmItems,
			},
			{
				Bucket:            "hot_storage",
				TargetStorageTB:   in.HotTargetStorageTB,
				CoveredStorageTB:  coveredHotStorage,
				RequiredStorageTB: requiredHotStorage,
				CoveredCount:      coveredHotCount,
				SelectedStorageTB: hotStorage,
				SelectedCores:     hotCores,
				SelectedCount:     len(hotItems),
				Items:             hotItems,
			},
			{
				Bucket:            "gpu",
				CoveredCount:      gpuCoveredServers,
				SelectedStorageTB: gpuStorage,
				SelectedCores:     gpuCores,
				SelectedCount:     len(gpuItems),
				Items:             gpuItems,
			},
		},
	}

	appendSectionItems := func(items []domain.RenewalItem) {
		for _, item := range items {
			item.Rank = len(plan.Items) + 1
			plan.Items = append(plan.Items, item)
			plan.SelectedCores += item.CPULogicalCores
			plan.SelectedStorageTB += item.StorageCapacityTB
		}
	}
	appendSectionItems(computeItems)
	appendSectionItems(warmItems)
	appendSectionItems(hotItems)
	appendSectionItems(gpuItems)
	plan.SelectedCount = len(plan.Items)

	// 回写 rank 到 sections 内部 item
	for si := range plan.Sections {
		for ii := range plan.Sections[si].Items {
			if idx := findPlanItemIndexBySN(plan.Items, plan.Sections[si].Items[ii].SN); idx >= 0 {
				plan.Sections[si].Items[ii].Rank = plan.Items[idx].Rank
			}
		}
	}

	if err := s.renewalRepo.SavePlan(ctx, plan); err != nil {
		return domain.RenewalPlan{}, err
	}
	return plan, nil
}

func (s *RenewalService) GetPlan(ctx context.Context, planID string) (domain.RenewalPlan, error) {
	return s.renewalRepo.GetPlan(ctx, planID)
}

func (s *RenewalService) ListPlans(ctx context.Context, filter ListPlansFilter) ([]domain.RenewalPlan, error) {
	plans, err := s.renewalRepo.ListPlans(ctx)
	if err != nil {
		return nil, err
	}

	planID := strings.TrimSpace(filter.PlanID)
	excludedPSA := normalizeText(filter.ExcludedPSA)
	excludedEnv := normalizeText(filter.ExcludedEnvironment)

	var fromDate *time.Time
	if strings.TrimSpace(filter.TargetDateFrom) != "" {
		parsed, err := parseDate(filter.TargetDateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid target_date_from: %v", err)
		}
		fromDate = &parsed
	}
	var toDate *time.Time
	if strings.TrimSpace(filter.TargetDateTo) != "" {
		parsed, err := parseDate(filter.TargetDateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid target_date_to: %v", err)
		}
		toDate = &parsed
	}
	if fromDate != nil && toDate != nil && fromDate.After(*toDate) {
		return nil, fmt.Errorf("target_date_from must be <= target_date_to")
	}

	out := make([]domain.RenewalPlan, 0, len(plans))
	for _, p := range plans {
		if planID != "" && !strings.Contains(p.PlanID, planID) {
			continue
		}
		if excludedPSA != "" && !containsNormalized(p.ExcludedPSAs, excludedPSA) {
			continue
		}
		if excludedEnv != "" && !containsNormalized(p.ExcludedEnvironments, excludedEnv) {
			continue
		}
		if (fromDate != nil || toDate != nil) && strings.TrimSpace(p.TargetDate) != "" {
			d, err := parseDate(p.TargetDate)
			if err != nil {
				continue
			}
			if fromDate != nil && d.Before(*fromDate) {
				continue
			}
			if toDate != nil && d.After(*toDate) {
				continue
			}
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *RenewalService) DeletePlan(ctx context.Context, planID string) error {
	return s.renewalRepo.DeletePlan(ctx, planID)
}

func splitByWhitelist(items []domain.RenewalItem) (must []domain.RenewalItem, normal []domain.RenewalItem) {
	must = make([]domain.RenewalItem, 0)
	normal = make([]domain.RenewalItem, 0)
	for _, item := range items {
		if item.SpecialPolicy == "whitelist" {
			must = append(must, item)
		} else {
			normal = append(normal, item)
		}
	}
	return must, normal
}

func normalizeBucket(scene string) string {
	n := normalizeText(scene)
	switch n {
	case "计算型", "计算", "compute", "generalcompute", "通用计算", "cpu":
		return "compute"
	case "温存储", "温", "warmstorage", "warm", "coldstorage", "温储":
		return "warm_storage"
	case "热存储", "热", "hotstorage", "hot", "热储":
		return "hot_storage"
	case "gpu", "gpu型", "gpu计算", "gpucompute":
		return "gpu"
	default:
		// 兼容历史数据：未知分类默认归到计算型
		return "compute"
	}
}

func normalizeEnv(v string) string {
	return normalizeText(v)
}

func normalizeText(v string) string {
	n := strings.ToLower(strings.TrimSpace(v))
	n = strings.ReplaceAll(n, " ", "")
	n = strings.ReplaceAll(n, "_", "")
	n = strings.ReplaceAll(n, "-", "")
	return n
}

func parseDate(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	layouts := []string{"2006-01-02", "2006/01/02", "2006/1/2", "2006-1-2", "20060102"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %s", v)
}

func failureRateKey(configType, manufacturer, model string) string {
	return strings.Join([]string{normalizeText(configType), normalizeText(manufacturer), normalizeText(model)}, "|")
}

func findPlanItemIndexBySN(items []domain.RenewalItem, sn string) int {
	for i := range items {
		if items[i].SN == sn {
			return i
		}
	}
	return -1
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

type psaMatcher struct {
	patterns map[string]struct{}
}

func newPSAMatcher() *psaMatcher {
	return &psaMatcher{patterns: map[string]struct{}{}}
}

// AddNormalized adds an exclusion pattern and returns true when newly added.
// Matching is slash-segment aware prefix matching:
// pattern "/ss" matches "/ss", "/ss/st", "/ss/st/a";
// pattern "/a" does NOT match "/aa" or "/ab".
func (m *psaMatcher) AddNormalized(v string) bool {
	if _, ok := m.patterns[v]; ok {
		return false
	}
	m.patterns[v] = struct{}{}
	return true
}

func (m *psaMatcher) MatchRaw(raw string) bool {
	for _, token := range splitPSATokens(raw) {
		for pattern := range m.patterns {
			if token == pattern || strings.HasPrefix(token, pattern+"/") {
				return true
			}
		}
	}
	return false
}

func splitPSATokens(raw string) []string {
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
		n := normalizeText(p)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

func containsNormalized(list []string, target string) bool {
	for _, item := range list {
		if normalizeText(item) == target {
			return true
		}
	}
	return false
}
