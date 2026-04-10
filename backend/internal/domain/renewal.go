package domain

type RenewalPlan struct {
	PlanID               string               `json:"plan_id"`
	TargetDate           string               `json:"target_date,omitempty"`
	ExcludedEnvironments []string             `json:"excluded_environments,omitempty"`
	ExcludedPSAs         []string             `json:"excluded_psas,omitempty"`
	TargetCores          int                  `json:"target_cores"`
	WarmTargetStorageTB  float64              `json:"warm_target_storage_tb"`
	HotTargetStorageTB   float64              `json:"hot_target_storage_tb"`
	CoveredComputeCores  int                  `json:"covered_compute_cores,omitempty"`
	CoveredWarmStorageTB float64              `json:"covered_warm_storage_tb,omitempty"`
	CoveredHotStorageTB  float64              `json:"covered_hot_storage_tb,omitempty"`
	RequiredComputeCores int                  `json:"required_compute_cores,omitempty"`
	RequiredWarmStorage  float64              `json:"required_warm_storage_tb,omitempty"`
	RequiredHotStorage   float64              `json:"required_hot_storage_tb,omitempty"`
	UnmatchedConfigCount int                  `json:"unmatched_config_count,omitempty"`
	UnmatchedConfigTypes []string             `json:"unmatched_config_types,omitempty"`
	GPUCurrentCards      int                  `json:"gpu_current_cards,omitempty"`
	GPUCurrentServers    int                  `json:"gpu_current_servers,omitempty"`
	GPUCoveredCards      int                  `json:"gpu_covered_cards,omitempty"`
	GPUCoveredServers    int                  `json:"gpu_covered_servers,omitempty"`
	GPURenewalCards      int                  `json:"gpu_renewal_cards,omitempty"`
	GPURenewalServers    int                  `json:"gpu_renewal_servers,omitempty"`
	SelectedCores        int                  `json:"selected_cores"`
	SelectedStorageTB    float64              `json:"selected_storage_tb"`
	SelectedCount        int                  `json:"selected_count"`
	Items                []RenewalItem        `json:"items"`
	NonRenewalItems      []NonRenewalItem     `json:"non_renewal_items,omitempty"`
	Sections             []RenewalPlanSection `json:"sections,omitempty"`
}

type RenewalPlanSection struct {
	Bucket            string        `json:"bucket"`
	TargetCores       int           `json:"target_cores,omitempty"`
	TargetStorageTB   float64       `json:"target_storage_tb,omitempty"`
	CoveredCores      int           `json:"covered_cores,omitempty"`
	CoveredStorageTB  float64       `json:"covered_storage_tb,omitempty"`
	RequiredCores     int           `json:"required_cores,omitempty"`
	RequiredStorageTB float64       `json:"required_storage_tb,omitempty"`
	CoveredCount      int           `json:"covered_count,omitempty"`
	SelectedCores     int           `json:"selected_cores,omitempty"`
	SelectedStorageTB float64       `json:"selected_storage_tb,omitempty"`
	SelectedCount     int           `json:"selected_count"`
	Items             []RenewalItem `json:"items"`
}

type RenewalItem struct {
	Rank                   int     `json:"rank"`
	Bucket                 string  `json:"bucket,omitempty"`
	SN                     string  `json:"sn"`
	Manufacturer           string  `json:"manufacturer"`
	Model                  string  `json:"model"`
	Environment            string  `json:"environment,omitempty"`
	ConfigType             string  `json:"config_type"`
	CPULogicalCores        int     `json:"cpu_logical_cores"`
	GPUCardCount           int     `json:"gpu_card_count,omitempty"`
	StorageCapacityTB      float64 `json:"storage_capacity_tb,omitempty"`
	PSA                    float64 `json:"psa"`
	ArchStandardizedFactor float64 `json:"arch_standardized_factor"`
	BaseScore              float64 `json:"base_score,omitempty"`
	AFROld                 float64 `json:"afr_old,omitempty"`
	AFRAvg                 float64 `json:"afr_avg,omitempty"`
	FailureAdjustFactor    float64 `json:"failure_adjust_factor,omitempty"`
	FinalScore             float64 `json:"final_score"`
	SpecialPolicy          string  `json:"special_policy,omitempty"`
}

type NonRenewalItem struct {
	SN           string  `json:"sn"`
	Bucket       string  `json:"bucket,omitempty"`
	Manufacturer string  `json:"manufacturer,omitempty"`
	Model        string  `json:"model,omitempty"`
	Environment  string  `json:"environment,omitempty"`
	ConfigType   string  `json:"config_type,omitempty"`
	PSA          float64 `json:"psa,omitempty"`
	FinalScore   float64 `json:"final_score,omitempty"`
	ReasonCode   string  `json:"reason_code"`
	Reason       string  `json:"reason"`
	ReasonDetail string  `json:"reason_detail,omitempty"`
	RankInBucket int     `json:"rank_in_bucket,omitempty"`
}
