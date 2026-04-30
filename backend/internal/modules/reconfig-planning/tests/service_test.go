package tests

import (
	"context"
	"testing"

	rcapp "computility-ops/backend/internal/modules/reconfig-planning/application"
	rcdomain "computility-ops/backend/internal/modules/reconfig-planning/domain"
)

type fakeReader struct{}

func (fakeReader) ListCandidates(ctx context.Context) ([]rcdomain.ReconfigCandidate, error) {
	_ = ctx
	return []rcdomain.ReconfigCandidate{{AssetID: "A2", AvgCPUUsage: 0.2, AvgMemUsage: 0.3, MonthlyCost: 1000}}, nil
}

func TestListSuggestions(t *testing.T) {
	svc := rcapp.NewService(fakeReader{})
	rows, err := svc.ListSuggestions(context.Background())
	if err != nil {
		t.Fatalf("ListSuggestions err=%v", err)
	}
	if len(rows) != 1 || rows[0].SubjectID != "A2" {
		t.Fatalf("unexpected suggestions: %+v", rows)
	}
}
