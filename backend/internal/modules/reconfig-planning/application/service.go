package application

import (
	"context"
	"fmt"

	rcdomain "computility-ops/backend/internal/modules/reconfig-planning/domain"
	"computility-ops/backend/internal/shared/kernel"
)

type CandidateReader interface {
	ListCandidates(ctx context.Context) ([]rcdomain.ReconfigCandidate, error)
}

type Service struct {
	reader CandidateReader
}

func NewService(reader CandidateReader) *Service { return &Service{reader: reader} }

func (s *Service) ListSuggestions(ctx context.Context) ([]rcdomain.ReconfigSuggestion, error) {
	rows, err := s.reader.ListCandidates(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]rcdomain.ReconfigSuggestion, 0, len(rows))
	for _, c := range rows {
		out = append(out, rcdomain.ReconfigSuggestion{
			Scenario:    "reconfig-planning",
			SubjectID:   c.AssetID,
			SubjectType: "asset",
			Summary:     fmt.Sprintf("资产%s建议进行改配评估", c.AssetID),
			Options: []kernel.DecisionOption{{
				OptionID:        "resize-config",
				Title:           "下调冗余配置",
				Action:          "reconfigure",
				RiskLevel:       kernel.DecisionRiskLow,
				ExpectedBenefit: "降低资源浪费与月成本",
				EstimatedSavings: c.MonthlyCost * 0.2,
				Evidence: []string{
					fmt.Sprintf("avg_cpu=%.2f", c.AvgCPUUsage),
					fmt.Sprintf("avg_mem=%.2f", c.AvgMemUsage),
				},
			}},
		})
	}
	return out, nil
}
