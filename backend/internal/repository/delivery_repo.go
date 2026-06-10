package repository

import (
	"context"

	"computility-ops/backend/internal/domain"
)

type DeliveryRepo interface {
	SaveArrivalPlan(ctx context.Context, item domain.ArrivalPlan) error
	GetArrivalPlan(ctx context.Context, planID string) (domain.ArrivalPlan, error)
	ListArrivalPlans(ctx context.Context) ([]domain.ArrivalPlan, error)
	DeleteArrivalPlan(ctx context.Context, planID string) error

	SaveDeviceArrival(ctx context.Context, item domain.DeviceArrivalRecord) error
	GetDeviceArrival(ctx context.Context, recordID string) (domain.DeviceArrivalRecord, error)
	ListDeviceArrivals(ctx context.Context) ([]domain.DeviceArrivalRecord, error)
	DeleteDeviceArrival(ctx context.Context, recordID string) error

	SaveAccessoryArrival(ctx context.Context, item domain.AccessoryArrivalRecord) error
	GetAccessoryArrival(ctx context.Context, recordID string) (domain.AccessoryArrivalRecord, error)
	ListAccessoryArrivals(ctx context.Context) ([]domain.AccessoryArrivalRecord, error)
	DeleteAccessoryArrival(ctx context.Context, recordID string) error
}
