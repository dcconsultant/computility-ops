package handler

import (
	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
)

type CreatePlanReq struct {
	TargetDate           string                     `json:"target_date" binding:"required"`
	ExcludedEnvironments []string                   `json:"excluded_environments"`
	ExcludedPSAs         []string                   `json:"excluded_psas"`
	TargetCores          int                        `json:"target_cores"`
	WarmTargetStorageTB  float64                    `json:"warm_target_storage_tb"`
	HotTargetStorageTB   float64                    `json:"hot_target_storage_tb"`
	DomesticBudget       float64                    `json:"domestic_budget" binding:"min=0"`
	IndiaBudget          float64                    `json:"india_budget" binding:"min=0"`
	Requirements         domain.RenewalRequirements `json:"requirements"`
}

type UpdateRenewalSettingsReq struct {
	TargetDate           string                     `json:"target_date" binding:"required"`
	ExcludedEnvironments []string                   `json:"excluded_environments"`
	ExcludedPSAs         []string                   `json:"excluded_psas"`
	Requirements         domain.RenewalRequirements `json:"requirements"`
	DomesticBudget       float64                    `json:"domestic_budget" binding:"min=0"`
	IndiaBudget          float64                    `json:"india_budget" binding:"min=0"`
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

type CreateContractReq struct {
	ContractName    string  `json:"contract_name" binding:"required"`
	PeriodStart     string  `json:"period_start" binding:"required"`
	PeriodEnd       string  `json:"period_end" binding:"required"`
	PreTaxAmount    float64 `json:"pre_tax_amount" binding:"min=0"`
	Supplier        string  `json:"supplier" binding:"required"`
	BusinessContact string  `json:"business_contact" binding:"required"`
	TechContact     string  `json:"tech_contact" binding:"required"`
}

type UpdateContractReq struct {
	ContractName    string  `json:"contract_name" binding:"required"`
	PeriodStart     string  `json:"period_start" binding:"required"`
	PeriodEnd       string  `json:"period_end" binding:"required"`
	PreTaxAmount    float64 `json:"pre_tax_amount" binding:"min=0"`
	Supplier        string  `json:"supplier" binding:"required"`
	BusinessContact string  `json:"business_contact" binding:"required"`
	TechContact     string  `json:"tech_contact" binding:"required"`
}

type CreateMetaModelReq struct {
	ModelCode   string `json:"model_code" binding:"required"`
	ModelName   string `json:"model_name" binding:"required"`
	Description string `json:"description"`
}

type UpdateMetaModelReq struct {
	ModelName   string `json:"model_name" binding:"required"`
	Description string `json:"description"`
}

type MetaEnumOptionReq struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled"`
}

type CreateMetaFieldReq struct {
	FieldCode      string              `json:"field_code" binding:"required"`
	FieldName      string              `json:"field_name" binding:"required"`
	Category       string              `json:"category"`
	ValueType      string              `json:"value_type" binding:"required"`
	Required       bool                `json:"required"`
	Unique         bool                `json:"unique"`
	Filterable     bool                `json:"filterable"`
	Sortable       bool                `json:"sortable"`
	Visible        bool                `json:"visible"`
	DefaultValue   string              `json:"default_value"`
	ValidationRule string              `json:"validation_rule"`
	EnumOptions    []MetaEnumOptionReq `json:"enum_options"`
}

type UpdateMetaFieldReq struct {
	FieldName      string              `json:"field_name" binding:"required"`
	Category       string              `json:"category"`
	ValueType      string              `json:"value_type" binding:"required"`
	Required       bool                `json:"required"`
	Unique         bool                `json:"unique"`
	Filterable     bool                `json:"filterable"`
	Sortable       bool                `json:"sortable"`
	Visible        bool                `json:"visible"`
	DefaultValue   string              `json:"default_value"`
	ValidationRule string              `json:"validation_rule"`
	EnumOptions    []MetaEnumOptionReq `json:"enum_options"`
}

type ReorderMetaFieldsReq struct {
	FieldIDs []string `json:"field_ids" binding:"required,min=1"`
}

type CreateMetaReferenceReq struct {
	SourceFieldID  string   `json:"source_field_id" binding:"required"`
	TargetModelID  string   `json:"target_model_id" binding:"required"`
	TargetFieldID  string   `json:"target_field_id" binding:"required"`
	DisplayFields  []string `json:"display_fields"`
	OnDeleteAction string   `json:"on_delete_action"`
}

type UpdateMetaReferenceReq = CreateMetaReferenceReq

type PublishMetaModelReq struct {
	ChangeSummary string `json:"change_summary"`
	PublishedBy   string `json:"published_by"`
}

type RollbackMetaModelReq struct {
	Version int `json:"version" binding:"required,min=1"`
}

type CloneMetaModelReq struct {
	ModelCode   string `json:"model_code" binding:"required"`
	ModelName   string `json:"model_name" binding:"required"`
	Description string `json:"description"`
}

type UpsertMetaRecordReq struct {
	Data map[string]any `json:"data" binding:"required"`
}

type UpsertSupplierReq struct {
	CompanyFullName   string `json:"company_full_name" binding:"required"`
	TaxNumber         string `json:"tax_number" binding:"required"`
	ProjectOwner      string `json:"project_owner"`
	ProjectOwnerPhone string `json:"project_owner_phone"`
	TechContact       string `json:"tech_contact"`
	TechContactPhone  string `json:"tech_contact_phone"`
	BusinessScope     string `json:"business_scope" binding:"required"`
}
