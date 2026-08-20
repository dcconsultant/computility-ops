package memory

import (
	"context"
	"sync"

	"computility-ops/backend/internal/domain"
)

type DeliveryDecisionRepo struct {
	mu     sync.RWMutex
	config *domain.DeliveryDecisionConfigState
}

func NewDeliveryDecisionRepo() *DeliveryDecisionRepo {
	return &DeliveryDecisionRepo{}
}

func (r *DeliveryDecisionRepo) GetConfig(_ context.Context) (domain.DeliveryDecisionConfigState, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.config == nil {
		return domain.DeliveryDecisionConfigState{}, false, nil
	}
	return *r.config, true, nil
}

func (r *DeliveryDecisionRepo) SaveConfig(_ context.Context, state domain.DeliveryDecisionConfigState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	x := state
	r.config = &x
	return nil
}
