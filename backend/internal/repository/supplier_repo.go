package repository

import (
	"context"

	"computility-ops/backend/internal/domain"
)

type SupplierRepo interface {
	SaveSupplier(ctx context.Context, supplier domain.Supplier) error
	GetSupplier(ctx context.Context, supplierID string) (domain.Supplier, error)
	ListSuppliers(ctx context.Context) ([]domain.Supplier, error)
	DeleteSupplier(ctx context.Context, supplierID string) error
}
