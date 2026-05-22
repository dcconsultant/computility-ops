package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"computility-ops/backend/internal/domain"
)

type SupplierRepo struct {
	mu        sync.RWMutex
	suppliers map[string]domain.Supplier
}

func NewSupplierRepo() *SupplierRepo {
	return &SupplierRepo{suppliers: map[string]domain.Supplier{}}
}

func (r *SupplierRepo) SaveSupplier(_ context.Context, supplier domain.Supplier) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.suppliers[supplier.SupplierID] = supplier
	return nil
}

func (r *SupplierRepo) GetSupplier(_ context.Context, supplierID string) (domain.Supplier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.suppliers[supplierID]
	if !ok {
		return domain.Supplier{}, fmt.Errorf("supplier %s not found", supplierID)
	}
	return v, nil
}

func (r *SupplierRepo) ListSuppliers(_ context.Context) ([]domain.Supplier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Supplier, 0, len(r.suppliers))
	for _, v := range r.suppliers {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *SupplierRepo) DeleteSupplier(_ context.Context, supplierID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.suppliers[supplierID]; !ok {
		return fmt.Errorf("supplier %s not found", supplierID)
	}
	delete(r.suppliers, supplierID)
	return nil
}
