package memory

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"computility-ops/backend/internal/domain"
)

type RenewalRepo struct {
	mu         sync.RWMutex
	plans      map[string]domain.RenewalPlan
	unitPrices []domain.RenewalUnitPrice
	settings   *domain.RenewalPlanSettings
}

func NewRenewalRepo() *RenewalRepo {
	return &RenewalRepo{plans: map[string]domain.RenewalPlan{}, unitPrices: []domain.RenewalUnitPrice{}}
}

func (r *RenewalRepo) SavePlan(_ context.Context, plan domain.RenewalPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plans[plan.PlanID] = plan
	return nil
}

func (r *RenewalRepo) GetPlan(_ context.Context, planID string) (domain.RenewalPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plans[planID]
	if !ok {
		return domain.RenewalPlan{}, fmt.Errorf("plan %s not found", planID)
	}
	return p, nil
}

func (r *RenewalRepo) ListPlans(_ context.Context) ([]domain.RenewalPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.RenewalPlan, 0, len(r.plans))
	for _, p := range r.plans {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		ii, _ := strconv.ParseInt(out[i].PlanID, 10, 64)
		jj, _ := strconv.ParseInt(out[j].PlanID, 10, 64)
		return ii > jj
	})
	return out, nil
}

func (r *RenewalRepo) DeletePlan(_ context.Context, planID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.plans[planID]; !ok {
		return fmt.Errorf("plan %s not found", planID)
	}
	delete(r.plans, planID)
	return nil
}

func (r *RenewalRepo) ListUnitPrices(_ context.Context) ([]domain.RenewalUnitPrice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.RenewalUnitPrice(nil), r.unitPrices...), nil
}

func (r *RenewalRepo) ReplaceUnitPrices(_ context.Context, prices []domain.RenewalUnitPrice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unitPrices = append([]domain.RenewalUnitPrice(nil), prices...)
	return nil
}

func (r *RenewalRepo) GetSettings(_ context.Context) (domain.RenewalPlanSettings, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.settings == nil {
		return domain.RenewalPlanSettings{}, false, nil
	}
	return *r.settings, true, nil
}

func (r *RenewalRepo) SaveSettings(_ context.Context, settings domain.RenewalPlanSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	x := settings
	r.settings = &x
	return nil
}
