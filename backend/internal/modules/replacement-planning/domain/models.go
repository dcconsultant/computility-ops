package domain

import "computility-ops/backend/internal/shared/kernel"

type ReplacementCandidate struct {
	AssetID           string  `json:"asset_id"`
	Model             string  `json:"model,omitempty"`
	AgeYears          float64 `json:"age_years,omitempty"`
	AnnualFailureRate float64 `json:"annual_failure_rate,omitempty"`
	AnnualMaintCost   float64 `json:"annual_maintenance_cost,omitempty"`
	CurrentTCO        float64 `json:"current_tco,omitempty"`
}

type ReplacementSuggestion = kernel.DecisionSuggestion
