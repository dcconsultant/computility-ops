package infrastructure

import (
	"os"

	rpdomain "computility-ops/backend/internal/modules/replacement-planning/domain"
	"gopkg.in/yaml.v3"
)

type rulesFile struct {
	Scoring rpdomain.ScoringRules `yaml:"scoring"`
}

func LoadScoringRules(path string) (rpdomain.ScoringRules, error) {
	if path == "" {
		return rpdomain.DefaultScoringRules(), nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return rpdomain.ScoringRules{}, err
	}
	var cfg rulesFile
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return rpdomain.ScoringRules{}, err
	}
	rules := cfg.Scoring
	if rules.ReplaceCostFactor <= 0 {
		rules.ReplaceCostFactor = rpdomain.DefaultScoringRules().ReplaceCostFactor
	}
	if rules.AnnualSavingFactor <= 0 {
		rules.AnnualSavingFactor = rpdomain.DefaultScoringRules().AnnualSavingFactor
	}
	return rules, nil
}
