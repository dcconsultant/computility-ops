package domain

type ScoringRules struct {
	MinAgeYears         float64 `yaml:"minAgeYears"`
	MinAnnualFailure    float64 `yaml:"minAnnualFailure"`
	ReplaceCostFactor   float64 `yaml:"replaceCostFactor"`
	AnnualSavingFactor  float64 `yaml:"annualSavingFactor"`
}

func DefaultScoringRules() ScoringRules {
	return ScoringRules{
		MinAgeYears:        3,
		MinAnnualFailure:   0.05,
		ReplaceCostFactor:  0.60,
		AnnualSavingFactor: 0.50,
	}
}
