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
	"github.com/xuri/excelize/v2"
)

type ArrivalPlanFilter struct {
	Keyword              string
	Category             string
	Supplier             string
	OrderNo              string
	MaterialCode         string
	EstimatedArrivalFrom string
	EstimatedArrivalTo   string
}

type DeviceArrivalFilter struct {
	Keyword           string
	Category          string
	Manufacturer      string
	PackageCode       string
	PurchaseRequestNo string
	PONo              string
	ArrivalFrom       string
	ArrivalTo         string
}

type AccessoryArrivalFilter struct {
	Keyword           string
	Supplier          string
	IDCRoom           string
	PurchaseRequestNo string
	PONo              string
	ArrivalFrom       string
	ArrivalTo         string
}

type DeliveryService struct {
	repo repository.DeliveryRepo
}

func NewDeliveryService(repo repository.DeliveryRepo) *DeliveryService {
	return &DeliveryService{repo: repo}
}

func (s *DeliveryService) CreateArrivalPlan(ctx context.Context, in domain.ArrivalPlan) (domain.ArrivalPlan, error) {
	in.PlanID = strconv.FormatInt(time.Now().UnixNano(), 10)
	item, err := validateArrivalPlan(in, true)
	if err != nil {
		return domain.ArrivalPlan{}, err
	}
	return item, s.repo.SaveArrivalPlan(ctx, item)
}

func (s *DeliveryService) UpdateArrivalPlan(ctx context.Context, planID string, in domain.ArrivalPlan) (domain.ArrivalPlan, error) {
	old, err := s.repo.GetArrivalPlan(ctx, strings.TrimSpace(planID))
	if err != nil {
		return domain.ArrivalPlan{}, err
	}
	in.PlanID = old.PlanID
	in.CreatedAt = old.CreatedAt
	item, err := validateArrivalPlan(in, false)
	if err != nil {
		return domain.ArrivalPlan{}, err
	}
	return item, s.repo.SaveArrivalPlan(ctx, item)
}

func (s *DeliveryService) ListArrivalPlans(ctx context.Context, filter ArrivalPlanFilter) ([]domain.ArrivalPlan, error) {
	list, err := s.repo.ListArrivalPlans(ctx)
	if err != nil {
		return nil, err
	}
	from, _ := parseOptionalDeliveryTime(filter.EstimatedArrivalFrom)
	to, _ := parseOptionalDeliveryTime(filter.EstimatedArrivalTo)
	out := make([]domain.ArrivalPlan, 0, len(list))
	for _, item := range list {
		if !matchesText(item.Category, filter.Category) || !matchesText(item.Supplier, filter.Supplier) || !matchesText(item.OrderNo, filter.OrderNo) || !matchesText(item.MaterialCode, filter.MaterialCode) {
			continue
		}
		if !matchesKeyword(filter.Keyword, item.Category, item.MaterialCode, item.MaterialName, item.Supplier, item.OrderNo, item.ReceivingAddress, item.AssetCodeRange) {
			continue
		}
		if !matchesTimeRange(item.EstimatedArrivalTime, from, to) {
			continue
		}
		out = append(out, item)
	}
	sortByUpdatedAt(out, func(i int) string { return out[i].UpdatedAt })
	return out, nil
}

func (s *DeliveryService) DeleteArrivalPlan(ctx context.Context, planID string) error {
	return s.repo.DeleteArrivalPlan(ctx, strings.TrimSpace(planID))
}

func (s *DeliveryService) CreateDeviceArrival(ctx context.Context, in domain.DeviceArrivalRecord) (domain.DeviceArrivalRecord, error) {
	in.RecordID = strconv.FormatInt(time.Now().UnixNano(), 10)
	item, err := validateDeviceArrival(in, true)
	if err != nil {
		return domain.DeviceArrivalRecord{}, err
	}
	return item, s.repo.SaveDeviceArrival(ctx, item)
}

func (s *DeliveryService) UpdateDeviceArrival(ctx context.Context, recordID string, in domain.DeviceArrivalRecord) (domain.DeviceArrivalRecord, error) {
	old, err := s.repo.GetDeviceArrival(ctx, strings.TrimSpace(recordID))
	if err != nil {
		return domain.DeviceArrivalRecord{}, err
	}
	in.RecordID = old.RecordID
	in.CreatedAt = old.CreatedAt
	item, err := validateDeviceArrival(in, false)
	if err != nil {
		return domain.DeviceArrivalRecord{}, err
	}
	return item, s.repo.SaveDeviceArrival(ctx, item)
}

