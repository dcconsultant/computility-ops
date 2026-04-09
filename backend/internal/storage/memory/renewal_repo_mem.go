package memory

import (
	"context"
	"fmt"
	"sync"

	"computility-ops/backend/internal/domain"
)

type RenewalRepo struct {
	mu    sync.RWMutex
	plans map[string]domain.RenewalPlan
}

func NewRenewalRepo() *RenewalRepo {
	return &RenewalRepo{plans: map[string]domain.RenewalPlan{}}
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
