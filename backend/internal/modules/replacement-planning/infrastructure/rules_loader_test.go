package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadScoringRules_FromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yml")
	content := []byte("scoring:\n  minAgeYears: 4\n  minAnnualFailure: 0.08\n  replaceCostFactor: 0.7\n  annualSavingFactor: 0.4\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write rules file: %v", err)
	}
	r, err := LoadScoringRules(path)
	if err != nil {
		t.Fatalf("LoadScoringRules err=%v", err)
	}
	if r.MinAgeYears != 4 || r.MinAnnualFailure != 0.08 || r.ReplaceCostFactor != 0.7 || r.AnnualSavingFactor != 0.4 {
		t.Fatalf("unexpected rules: %+v", r)
	}
}
