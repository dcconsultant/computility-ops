package domain

type RenewalPlan struct {
	PlanID        string        `json:"plan_id"`
	TargetCores   int           `json:"target_cores"`
	SelectedCores int           `json:"selected_cores"`
	SelectedCount int           `json:"selected_count"`
	Items         []RenewalItem `json:"items"`
}

type RenewalItem struct {
	Rank                   int     `json:"rank"`
	SN                     string  `json:"sn"`
	Manufacturer           string  `json:"manufacturer"`
	Model                  string  `json:"model"`
	ConfigType             string  `json:"config_type"`
	CPULogicalCores        int     `json:"cpu_logical_cores"`
	PSA                    float64 `json:"psa"`
	ArchStandardizedFactor float64 `json:"arch_standardized_factor"`
	FinalScore             float64 `json:"final_score"`
	SpecialPolicy          string  `json:"special_policy,omitempty"`
}
