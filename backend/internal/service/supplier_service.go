package service

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

var (
	cnMobileRegex  = regexp.MustCompile(`^1\d{10}$`)
	landlineRegex  = regexp.MustCompile(`^0\d{2,3}-?\d{7,8}$`)
	taxNumberRegex = regexp.MustCompile(`^[0-9A-Z]{15,20}$`)
)

var AllowedBusinessScopes = []string{"服务器", "网络", "机柜服务", "维保", "机电设备", "软件", "零星工程", "总包"}

var allowedBusinessScopeSet = map[string]struct{}{
	"服务器": {}, "网络": {}, "机柜服务": {}, "维保": {}, "机电设备": {}, "软件": {}, "零星工程": {}, "总包": {},
}

type UpsertSupplierInput struct {
	CompanyFullName   string
	TaxNumber         string
	ProjectOwner      string
	ProjectOwnerPhone string
	TechContact       string
	TechContactPhone  string
	BusinessScope     string
}

type SupplierImportFailure struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type SupplierImportResult struct {
	Created  int                     `json:"created"`
	Updated  int                     `json:"updated"`
	Failed   int                     `json:"failed"`
	Failures []SupplierImportFailure `json:"failures,omitempty"`
}

type SupplierService struct {
	repo         repository.SupplierRepo
	contractRepo repository.ContractRepo
}

func NewSupplierService(repo repository.SupplierRepo, contractRepo repository.ContractRepo) *SupplierService {
	return &SupplierService{repo: repo, contractRepo: contractRepo}
}

func (s *SupplierService) CreateSupplier(ctx context.Context, in UpsertSupplierInput) (domain.Supplier, error) {
	item, err := validateAndBuildSupplier(domain.Supplier{
		SupplierID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		CompanyFullName:   strings.TrimSpace(in.CompanyFullName),
		TaxNumber:         normalizeTaxNumber(in.TaxNumber),
		ProjectOwner:      strings.TrimSpace(in.ProjectOwner),
		ProjectOwnerPhone: normalizePhone(in.ProjectOwnerPhone),
		TechContact:       strings.TrimSpace(in.TechContact),
		TechContactPhone:  normalizePhone(in.TechContactPhone),
		BusinessScope:     normalizeBusinessScope(in.BusinessScope),
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
	old.TaxNumber = normalizeTaxNumber(in.TaxNumber)
	old.ProjectOwner = strings.TrimSpace(in.ProjectOwner)
	old.ProjectOwnerPhone = normalizePhone(in.ProjectOwnerPhone)
	old.TechContact = strings.TrimSpace(in.TechContact)
	old.TechContactPhone = normalizePhone(in.TechContactPhone)
	old.BusinessScope = normalizeBusinessScope(in.BusinessScope)

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

func (s *SupplierService) ImportSuppliers(ctx context.Context, rows []UpsertSupplierInput) (SupplierImportResult, error) {
	result := SupplierImportResult{}
	existing, err := s.repo.ListSuppliers(ctx)
	if err != nil {
		return result, err
	}
	byTax := make(map[string]domain.Supplier, len(existing))
	for _, item := range existing {
		byTax[normalizeTaxNumber(item.TaxNumber)] = item
	}

	for i, row := range rows {
		rowNo := i + 2
		tax := normalizeTaxNumber(row.TaxNumber)
		if strings.TrimSpace(tax) == "" && strings.TrimSpace(row.CompanyFullName) == "" && strings.TrimSpace(row.BusinessScope) == "" {
			continue
		}
		if tax == "" {
			result.Failed++
			result.Failures = append(result.Failures, SupplierImportFailure{Row: rowNo, Reason: "tax_number is required"})
			continue
		}
		if old, ok := byTax[tax]; ok {
			if _, err := s.UpdateSupplier(ctx, old.SupplierID, row); err != nil {
				result.Failed++
				result.Failures = append(result.Failures, SupplierImportFailure{Row: rowNo, Reason: err.Error()})
				continue
			}
			result.Updated++
		} else {
			if _, err := s.CreateSupplier(ctx, row); err != nil {
				result.Failed++
				result.Failures = append(result.Failures, SupplierImportFailure{Row: rowNo, Reason: err.Error()})
				continue
			}
			result.Created++
		}
	}
	return result, nil
}

func (s *SupplierService) DeleteSupplier(ctx context.Context, supplierID string) error {
	supplierID = strings.TrimSpace(supplierID)
	item, err := s.repo.GetSupplier(ctx, supplierID)
	if err != nil {
		return err
	}
	if s.contractRepo != nil {
		contracts, err := s.contractRepo.ListContracts(ctx)
		if err != nil {
			return err
		}
		for _, c := range contracts {
			if strings.TrimSpace(c.Supplier) == strings.TrimSpace(item.CompanyFullName) || strings.TrimSpace(c.Supplier) == supplierID {
				return fmt.Errorf("supplier is referenced by contract %s (%s), cannot delete", c.ContractID, c.ContractName)
			}
		}
	}
	return s.repo.DeleteSupplier(ctx, supplierID)
}

func validateAndBuildSupplier(item domain.Supplier, isCreate bool) (domain.Supplier, error) {
	if strings.TrimSpace(item.CompanyFullName) == "" {
		return domain.Supplier{}, fmt.Errorf("company_full_name is required")
	}
	if strings.TrimSpace(item.TaxNumber) == "" {
		return domain.Supplier{}, fmt.Errorf("tax_number is required")
	}
	if !taxNumberRegex.MatchString(strings.ToUpper(strings.TrimSpace(item.TaxNumber))) {
		return domain.Supplier{}, fmt.Errorf("invalid tax_number format")
	}
	if strings.TrimSpace(item.ProjectOwnerPhone) != "" && !isValidPhone(item.ProjectOwnerPhone) {
		return domain.Supplier{}, fmt.Errorf("invalid project_owner_phone format")
	}
	if strings.TrimSpace(item.TechContactPhone) != "" && !isValidPhone(item.TechContactPhone) {
		return domain.Supplier{}, fmt.Errorf("invalid tech_contact_phone format")
	}
	if strings.TrimSpace(item.BusinessScope) == "" {
		return domain.Supplier{}, fmt.Errorf("business_scope is required")
	}
	if err := validateBusinessScope(item.BusinessScope); err != nil {
		return domain.Supplier{}, err
	}
	now := time.Now().Format(time.RFC3339)
	if isCreate || strings.TrimSpace(item.CreatedAt) == "" {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	return item, nil
}

func normalizeTaxNumber(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func normalizePhone(s string) string {
	return strings.TrimSpace(s)
}

func isValidPhone(s string) bool {
	v := strings.ReplaceAll(strings.TrimSpace(s), " ", "")
	return cnMobileRegex.MatchString(v) || landlineRegex.MatchString(v)
}

func validateBusinessScope(raw string) error {
	scopes := splitBusinessScopes(raw)
	if len(scopes) == 0 {
		return fmt.Errorf("business_scope is required")
	}
	for _, scope := range scopes {
		if _, ok := allowedBusinessScopeSet[scope]; !ok {
			return fmt.Errorf("invalid business_scope option: %s", scope)
		}
	}
	return nil
}

func normalizeBusinessScope(raw string) string {
	scopes := splitBusinessScopes(raw)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return strings.Join(out, "、")
}

func splitBusinessScopes(raw string) []string {
	replacer := strings.NewReplacer("，", ",", "、", ",", ";", ",", "；", ",", "|", ",", "/", ",")
	normalized := replacer.Replace(strings.TrimSpace(raw))
	parts := strings.Split(normalized, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}
