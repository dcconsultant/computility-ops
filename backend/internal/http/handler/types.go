package handler

type CreatePlanReq struct {
	TargetDate           string   `json:"target_date" binding:"required"`
	ExcludedEnvironments []string `json:"excluded_environments"`
	ExcludedPSAs         []string `json:"excluded_psas"`
	TargetCores          int      `json:"target_cores" binding:"required,min=1"`
	WarmTargetStorageTB  float64  `json:"warm_target_storage_tb" binding:"required,min=0"`
	HotTargetStorageTB   float64  `json:"hot_target_storage_tb" binding:"required,min=0"`
}
