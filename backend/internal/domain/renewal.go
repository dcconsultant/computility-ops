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
	SelectedCores        int                  `json:"selected_cores"`
	SelectedStorageTB    float64              `json:"selected_storage_tb"`
	SelectedCount        int                  `json:"selected_count"`
	Items                []RenewalItem        `json:"items"`
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
