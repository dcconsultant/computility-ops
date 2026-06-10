package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"computility-ops/backend/internal/domain"
)

type DeliveryRepo struct {
	mu                sync.RWMutex
	arrivalPlans      map[string]domain.ArrivalPlan
	deviceArrivals    map[string]domain.DeviceArrivalRecord
	accessoryArrivals map[string]domain.AccessoryArrivalRecord
}

func NewDeliveryRepo() *DeliveryRepo {
	return &DeliveryRepo{
		arrivalPlans:      map[string]domain.ArrivalPlan{},
		deviceArrivals:    map[string]domain.DeviceArrivalRecord{},
		accessoryArrivals: map[string]domain.AccessoryArrivalRecord{},
	}
}

func (r *DeliveryRepo) SaveArrivalPlan(_ context.Context, item domain.ArrivalPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.arrivalPlans[item.PlanID] = item
	return nil
}

func (r *DeliveryRepo) GetArrivalPlan(_ context.Context, planID string) (domain.ArrivalPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.arrivalPlans[planID]
	if !ok {
		return domain.ArrivalPlan{}, fmt.Errorf("arrival plan %s not found", planID)
	}
	return item, nil
}

func (r *DeliveryRepo) ListArrivalPlans(_ context.Context) ([]domain.ArrivalPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.ArrivalPlan, 0, len(r.arrivalPlans))
	for _, item := range r.arrivalPlans {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *DeliveryRepo) DeleteArrivalPlan(_ context.Context, planID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.arrivalPlans[planID]; !ok {
		return fmt.Errorf("arrival plan %s not found", planID)
	}
	delete(r.arrivalPlans, planID)
	return nil
}

func (r *DeliveryRepo) SaveDeviceArrival(_ context.Context, item domain.DeviceArrivalRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deviceArrivals[item.RecordID] = item
	return nil
}

func (r *DeliveryRepo) GetDeviceArrival(_ context.Context, recordID string) (domain.DeviceArrivalRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.deviceArrivals[recordID]
	if !ok {
		return domain.DeviceArrivalRecord{}, fmt.Errorf("device arrival %s not found", recordID)
	}
	return item, nil
}

func (r *DeliveryRepo) ListDeviceArrivals(_ context.Context) ([]domain.DeviceArrivalRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.DeviceArrivalRecord, 0, len(r.deviceArrivals))
	for _, item := range r.deviceArrivals {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *DeliveryRepo) DeleteDeviceArrival(_ context.Context, recordID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.deviceArrivals[recordID]; !ok {
		return fmt.Errorf("device arrival %s not found", recordID)
	}
	delete(r.deviceArrivals, recordID)
	return nil
}

func (r *DeliveryRepo) SaveAccessoryArrival(_ context.Context, item domain.AccessoryArrivalRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accessoryArrivals[item.RecordID] = item
	return nil
}

func (r *DeliveryRepo) GetAccessoryArrival(_ context.Context, recordID string) (domain.AccessoryArrivalRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.accessoryArrivals[recordID]
	if !ok {
		return domain.AccessoryArrivalRecord{}, fmt.Errorf("accessory arrival %s not found", recordID)
	}
	return item, nil
}

func (r *DeliveryRepo) ListAccessoryArrivals(_ context.Context) ([]domain.AccessoryArrivalRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.AccessoryArrivalRecord, 0, len(r.accessoryArrivals))
	for _, item := range r.accessoryArrivals {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *DeliveryRepo) DeleteAccessoryArrival(_ context.Context, recordID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.accessoryArrivals[recordID]; !ok {
		return fmt.Errorf("accessory arrival %s not found", recordID)
	}
	delete(r.accessoryArrivals, recordID)
	return nil
}
