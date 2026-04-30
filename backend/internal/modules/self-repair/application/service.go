package application

import (
	"context"
	"fmt"

	srdomain "computility-ops/backend/internal/modules/self-repair/domain"
	"computility-ops/backend/internal/shared/kernel"
)

type CaseReader interface {
	ListCases(ctx context.Context) ([]srdomain.SelfRepairCase, error)
}

type Service struct {
	reader CaseReader
}

func NewService(reader CaseReader) *Service { return &Service{reader: reader} }

func (s *Service) ListSuggestions(ctx context.Context) ([]srdomain.SelfRepairSuggestion, error) {
	rows, err := s.reader.ListCases(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]srdomain.SelfRepairSuggestion, 0, len(rows))
	for _, c := range rows {
		risk := kernel.DecisionRiskMedium
		if c.SparePartAvailable && c.EngineerSkillMatched {
			risk = kernel.DecisionRiskLow
		}
		out = append(out, srdomain.SelfRepairSuggestion{
			Scenario:    "self-repair",
			SubjectID:   c.CaseID,
			SubjectType: "fault-case",
			Summary:     fmt.Sprintf("故障单%s可评估自维修", c.CaseID),
			Options: []kernel.DecisionOption{{
				OptionID:        "self-repair-standard",
				Title:           "执行标准自维修方案",
				Action:          "self_repair",
				RiskLevel:       risk,
				ExpectedBenefit: "缩短修复时长并降低外包成本",
				Evidence: []string{
					fmt.Sprintf("spare_part_available=%t", c.SparePartAvailable),
					fmt.Sprintf("engineer_skill_matched=%t", c.EngineerSkillMatched),
				},
			}},
		})
	}
	return out, nil
}
