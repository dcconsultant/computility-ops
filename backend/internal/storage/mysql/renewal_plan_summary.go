package mysql

import "computility-ops/backend/internal/domain"

type renewalPlanListPayload struct {
	PlanID               string                      `json:"plan_id"`
	Status               string                      `json:"status,omitempty"`
	EffectiveAt          string                      `json:"effective_at,omitempty"`
	TargetDate           string                      `json:"target_date,omitempty"`
	ExcludedEnvironments []string                    `json:"excluded_environments,omitempty"`
	ExcludedPSAs         []string                    `json:"excluded_psas,omitempty"`
	IdleStoppedPSAs      []string                    `json:"idle_stopped_psas,omitempty"`
	TargetCores          int                         `json:"target_cores"`
	WarmTargetStorageTB  float64                     `json:"warm_target_storage_tb"`
	HotTargetStorageTB   float64                     `json:"hot_target_storage_tb"`
	DomesticBudget       float64                     `json:"domestic_budget,omitempty"`
	IndiaBudget          float64                     `json:"india_budget,omitempty"`
	TotalServersNoPSA    int                         `json:"total_servers_no_psa,omitempty"`
	DomesticServersNoPSA int                         `json:"domestic_servers_no_psa,omitempty"`
	IndiaServersNoPSA    int                         `json:"india_servers_no_psa,omitempty"`
	Requirements         domain.RenewalRequirements  `json:"requirements,omitempty"`
	CoveredComputeCores  int                         `json:"covered_compute_cores,omitempty"`
	CoveredWarmStorageTB float64                     `json:"covered_warm_storage_tb,omitempty"`
	CoveredHotStorageTB  float64                     `json:"covered_hot_storage_tb,omitempty"`
	RequiredComputeCores int                         `json:"required_compute_cores,omitempty"`
	RequiredWarmStorage  float64                     `json:"required_warm_storage_tb,omitempty"`
	RequiredHotStorage   float64                     `json:"required_hot_storage_tb,omitempty"`
	UnmatchedConfigCount int                         `json:"unmatched_config_count,omitempty"`
	UnmatchedConfigTypes []string                    `json:"unmatched_config_types,omitempty"`
	GPUCurrentCards      int                         `json:"gpu_current_cards,omitempty"`
	GPUCurrentServers    int                         `json:"gpu_current_servers,omitempty"`
	GPUCoveredCards      int                         `json:"gpu_covered_cards,omitempty"`
	GPUCoveredServers    int                         `json:"gpu_covered_servers,omitempty"`
	GPURenewalCards      int                         `json:"gpu_renewal_cards,omitempty"`
	GPURenewalServers    int                         `json:"gpu_renewal_servers,omitempty"`
	SelectedCores        int                         `json:"selected_cores"`
	SelectedStorageTB    float64                     `json:"selected_storage_tb"`
	SelectedCount        int                         `json:"selected_count"`
	Sections             []renewalPlanSectionSummary `json:"sections,omitempty"`
	MinimalRenewalError  string                      `json:"minimal_renewal_error,omitempty"`
}

type renewalPlanSectionSummary struct {
	Bucket            string  `json:"bucket"`
	TargetCores       int     `json:"target_cores,omitempty"`
	TargetStorageTB   float64 `json:"target_storage_tb,omitempty"`
	CoveredCores      int     `json:"covered_cores,omitempty"`
	CoveredStorageTB  float64 `json:"covered_storage_tb,omitempty"`
	RequiredCores     int     `json:"required_cores,omitempty"`
	RequiredStorageTB float64 `json:"required_storage_tb,omitempty"`
	CoveredCount      int     `json:"covered_count,omitempty"`
	SelectedCores     int     `json:"selected_cores,omitempty"`
	SelectedStorageTB float64 `json:"selected_storage_tb,omitempty"`
	SelectedCount     int     `json:"selected_count"`
}

func unmarshalRenewalPlanListSummary(payload string) (domain.RenewalPlan, error) {
	var in renewalPlanListPayload
	if err := unmarshalJSONPayload(payload, &in); err != nil {
		return domain.RenewalPlan{}, err
	}
	out := domain.RenewalPlan{
		PlanID:               in.PlanID,
		Status:               in.Status,
		EffectiveAt:          in.EffectiveAt,
		TargetDate:           in.TargetDate,
		ExcludedEnvironments: in.ExcludedEnvironments,
		ExcludedPSAs:         in.ExcludedPSAs,
		IdleStoppedPSAs:      in.IdleStoppedPSAs,
		TargetCores:          in.TargetCores,
		WarmTargetStorageTB:  in.WarmTargetStorageTB,
		HotTargetStorageTB:   in.HotTargetStorageTB,
		DomesticBudget:       in.DomesticBudget,
		IndiaBudget:          in.IndiaBudget,
		TotalServersNoPSA:    in.TotalServersNoPSA,
		DomesticServersNoPSA: in.DomesticServersNoPSA,
		IndiaServersNoPSA:    in.IndiaServersNoPSA,
		Requirements:         in.Requirements,
		CoveredComputeCores:  in.CoveredComputeCores,
		CoveredWarmStorageTB: in.CoveredWarmStorageTB,
		CoveredHotStorageTB:  in.CoveredHotStorageTB,
		RequiredComputeCores: in.RequiredComputeCores,
		RequiredWarmStorage:  in.RequiredWarmStorage,
		RequiredHotStorage:   in.RequiredHotStorage,
		UnmatchedConfigCount: in.UnmatchedConfigCount,
		UnmatchedConfigTypes: in.UnmatchedConfigTypes,
		GPUCurrentCards:      in.GPUCurrentCards,
		GPUCurrentServers:    in.GPUCurrentServers,
		GPUCoveredCards:      in.GPUCoveredCards,
		GPUCoveredServers:    in.GPUCoveredServers,
		GPURenewalCards:      in.GPURenewalCards,
		GPURenewalServers:    in.GPURenewalServers,
		SelectedCores:        in.SelectedCores,
		SelectedStorageTB:    in.SelectedStorageTB,
		SelectedCount:        in.SelectedCount,
		MinimalRenewalError:  in.MinimalRenewalError,
	}
	if len(in.Sections) > 0 {
		out.Sections = make([]domain.RenewalPlanSection, 0, len(in.Sections))
		for _, sec := range in.Sections {
			out.Sections = append(out.Sections, domain.RenewalPlanSection{
				Bucket:            sec.Bucket,
				TargetCores:       sec.TargetCores,
				TargetStorageTB:   sec.TargetStorageTB,
				CoveredCores:      sec.CoveredCores,
				CoveredStorageTB:  sec.CoveredStorageTB,
				RequiredCores:     sec.RequiredCores,
				RequiredStorageTB: sec.RequiredStorageTB,
				CoveredCount:      sec.CoveredCount,
				SelectedCores:     sec.SelectedCores,
				SelectedStorageTB: sec.SelectedStorageTB,
				SelectedCount:     sec.SelectedCount,
			})
		}
	}
	return out, nil
}
