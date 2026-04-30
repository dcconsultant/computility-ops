package kernel

type DecisionRiskLevel string

const (
	DecisionRiskLow    DecisionRiskLevel = "low"
	DecisionRiskMedium DecisionRiskLevel = "medium"
	DecisionRiskHigh   DecisionRiskLevel = "high"
)

type DecisionOption struct {
	OptionID         string            `json:"option_id"`
	Title            string            `json:"title"`
	Action           string            `json:"action"`
	ExpectedBenefit  string            `json:"expected_benefit,omitempty"`
	EstimatedROI     string            `json:"estimated_roi,omitempty"`
	RiskLevel        DecisionRiskLevel `json:"risk_level"`
	RiskNotes        []string          `json:"risk_notes,omitempty"`
	Evidence         []string          `json:"evidence,omitempty"`
	EstimatedCost    float64           `json:"estimated_cost,omitempty"`
	EstimatedSavings float64           `json:"estimated_savings,omitempty"`
}

type DecisionSuggestion struct {
	Scenario    string           `json:"scenario"`
	SubjectID   string           `json:"subject_id"`
	SubjectType string           `json:"subject_type"`
	Summary     string           `json:"summary"`
	Options     []DecisionOption `json:"options"`
}
