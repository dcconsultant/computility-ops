package domain

type RenewalTargetMode string

const (
	RenewalTargetModeManual   RenewalTargetMode = "manual"
	RenewalTargetModeMaximize RenewalTargetMode = "maximize"
)

type RenewalSceneTarget struct {
	Mode   RenewalTargetMode `json:"mode"`
	Target float64           `json:"target"`
}

type RenewalRegionTargets struct {
	Compute     RenewalSceneTarget `json:"compute"`
	WarmStorage RenewalSceneTarget `json:"warm_storage"`
	HotStorage  RenewalSceneTarget `json:"hot_storage"`
	GPU         RenewalSceneTarget `json:"gpu"`
}

type RenewalRequirements struct {
	Domestic RenewalRegionTargets `json:"domestic"`
	India    RenewalRegionTargets `json:"india"`
}

type RenewalPlanSettings struct {
	TargetDate           string              `json:"target_date"`
	ExcludedEnvironments []string            `json:"excluded_environments"`
	ExcludedPSAs         []string            `json:"excluded_psas"`
	Requirements         RenewalRequirements `json:"requirements"`
	DomesticBudget       float64             `json:"domestic_budget"`
	IndiaBudget          float64             `json:"india_budget"`
}

type RenewalUnitPrice struct {
	Country       string  `json:"country"`
	SceneCategory string  `json:"scene_category"`
	UnitPrice     float64 `json:"unit_price"`
}

type RenewalPlan struct {
	PlanID               string           `json:"plan_id"`
	TargetDate           string           `json:"target_date,omitempty"`
	ExcludedEnvironments []string         `json:"excluded_environments,omitempty"`
	ExcludedPSAs         []string         `json:"excluded_psas,omitempty"`
	TargetCores          int              `json:"target_cores"`
	WarmTargetStorageTB  float64          `json:"warm_target_storage_tb"`
	HotTargetStorageTB   float64          `json:"hot_target_storage_tb"`
	DomesticBudget       float64          `json:"domestic_budget,omitempty"`
	IndiaBudget          float64          `json:"india_budget,omitempty"`
	Requirements         RenewalRequirements `json:"requirements,omitempty"`
	SelectedCores        int              `json:"selected_cores"`
	SelectedStorageTB    float64          `json:"selected_storage_tb"`
	SelectedCount        int              `json:"selected_count"`
}
