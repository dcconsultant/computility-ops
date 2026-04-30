package application

import (
	"context"
	"fmt"

	rpdomain "computility-ops/backend/internal/modules/replacement-planning/domain"
	"computility-ops/backend/internal/shared/kernel"
)

type CandidateReader interface {
	ListCandidates(ctx context.Context) ([]rpdomain.ReplacementCandidate, error)
}

type Service struct {
	reader CandidateReader
	rules  rpdomain.ScoringRules
}

func NewService(reader CandidateReader) *Service {
	return &Service{reader: reader, rules: rpdomain.DefaultScoringRules()}
}

func NewServiceWithRules(reader CandidateReader, rules rpdomain.ScoringRules) *Service {
	if rules.ReplaceCostFactor <= 0 || rules.AnnualSavingFactor <= 0 {
		rules = rpdomain.DefaultScoringRules()
	}
	return &Service{reader: reader, rules: rules}
}

func (s *Service) ListSuggestions(ctx context.Context) ([]rpdomain.ReplacementSuggestion, error) {
	candidates, err := s.reader.ListCandidates(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]rpdomain.ReplacementSuggestion, 0, len(candidates))
	for _, c := range candidates {
		if c.AgeYears < s.rules.MinAgeYears && c.AnnualFailureRate < s.rules.MinAnnualFailure {
			continue
		}
		out = append(out, rpdomain.ReplacementSuggestion{
			Scenario:    "replacement-planning",
			SubjectID:   c.AssetID,
			SubjectType: "asset",
			Summary:     fmt.Sprintf("资产%s建议进入替换评估", c.AssetID),
			Options: []kernel.DecisionOption{{
				OptionID:         "replace-new-model",
				Title:            "替换为新机型",
				Action:           "replace",
				RiskLevel:        kernel.DecisionRiskMedium,
				ExpectedBenefit:  "降低故障与维护成本",
				EstimatedCost:    c.CurrentTCO * s.rules.ReplaceCostFactor,
				EstimatedSavings: c.AnnualMaintCost * s.rules.AnnualSavingFactor,
				Evidence: []string{
					fmt.Sprintf("age_years=%.1f", c.AgeYears),
					fmt.Sprintf("afr=%.4f", c.AnnualFailureRate),
				},
			}},
		})
	}
	return out, nil
}
