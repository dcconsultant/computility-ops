package tests

import (
	"context"
	"testing"

	srapp "computility-ops/backend/internal/modules/self-repair/application"
	srdomain "computility-ops/backend/internal/modules/self-repair/domain"
)

type fakeReader struct{}

func (fakeReader) ListCases(ctx context.Context) ([]srdomain.SelfRepairCase, error) {
	_ = ctx
	return []srdomain.SelfRepairCase{{CaseID: "F1", AssetID: "A3", SparePartAvailable: true, EngineerSkillMatched: true}}, nil
}

func TestListSuggestions(t *testing.T) {
	svc := srapp.NewService(fakeReader{})
	rows, err := svc.ListSuggestions(context.Background())
	if err != nil {
		t.Fatalf("ListSuggestions err=%v", err)
	}
	if len(rows) != 1 || rows[0].SubjectID != "F1" {
		t.Fatalf("unexpected suggestions: %+v", rows)
	}
}
