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

type CreatePlanInput struct {
	TargetDate           string
	ExcludedEnvironments []string
	TargetCores          int
	WarmTargetStorageTB  float64
	HotTargetStorageTB   float64
}

type RenewalService struct {
	serverRepo  repository.ServerRepo
	datasetRepo repository.DatasetRepo
	renewalRepo repository.RenewalPlanRepo
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
	packageModelRates, err := s.datasetRepo.ListPackageModelFailureRates(ctx)
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

	pkgMap := map[string]domain.HostPackageConfig{}
	for _, p := range packages {
		pkgMap[strings.TrimSpace(p.ConfigType)] = p
	}

	pkgAFRAvg := map[string]float64{}
	for _, p := range packageRates {
		pkgAFRAvg[strings.TrimSpace(p.ConfigType)] = p.FailureRate
	}

	pkgModelAFROld := map[string]float64{}
	for _, p := range packageModelRates {
		key := failureRateKey(p.ConfigType, p.Manufacturer, p.Model)
		pkgModelAFROld[key] = p.FailureRate
	}

	specialMap := map[string]string{}
	for _, sp := range specialRules {
		if sp.SN == "" {
			continue
		}
		specialMap[sp.SN] = sp.Policy
	}

	bucketItems := map[string][]domain.RenewalItem{
		"compute":      {},
		"warm_storage": {},
		"hot_storage":  {},
		"gpu":          {},
	}

	for _, srv := range servers {
		if excludedSet[normalizeEnv(srv.Environment)] {
			continue
		}
		if strings.TrimSpace(srv.WarrantyEndDate) == "" {
			continue
		}
		wd, err := parseDate(srv.WarrantyEndDate)
		if err != nil {
			return domain.RenewalPlan{}, fmt.Errorf("invalid warranty_end_date for sn=%s: %v", srv.SN, err)
		}
		if !wd.Before(targetDate) {
			continue
		}

		pkg, ok := pkgMap[srv.ConfigType]
		if !ok {
			return domain.RenewalPlan{}, fmt.Errorf("missing package config for config_type=%s", srv.ConfigType)
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
		baseScore := srv.PSA * coef
		finalScore := baseScore

		item := domain.RenewalItem{
			SN:                     srv.SN,
			Bucket:                 bucket,
			Manufacturer:           srv.Manufacturer,
			Model:                  srv.Model,
			Environment:            srv.Environment,
			ConfigType:             srv.ConfigType,
			CPULogicalCores:        cores,
			StorageCapacityTB:      pkg.StorageCapacityTB,
			PSA:                    srv.PSA,
			ArchStandardizedFactor: coef,
			BaseScore:              baseScore,
			FinalScore:             finalScore,
		}

		if bucket == "warm_storage" || bucket == "hot_storage" {
			afrAvg := pkgAFRAvg[strings.TrimSpace(srv.ConfigType)]
			afrOld := pkgModelAFROld[failureRateKey(srv.ConfigType, srv.Manufacturer, srv.Model)]
			if afrAvg > 0 && afrOld > 0 {
				factor := math.Exp(-(afrOld/afrAvg - 1))
				item.AFRAvg = afrAvg
				item.AFROld = afrOld
				item.FailureAdjustFactor = factor
				item.FinalScore = baseScore * factor
			}
		}

		specialPolicy := specialMap[srv.SN]
		if specialPolicy == "blacklist" {
			continue
		}
		if specialPolicy == "whitelist" {
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

	computeItems, computeCores := selectByCores(bucketItems["compute"], in.TargetCores)
	warmItems, warmStorage, warmCores := selectByStorage(bucketItems["warm_storage"], in.WarmTargetStorageTB)
	hotItems, hotStorage, hotCores := selectByStorage(bucketItems["hot_storage"], in.HotTargetStorageTB)
	gpuItems := bucketItems["gpu"] // 全部续保（已应用环境过滤、到期过滤、blacklist）
	gpuCores := 0
	gpuStorage := 0.0
	for _, item := range gpuItems {
		gpuCores += item.CPULogicalCores
		gpuStorage += item.StorageCapacityTB
	}

	plan := domain.RenewalPlan{
		PlanID:               strconv.FormatInt(time.Now().Unix(), 10),
		TargetDate:           targetDate.Format("2006-01-02"),
		ExcludedEnvironments: excludedCanonical,
		TargetCores:          in.TargetCores,
		WarmTargetStorageTB:  in.WarmTargetStorageTB,
		HotTargetStorageTB:   in.HotTargetStorageTB,
		Sections: []domain.RenewalPlanSection{
			{
				Bucket:        "compute",
				TargetCores:   in.TargetCores,
				SelectedCores: computeCores,
				SelectedCount: len(computeItems),
				Items:         computeItems,
			},
			{
				Bucket:            "warm_storage",
				TargetStorageTB:   in.WarmTargetStorageTB,
				SelectedStorageTB: warmStorage,
				SelectedCores:     warmCores,
				SelectedCount:     len(warmItems),
				Items:             warmItems,
			},
			{
				Bucket:            "hot_storage",
				TargetStorageTB:   in.HotTargetStorageTB,
				SelectedStorageTB: hotStorage,
				SelectedCores:     hotCores,
				SelectedCount:     len(hotItems),
				Items:             hotItems,
			},
			{
				Bucket:            "gpu",
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
