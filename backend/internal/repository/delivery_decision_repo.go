package repository

import (
	"context"

	"computility-ops/backend/internal/domain"
)

type DeliveryDecisionRepo interface {
	GetConfig(ctx context.Context) (domain.DeliveryDecisionConfigState, bool, error)
	SaveConfig(ctx context.Context, state domain.DeliveryDecisionConfigState) error
}