func (s *DeliveryService) ListDeviceArrivals(ctx context.Context, filter DeviceArrivalFilter) ([]domain.DeviceArrivalRecord, error) {
	list, err := s.repo.ListDeviceArrivals(ctx)
	if err != nil {
		return nil, err
	}
	from, _ := parseOptionalDeliveryTime(filter.ArrivalFrom)
	to, _ := parseOptionalDeliveryTime(filter.ArrivalTo)
	out := make([]domain.DeviceArrivalRecord, 0, len(list))
	for _, item := range list {
		if !matchesText(item.Category, filter.Category) || !matchesText(item.Manufacturer, filter.Manufacturer) || !matchesText(item.PackageCode, filter.PackageCode) || !matchesText(item.PurchaseRequestNo, filter.PurchaseRequestNo) || !matchesText(item.PONo, filter.PONo) {
			continue
		}
		if !matchesKeyword(filter.Keyword, item.Category, item.PackageCode, item.PackageType, item.MaterialServiceCode, item.MaterialServiceDescription, item.Manufacturer, item.ReceivingLocation, item.PurchaseRequestNo, item.PONo) {
			continue
		}
		if !matchesTimeRange(item.ActualArrivalTime, from, to) {
			continue
		}
		out = append(out, item)
	}
	sortByUpdatedAt(out, func(i int) string { return out[i].UpdatedAt })
	return out, nil
}

func (s *DeliveryService) DeleteDeviceArrival(ctx context.Context, recordID string) error {
	return s.repo.DeleteDeviceArrival(ctx, strings.TrimSpace(recordID))
}

func (s *DeliveryService) CreateAccessoryArrival(ctx context.Context, in domain.AccessoryArrivalRecord) (domain.AccessoryArrivalRecord, error) {
	in.RecordID = strconv.FormatInt(time.Now().UnixNano(), 10)
	item, err := validateAccessoryArrival(in, true)
	if err != nil {
		return domain.AccessoryArrivalRecord{}, err
	}
	return item, s.repo.SaveAccessoryArrival(ctx, item)
}

func (s *DeliveryService) UpdateAccessoryArrival(ctx context.Context, recordID string, in domain.AccessoryArrivalRecord) (domain.AccessoryArrivalRecord, error) {
	old, err := s.repo.GetAccessoryArrival(ctx, strings.TrimSpace(recordID))
	if err != nil {
		return domain.AccessoryArrivalRecord{}, err
	}
	in.RecordID = old.RecordID
	in.CreatedAt = old.CreatedAt
	item, err := validateAccessoryArrival(in, false)
	if err != nil {
		return domain.AccessoryArrivalRecord{}, err
	}
	return item, s.repo.SaveAccessoryArrival(ctx, item)
}

func (s *DeliveryService) ListAccessoryArrivals(ctx context.Context, filter AccessoryArrivalFilter) ([]domain.AccessoryArrivalRecord, error) {
	list, err := s.repo.ListAccessoryArrivals(ctx)
	if err != nil {
		return nil, err
	}
	from, _ := parseOptionalDeliveryTime(filter.ArrivalFrom)
	to, _ := parseOptionalDeliveryTime(filter.ArrivalTo)
	out := make([]domain.AccessoryArrivalRecord, 0, len(list))
	for _, item := range list {
		if !matchesText(item.Supplier, filter.Supplier) || !matchesText(item.IDCRoom, filter.IDCRoom) || !matchesText(item.PurchaseRequestNo, filter.PurchaseRequestNo) || !matchesText(item.PONo, filter.PONo) {
			continue
		}
		if !matchesKeyword(filter.Keyword, item.PurchaseRequestNo, item.MaterialServiceCode, item.MaterialServiceDescription, item.Supplier, item.IDCRoom, item.PurchaseBackground, item.PONo) {
			continue
		}
		if !matchesTimeRange(item.ArrivalTime, from, to) {
			continue
		}
		out = append(out, item)
	}
	sortByUpdatedAt(out, func(i int) string { return out[i].UpdatedAt })
	return out, nil
}

