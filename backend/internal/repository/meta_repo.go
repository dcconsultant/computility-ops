package repository

import (
	"context"

	"computility-ops/backend/internal/domain"
)

type MetaRepo interface {
	CreateModel(ctx context.Context, model domain.MetaModel) error
	GetModel(ctx context.Context, modelID string) (domain.MetaModel, error)
	GetModelByCode(ctx context.Context, modelCode string) (domain.MetaModel, error)
	ListModels(ctx context.Context, status string) ([]domain.MetaModel, error)
	UpdateModel(ctx context.Context, model domain.MetaModel) error
	DeleteModel(ctx context.Context, modelID string) error

	ListFields(ctx context.Context, modelID string) ([]domain.MetaField, error)
	CreateField(ctx context.Context, field domain.MetaField) error
	GetField(ctx context.Context, modelID, fieldID string) (domain.MetaField, error)
	GetFieldByCode(ctx context.Context, modelID, fieldCode string) (domain.MetaField, error)
	UpdateField(ctx context.Context, field domain.MetaField) error
	DeleteField(ctx context.Context, modelID, fieldID string) error
	ReorderFields(ctx context.Context, modelID string, order []FieldOrderItem) error

	ListReferences(ctx context.Context, modelID string) ([]domain.MetaReference, error)
	CreateReference(ctx context.Context, ref domain.MetaReference) error
	GetReference(ctx context.Context, modelID, refID string) (domain.MetaReference, error)
	UpdateReference(ctx context.Context, ref domain.MetaReference) error
	DeleteReference(ctx context.Context, modelID, refID string) error

	CountRecords(ctx context.Context, modelID string) (int64, error)

	CreateVersion(ctx context.Context, version domain.MetaModelVersion) error
	ListVersions(ctx context.Context, modelID string) ([]domain.MetaModelVersion, error)
	GetVersion(ctx context.Context, modelID string, versionNo int) (domain.MetaModelVersion, error)
}

type FieldOrderItem struct {
	FieldID string
	SortNo  int
}
