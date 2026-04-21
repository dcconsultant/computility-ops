package handler

import (
	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
)

type CreatePlanReq struct {
	TargetDate           string   `json:"target_date" binding:"required"`
	ExcludedEnvironments []string `json:"excluded_environments"`
	ExcludedPSAs         []string `json:"excluded_psas"`
	TargetCores          int      `json:"target_cores" binding:"required,min=1"`
	WarmTargetStorageTB  float64  `json:"warm_target_storage_tb" binding:"required,min=0"`
	HotTargetStorageTB   float64  `json:"hot_target_storage_tb" binding:"required,min=0"`
	DomesticBudget       float64  `json:"domestic_budget" binding:"min=0"`
	IndiaBudget          float64  `json:"india_budget" binding:"min=0"`
}

type ListPlansReq struct {
	PlanID              string `form:"plan_id"`
	TargetDateFrom      string `form:"target_date_from"`
	TargetDateTo        string `form:"target_date_to"`
	ExcludedPSA         string `form:"excluded_psa"`
	ExcludedEnvironment string `form:"excluded_environment"`
}

type UpdateRenewalUnitPricesReq struct {
	Prices []domain.RenewalUnitPrice `json:"prices" binding:"required,min=1"`
}

type ExportYearFaultAnalysisReq struct {
	Year int                        `json:"year"`
	Rows []service.FaultAnalysisRow `json:"rows" binding:"required"`
}
