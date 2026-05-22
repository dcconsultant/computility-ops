package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type UpsertSupplierInput struct {
	CompanyFullName   string
	TaxNumber         string
	ProjectOwner      string
	ProjectOwnerPhone string
	TechContact       string
	TechContactPhone  string
	BusinessScope     string
}

type SupplierService struct {
	repo repository.SupplierRepo
}

func NewSupplierService(repo repository.SupplierRepo) *SupplierService {
	return &SupplierService{repo: repo}
}

func (s *SupplierService) CreateSupplier(ctx context.Context, in UpsertSupplierInput) (domain.Supplier, error) {
	item, err := validateAndBuildSupplier(domain.Supplier{
		SupplierID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		CompanyFullName:   strings.TrimSpace(in.CompanyFullName),
		TaxNumber:         strings.TrimSpace(in.TaxNumber),
		ProjectOwner:      strings.TrimSpace(in.ProjectOwner),
		ProjectOwnerPhone: strings.TrimSpace(in.ProjectOwnerPhone),
		TechContact:       strings.TrimSpace(in.TechContact),
		TechContactPhone:  strings.TrimSpace(in.TechContactPhone),
		BusinessScope:     strings.TrimSpace(in.BusinessScope),
	}, true)
	if err != nil {
		return domain.Supplier{}, err
	}
	if err := s.repo.SaveSupplier(ctx, item); err != nil {
		return domain.Supplier{}, err
	}
	return item, nil
}

func (s *SupplierService) UpdateSupplier(ctx context.Context, supplierID string, in UpsertSupplierInput) (domain.Supplier, error) {
	old, err := s.repo.GetSupplier(ctx, strings.TrimSpace(supplierID))
	if err != nil {
		return domain.Supplier{}, err
	}
	old.CompanyFullName = strings.TrimSpace(in.CompanyFullName)
	old.TaxNumber = strings.TrimSpace(in.TaxNumber)
	old.ProjectOwner = strings.TrimSpace(in.ProjectOwner)
	old.ProjectOwnerPhone = strings.TrimSpace(in.ProjectOwnerPhone)
	old.TechContact = strings.TrimSpace(in.TechContact)
	old.TechContactPhone = strings.TrimSpace(in.TechContactPhone)
	old.BusinessScope = strings.TrimSpace(in.BusinessScope)

	item, err := validateAndBuildSupplier(old, false)
	if err != nil {
		return domain.Supplier{}, err
	}
	if err := s.repo.SaveSupplier(ctx, item); err != nil {
		return domain.Supplier{}, err
	}
	return item, nil
}

func (s *SupplierService) GetSupplier(ctx context.Context, supplierID string) (domain.Supplier, error) {
	return s.repo.GetSupplier(ctx, strings.TrimSpace(supplierID))
}

func (s *SupplierService) ListSuppliers(ctx context.Context, keyword string) ([]domain.Supplier, error) {
	list, err := s.repo.ListSuppliers(ctx)
	if err != nil {
		return nil, err
	}
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw != "" {
		filtered := make([]domain.Supplier, 0, len(list))
		for _, item := range list {
			bag := strings.ToLower(strings.Join([]string{item.CompanyFullName, item.TaxNumber, item.ProjectOwner, item.TechContact, item.BusinessScope}, " "))
			if strings.Contains(bag, kw) {
				filtered = append(filtered, item)
			}
		}
		list = filtered
	}
	sort.Slice(list, func(i, j int) bool {
		return strings.TrimSpace(list[i].UpdatedAt) > strings.TrimSpace(list[j].UpdatedAt)
	})
	return list, nil
}

func (s *SupplierService) DeleteSupplier(ctx context.Context, supplierID string) error {
	return s.repo.DeleteSupplier(ctx, strings.TrimSpace(supplierID))
}

func validateAndBuildSupplier(item domain.Supplier, isCreate bool) (domain.Supplier, error) {
	if strings.TrimSpace(item.CompanyFullName) == "" {
		return domain.Supplier{}, fmt.Errorf("company_full_name is required")
	}
	if strings.TrimSpace(item.TaxNumber) == "" {
		return domain.Supplier{}, fmt.Errorf("tax_number is required")
	}
	if strings.TrimSpace(item.BusinessScope) == "" {
		return domain.Supplier{}, fmt.Errorf("business_scope is required")
	}
	now := time.Now().Format(time.RFC3339)
	if isCreate || strings.TrimSpace(item.CreatedAt) == "" {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item, nil
}
