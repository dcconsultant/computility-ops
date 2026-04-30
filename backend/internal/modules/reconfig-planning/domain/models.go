package domain

import "computility-ops/backend/internal/shared/kernel"

type ReconfigCandidate struct {
	AssetID       string  `json:"asset_id"`
	CurrentCPU    int     `json:"current_cpu,omitempty"`
	CurrentMemory int     `json:"current_memory_gb,omitempty"`
	AvgCPUUsage   float64 `json:"avg_cpu_usage,omitempty"`
	AvgMemUsage   float64 `json:"avg_memory_usage,omitempty"`
	MonthlyCost   float64 `json:"monthly_cost,omitempty"`
}

type ReconfigSuggestion = kernel.DecisionSuggestion