func (s *DeliveryService) DeleteAccessoryArrival(ctx context.Context, recordID string) error {
	return s.repo.DeleteAccessoryArrival(ctx, strings.TrimSpace(recordID))
}

func validateArrivalPlan(item domain.ArrivalPlan, isCreate bool) (domain.ArrivalPlan, error) {
	item.Category = strings.TrimSpace(item.Category)
	if !isAllowed(item.Category, "服务器", "网络设备", "耗材及配件") {
		return domain.ArrivalPlan{}, fmt.Errorf("invalid category")
	}
	if item.Quantity <= 0 {
		return domain.ArrivalPlan{}, fmt.Errorf("quantity must be > 0")
	}
	item.MaterialCode = strings.TrimSpace(item.MaterialCode)
	item.MaterialName = strings.TrimSpace(item.MaterialName)
	item.ReceivingAddress = strings.TrimSpace(item.ReceivingAddress)
	item.Supplier = strings.TrimSpace(item.Supplier)
	item.OrderNo = strings.TrimSpace(item.OrderNo)
	item.AssetCodeRange = strings.TrimSpace(item.AssetCodeRange)
	item.Remark = strings.TrimSpace(item.Remark)
	if item.MaterialCode == "" || item.MaterialName == "" || item.ReceivingAddress == "" || item.Supplier == "" || item.OrderNo == "" {
		return domain.ArrivalPlan{}, fmt.Errorf("material_code, material_name, receiving_address, supplier and order_no are required")
	}
	normalized, err := normalizeDeliveryTime(item.EstimatedArrivalTime, "estimated_arrival_time")
	if err != nil {
		return domain.ArrivalPlan{}, err
	}
	item.EstimatedArrivalTime = normalized
	stampDeliveryAudit(&item.CreatedAt, &item.UpdatedAt, isCreate)
	return item, nil
}

func validateDeviceArrival(item domain.DeviceArrivalRecord, isCreate bool) (domain.DeviceArrivalRecord, error) {
	item.Category = strings.TrimSpace(item.Category)
	if !isAllowed(item.Category, "服务器", "网络设备") {
		return domain.DeviceArrivalRecord{}, fmt.Errorf("invalid category")
	}
	if item.Quantity <= 0 {
		return domain.DeviceArrivalRecord{}, fmt.Errorf("quantity must be > 0")
	}
	item.PackageCode = strings.TrimSpace(item.PackageCode)
	item.PackageType = strings.TrimSpace(item.PackageType)
	item.MaterialServiceCode = strings.TrimSpace(item.MaterialServiceCode)
	item.MaterialServiceDescription = strings.TrimSpace(item.MaterialServiceDescription)
	item.Manufacturer = strings.TrimSpace(item.Manufacturer)
	item.ReceivingLocation = strings.TrimSpace(item.ReceivingLocation)
	item.PurchaseRequestNo = strings.TrimSpace(item.PurchaseRequestNo)
	item.PONo = strings.TrimSpace(item.PONo)
	if item.PackageCode == "" || item.MaterialServiceCode == "" || item.Manufacturer == "" || item.ReceivingLocation == "" || item.PurchaseRequestNo == "" || item.PONo == "" {
		return domain.DeviceArrivalRecord{}, fmt.Errorf("package_code, material_service_code, manufacturer, receiving_location, purchase_request_no and po_no are required")
	}
	srmAt, err := normalizeDeliveryTime(item.SRMRequirementSubmittedAt, "srm_requirement_submitted_at")
	if err != nil {
		return domain.DeviceArrivalRecord{}, err
	}
	arrivalAt, err := normalizeDeliveryTime(item.ActualArrivalTime, "actual_arrival_time")
	if err != nil {
		return domain.DeviceArrivalRecord{}, err
	}
	if isTimeBefore(arrivalAt, srmAt) {
		return domain.DeviceArrivalRecord{}, fmt.Errorf("actual_arrival_time must be >= srm_requirement_submitted_at")
	}
	item.SRMRequirementSubmittedAt = srmAt
	item.ActualArrivalTime = arrivalAt
	stampDeliveryAudit(&item.CreatedAt, &item.UpdatedAt, isCreate)
	return item, nil
}

