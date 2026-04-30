package tests

import (
	"context"
	"testing"

	rpapp "computility-ops/backend/internal/modules/replacement-planning/application"
	rpdomain "computility-ops/backend/internal/modules/replacement-planning/domain"
)

type fakeReader struct{}

func (fakeReader) ListCandidates(ctx context.Context) ([]rpdomain.ReplacementCandidate, error) {
	_ = ctx
	return []rpdomain.ReplacementCandidate{{AssetID: "A1", AgeYears: 6, AnnualFailureRate: 0.12, AnnualMaintCost: 2000, CurrentTCO: 10000}}, nil
}

func TestListSuggestions(t *testing.T) {
	svc := rpapp.NewService(fakeReader{})
	rows, err := svc.ListSuggestions(context.Background())
	if err != nil {
		t.Fatalf("ListSuggestions err=%v", err)
	}
	if len(rows) != 1 || rows[0].SubjectID != "A1" {
		t.Fatalf("unexpected suggestions: %+v", rows)
	}
}
