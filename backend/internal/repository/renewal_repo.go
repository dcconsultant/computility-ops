package repository

import (
	"context"

	"computility-ops/backend/internal/domain"
)

type RenewalPlanRepo interface {
	SavePlan(ctx context.Context, plan domain.RenewalPlan) error
	GetPlan(ctx context.Context, planID string) (domain.RenewalPlan, error)
}
