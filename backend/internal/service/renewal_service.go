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

type RenewalService struct {
	serverRepo  repository.ServerRepo
	datasetRepo repository.DatasetRepo
	renewalRepo repository.RenewalPlanRepo
}

func NewRenewalService(serverRepo repository.ServerRepo, datasetRepo repository.DatasetRepo, renewalRepo repository.RenewalPlanRepo) *RenewalService {
	return &RenewalService{serverRepo: serverRepo, datasetRepo: datasetRepo, renewalRepo: renewalRepo}
}

func (s *RenewalService) CreatePlan(ctx context.Context, targetCores int) (domain.RenewalPlan, error) {
	if targetCores <= 0 {
		return domain.RenewalPlan{}, fmt.Errorf("target_cores must be > 0")
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

	pkgMap := map[string]domain.HostPackageConfig{}
	for _, p := range packages {
		pkgMap[strings.TrimSpace(p.ConfigType)] = p
	}
	specialMap := map[string]string{}
	for _, sp := range specialRules {
		if sp.SN == "" {
			continue
		}
		specialMap[sp.SN] = sp.Policy
	}

	mustRenew := make([]domain.RenewalItem, 0)
	normal := make([]domain.RenewalItem, 0)
	for _, srv := range servers {
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
		item := domain.RenewalItem{
			SN:                     srv.SN,
			Manufacturer:           srv.Manufacturer,
			Model:                  srv.Model,
			ConfigType:             srv.ConfigType,
			CPULogicalCores:        cores,
			PSA:                    srv.PSA,
			ArchStandardizedFactor: coef,
			FinalScore:             srv.PSA * coef,
		}
		switch specialMap[srv.SN] {
		case "blacklist":
			continue
		case "whitelist":
			item.SpecialPolicy = "whitelist"
			mustRenew = append(mustRenew, item)
		default:
			normal = append(normal, item)
		}
	}

	sortItems := func(items []domain.RenewalItem) {
		sort.Slice(items, func(i, j int) bool {
			if items[i].FinalScore != items[j].FinalScore {
				return items[i].FinalScore > items[j].FinalScore
			}
			if items[i].CPULogicalCores != items[j].CPULogicalCores {
				return items[i].CPULogicalCores > items[j].CPULogicalCores
			}
			return items[i].SN < items[j].SN
		})
	}
	sortItems(mustRenew)
	sortItems(normal)

	plan := domain.RenewalPlan{PlanID: strconv.FormatInt(time.Now().Unix(), 10), TargetCores: targetCores}
	appendItem := func(item domain.RenewalItem) {
		item.Rank = len(plan.Items) + 1
		plan.Items = append(plan.Items, item)
		plan.SelectedCores += item.CPULogicalCores
	}
	for _, item := range mustRenew {
		appendItem(item)
	}
	for _, item := range normal {
		if plan.SelectedCores >= targetCores {
			break
		}
		appendItem(item)
	}
	plan.SelectedCount = len(plan.Items)

	if err := s.renewalRepo.SavePlan(ctx, plan); err != nil {
		return domain.RenewalPlan{}, err
	}
	return plan, nil
}

func (s *RenewalService) GetPlan(ctx context.Context, planID string) (domain.RenewalPlan, error) {
	return s.renewalRepo.GetPlan(ctx, planID)
}
