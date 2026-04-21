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
