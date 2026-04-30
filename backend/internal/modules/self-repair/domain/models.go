package domain

import "computility-ops/backend/internal/shared/kernel"

type SelfRepairCase struct {
	CaseID               string `json:"case_id"`
	AssetID              string `json:"asset_id"`
	FaultCategory        string `json:"fault_category,omitempty"`
	SparePartAvailable   bool   `json:"spare_part_available"`
	EngineerSkillMatched bool   `json:"engineer_skill_matched"`
}

type SelfRepairSuggestion = kernel.DecisionSuggestion
