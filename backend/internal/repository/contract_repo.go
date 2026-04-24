package repository

import (
	"context"

	"computility-ops/backend/internal/domain"
)

type ContractRepo interface {
	SaveContract(ctx context.Context, contract domain.Contract) error
	GetContract(ctx context.Context, contractID string) (domain.Contract, error)
	ListContracts(ctx context.Context) ([]domain.Contract, error)
	DeleteContract(ctx context.Context, contractID string) error

	SaveAttachment(ctx context.Context, attachment domain.ContractAttachmentBlob) error
	GetAttachment(ctx context.Context, contractID, attachmentID string) (domain.ContractAttachmentBlob, error)
	DeleteAttachment(ctx context.Context, contractID, attachmentID string) error
}
