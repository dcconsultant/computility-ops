package application

import (
	"context"
	"strings"

	legacy "computility-ops/backend/internal/domain"
	contractdomain "computility-ops/backend/internal/modules/contract/domain"
)

type Repository interface {
	ListContracts(ctx context.Context) ([]legacy.Contract, error)
	GetContract(ctx context.Context, contractID string) (legacy.Contract, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListContracts(ctx context.Context) ([]contractdomain.Contract, error) {
	rows, err := s.repo.ListContracts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]contractdomain.Contract, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapLegacyContract(row))
	}
	return out, nil
}

func (s *Service) GetContract(ctx context.Context, contractID string) (contractdomain.Contract, error) {
	row, err := s.repo.GetContract(ctx, strings.TrimSpace(contractID))
	if err != nil {
		return contractdomain.Contract{}, err
	}
	return mapLegacyContract(row), nil
}

func mapLegacyContract(row legacy.Contract) contractdomain.Contract {
	attachments := make([]contractdomain.ContractAttachment, 0, len(row.Attachments))
	for _, att := range row.Attachments {
		attachments = append(attachments, contractdomain.ContractAttachment{
			AttachmentID: att.AttachmentID,
			FileName:     att.FileName,
			FileSize:     att.FileSize,
			MimeType:     att.MimeType,
			UploadedAt:   att.UploadedAt,
		})
	}
	return contractdomain.Contract{
		ContractID:      row.ContractID,
		ContractName:    row.ContractName,
		PeriodStart:     row.PeriodStart,
		PeriodEnd:       row.PeriodEnd,
		PreTaxAmount:    row.PreTaxAmount,
		Supplier:        row.Supplier,
		BusinessContact: row.BusinessContact,
		TechContact:     row.TechContact,
		Attachments:     attachments,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