func validateAccessoryArrival(item domain.AccessoryArrivalRecord, isCreate bool) (domain.AccessoryArrivalRecord, error) {
	if item.Quantity <= 0 {
		return domain.AccessoryArrivalRecord{}, fmt.Errorf("quantity must be > 0")
	}
	item.PurchaseRequestNo = strings.TrimSpace(item.PurchaseRequestNo)
	item.MaterialServiceCode = strings.TrimSpace(item.MaterialServiceCode)
	item.MaterialServiceDescription = strings.TrimSpace(item.MaterialServiceDescription)
	item.Supplier = strings.TrimSpace(item.Supplier)
	item.IDCRoom = strings.TrimSpace(item.IDCRoom)
	item.PurchaseBackground = strings.TrimSpace(item.PurchaseBackground)
	item.PONo = strings.TrimSpace(item.PONo)
	if item.PurchaseRequestNo == "" || item.MaterialServiceCode == "" || item.Supplier == "" || item.IDCRoom == "" || item.PONo == "" {
		return domain.AccessoryArrivalRecord{}, fmt.Errorf("purchase_request_no, material_service_code, supplier, idc_room and po_no are required")
	}
	srmAt, err := normalizeDeliveryTime(item.SRMRequirementSubmittedAt, "srm_requirement_submitted_at")
	if err != nil {
		return domain.AccessoryArrivalRecord{}, err
	}
	arrivalAt, err := normalizeDeliveryTime(item.ArrivalTime, "arrival_time")
	if err != nil {
		return domain.AccessoryArrivalRecord{}, err
	}
	if isTimeBefore(arrivalAt, srmAt) {
		return domain.AccessoryArrivalRecord{}, fmt.Errorf("arrival_time must be >= srm_requirement_submitted_at")
	}
	item.SRMRequirementSubmittedAt = srmAt
	item.ArrivalTime = arrivalAt
	stampDeliveryAudit(&item.CreatedAt, &item.UpdatedAt, isCreate)
	return item, nil
}

func stampDeliveryAudit(createdAt, updatedAt *string, isCreate bool) {
	now := time.Now().Format(time.RFC3339)
	if isCreate || strings.TrimSpace(*createdAt) == "" {
		*createdAt = now
	}
	*updatedAt = now
}

func normalizeDeliveryTime(raw, field string) (string, error) {
	t, err := parseOptionalDeliveryTime(raw)
	if err != nil || t == nil {
		if err != nil {
			return "", fmt.Errorf("invalid %s: %v", field, err)
		}
		return "", fmt.Errorf("invalid %s", field)
	}
	return t.Format("2006-01-02 15:04:05"), nil
}

func parseOptionalDeliveryTime(raw string) (*time.Time, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	layouts := []string{
		"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02",
		"2006-1-2 15:04:05", "2006-1-2 15:04", "2006-1-2",
		"2006/01/02 15:04:05", "2006/01/02 15:04", "2006/01/02",
		"2006/1/2 15:04:05", "2006/1/2 15:04", "2006/1/2",
		"2006.01.02", "2006.1.2", "2006年01月02日", "2006年1月2日",
		"20060102", time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return &t, nil
		}
	}
	if serial, err := strconv.ParseFloat(v, 64); err == nil && serial > 0 {
		if t, err := excelize.ExcelDateToTime(serial, false); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("unsupported time format: %s", raw)
}

func isTimeBefore(a, b string) bool {
	at, _ := parseOptionalDeliveryTime(a)
	bt, _ := parseOptionalDeliveryTime(b)
	return at != nil && bt != nil && at.Before(*bt)
}

func matchesTimeRange(value string, from, to *time.Time) bool {
	t, err := parseOptionalDeliveryTime(value)
	if err != nil || t == nil {
		return from == nil && to == nil
	}
	if from != nil && t.Before(*from) {
		return false
	}
	if to != nil && t.After(*to) {
		return false
	}
	return true
}

func matchesText(value, filter string) bool {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(value)), strings.ToLower(filter))
}

func matchesKeyword(keyword string, values ...string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	for _, v := range values {
		if strings.Contains(strings.ToLower(strings.TrimSpace(v)), keyword) {
			return true
		}
	}
	return false
}

func isAllowed(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}

func sortByUpdatedAt[T any](list []T, updatedAt func(int) string) {
	sort.Slice(list, func(i, j int) bool {
		return strings.TrimSpace(updatedAt(i)) > strings.TrimSpace(updatedAt(j))
	})
}
