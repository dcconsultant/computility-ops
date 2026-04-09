package mysql

import (
	"context"
	"errors"

	"computility-ops/backend/internal/domain"
)

type RenewalRepo struct{}

func NewRenewalRepo(_ string) *RenewalRepo { return &RenewalRepo{} }

func (r *RenewalRepo) SavePlan(_ context.Context, _ domain.RenewalPlan) error {
	return errors.New("mysql repo not implemented in phase 1")
}

func (r *RenewalRepo) GetPlan(_ context.Context, _ string) (domain.RenewalPlan, error) {
	return domain.RenewalPlan{}, errors.New("mysql repo not implemented in phase 1")
}

func (r *RenewalRepo) ListPlans(_ context.Context) ([]domain.RenewalPlan, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}

func (r *RenewalRepo) DeletePlan(_ context.Context, _ string) error {
	return errors.New("mysql repo not implemented in phase 1")
}
