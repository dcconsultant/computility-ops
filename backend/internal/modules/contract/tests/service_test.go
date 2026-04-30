package tests

import (
	"context"
	"testing"

	legacy "computility-ops/backend/internal/domain"
	contractapp "computility-ops/backend/internal/modules/contract/application"
)

type fakeRepo struct{}

func (fakeRepo) ListContracts(ctx context.Context) ([]legacy.Contract, error) {
	_ = ctx
	return []legacy.Contract{{ContractID: "C1", ContractName: "合同1", Supplier: "供应商A"}}, nil
}

func (fakeRepo) GetContract(ctx context.Context, contractID string) (legacy.Contract, error) {
	_ = ctx
	return legacy.Contract{ContractID: contractID, ContractName: "合同详情", Supplier: "供应商B"}, nil
}

func TestListContracts(t *testing.T) {
	svc := contractapp.NewService(fakeRepo{})
	rows, err := svc.ListContracts(context.Background())
	if err != nil {
		t.Fatalf("ListContracts err=%v", err)
	}
	if len(rows) != 1 || rows[0].ContractID != "C1" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestGetContract(t *testing.T) {
	svc := contractapp.NewService(fakeRepo{})
	row, err := svc.GetContract(context.Background(), " C2 ")
	if err != nil {
		t.Fatalf("GetContract err=%v", err)
	}
	if row.ContractID != "C2" {
		t.Fatalf("expected trimmed ID C2, got %s", row.ContractID)
	}
}
