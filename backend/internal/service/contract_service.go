package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type CreateContractInput struct {
	ContractName    string
	PeriodStart     string
	PeriodEnd       string
	PreTaxAmount    float64
	Supplier        string
	BusinessContact string
	TechContact     string
}

type UpdateContractInput struct {
	ContractName    string
	PeriodStart     string
	PeriodEnd       string
	PreTaxAmount    float64
	Supplier        string
	BusinessContact string
	TechContact     string
}

type ContractService struct {
	repo      repository.ContractRepo
	uploadDir string
}

func NewContractService(repo repository.ContractRepo) *ContractService {
	return &ContractService{
		repo:      repo,
		uploadDir: filepath.Join(os.TempDir(), "computility_ops", "contract_attachments"),
	}
}

func (s *ContractService) CreateContract(ctx context.Context, in CreateContractInput) (domain.Contract, error) {
	contract, err := validateAndBuildContract(domain.Contract{
		ContractID:      strconv.FormatInt(time.Now().UnixNano(), 10),
		ContractName:    strings.TrimSpace(in.ContractName),
		PeriodStart:     strings.TrimSpace(in.PeriodStart),
		PeriodEnd:       strings.TrimSpace(in.PeriodEnd),
		PreTaxAmount:    in.PreTaxAmount,
		Supplier:        strings.TrimSpace(in.Supplier),
		BusinessContact: strings.TrimSpace(in.BusinessContact),
		TechContact:     strings.TrimSpace(in.TechContact),
		Attachments:     []domain.ContractAttachment{},
	}, true)
	if err != nil {
		return domain.Contract{}, err
	}
	if err := s.repo.SaveContract(ctx, contract); err != nil {
		return domain.Contract{}, err
	}
	return contract, nil
}

func (s *ContractService) UpdateContract(ctx context.Context, contractID string, in UpdateContractInput) (domain.Contract, error) {
	old, err := s.repo.GetContract(ctx, strings.TrimSpace(contractID))
	if err != nil {
		return domain.Contract{}, err
	}
	old.ContractName = strings.TrimSpace(in.ContractName)
	old.PeriodStart = strings.TrimSpace(in.PeriodStart)
	old.PeriodEnd = strings.TrimSpace(in.PeriodEnd)
	old.PreTaxAmount = in.PreTaxAmount
	old.Supplier = strings.TrimSpace(in.Supplier)
	old.BusinessContact = strings.TrimSpace(in.BusinessContact)
	old.TechContact = strings.TrimSpace(in.TechContact)

	contract, err := validateAndBuildContract(old, false)
	if err != nil {
		return domain.Contract{}, err
	}
	if err := s.repo.SaveContract(ctx, contract); err != nil {
		return domain.Contract{}, err
	}
	return contract, nil
}

func (s *ContractService) ListContracts(ctx context.Context) ([]domain.Contract, error) {
	list, err := s.repo.ListContracts(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return strings.TrimSpace(list[i].UpdatedAt) > strings.TrimSpace(list[j].UpdatedAt)
	})
	return list, nil
}

func (s *ContractService) GetContract(ctx context.Context, contractID string) (domain.Contract, error) {
	return s.repo.GetContract(ctx, strings.TrimSpace(contractID))
}

func (s *ContractService) DeleteContract(ctx context.Context, contractID string) error {
	contractID = strings.TrimSpace(contractID)
	contract, err := s.repo.GetContract(ctx, contractID)
	if err != nil {
		return err
	}
	for _, att := range contract.Attachments {
		if err := s.deleteAttachmentFile(contractID, att.AttachmentID); err != nil {
			return err
		}
	}
	return s.repo.DeleteContract(ctx, contractID)
}

func (s *ContractService) SaveAttachment(ctx context.Context, contractID string, file *multipart.FileHeader) (domain.Contract, domain.ContractAttachment, error) {
	if file == nil {
		return domain.Contract{}, domain.ContractAttachment{}, fmt.Errorf("file is required")
	}
	contractID = strings.TrimSpace(contractID)
	contract, err := s.repo.GetContract(ctx, contractID)
	if err != nil {
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	attachmentID := strconv.FormatInt(time.Now().UnixNano(), 10)
	storagePath := s.attachmentPath(contractID, attachmentID)
	if err := os.MkdirAll(filepath.Dir(storagePath), 0o755); err != nil {
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	src, err := file.Open()
	if err != nil {
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	defer src.Close()
	dst, err := os.Create(storagePath)
	if err != nil {
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	if _, err := dst.ReadFrom(src); err != nil {
		dst.Close()
		_ = os.Remove(storagePath)
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	if err := dst.Close(); err != nil {
		return domain.Contract{}, domain.ContractAttachment{}, err
	}

	attachment := domain.ContractAttachment{
		AttachmentID: attachmentID,
		FileName:     strings.TrimSpace(file.Filename),
		FileSize:     file.Size,
		MimeType:     strings.TrimSpace(file.Header.Get("Content-Type")),
		UploadedAt:   time.Now().Format(time.RFC3339),
	}
	if err := s.repo.SaveAttachment(ctx, domain.ContractAttachmentBlob{
		ContractID:   contractID,
		AttachmentID: attachmentID,
		FileName:     attachment.FileName,
		StoragePath:  storagePath,
		FileSize:     attachment.FileSize,
		MimeType:     attachment.MimeType,
		CreatedAt:    time.Now(),
	}); err != nil {
		_ = os.Remove(storagePath)
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	contract.Attachments = append(contract.Attachments, attachment)
	contract, err = validateAndBuildContract(contract, false)
	if err != nil {
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	if err := s.repo.SaveContract(ctx, contract); err != nil {
		return domain.Contract{}, domain.ContractAttachment{}, err
	}
	return contract, attachment, nil
}

func (s *ContractService) GetAttachment(ctx context.Context, contractID, attachmentID string) (domain.ContractAttachmentBlob, error) {
	return s.repo.GetAttachment(ctx, strings.TrimSpace(contractID), strings.TrimSpace(attachmentID))
}

func (s *ContractService) DeleteAttachment(ctx context.Context, contractID, attachmentID string) (domain.Contract, error) {
	contractID = strings.TrimSpace(contractID)
	attachmentID = strings.TrimSpace(attachmentID)
	contract, err := s.repo.GetContract(ctx, contractID)
	if err != nil {
		return domain.Contract{}, err
	}
	if err := s.deleteAttachmentFile(contractID, attachmentID); err != nil {
		return domain.Contract{}, err
	}
	if err := s.repo.DeleteAttachment(ctx, contractID, attachmentID); err != nil {
		return domain.Contract{}, err
	}
	next := make([]domain.ContractAttachment, 0, len(contract.Attachments))
	for _, item := range contract.Attachments {
		if item.AttachmentID == attachmentID {
			continue
		}
		next = append(next, item)
	}
	contract.Attachments = next
	contract, err = validateAndBuildContract(contract, false)
	if err != nil {
		return domain.Contract{}, err
	}
	if err := s.repo.SaveContract(ctx, contract); err != nil {
		return domain.Contract{}, err
	}
	return contract, nil
}

func validateAndBuildContract(contract domain.Contract, isCreate bool) (domain.Contract, error) {
	if strings.TrimSpace(contract.ContractName) == "" {
		return domain.Contract{}, fmt.Errorf("contract_name is required")
	}
	if strings.TrimSpace(contract.PeriodStart) == "" || strings.TrimSpace(contract.PeriodEnd) == "" {
		return domain.Contract{}, fmt.Errorf("period_start and period_end are required")
	}
	start, err := parseDate(contract.PeriodStart)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("invalid period_start: %v", err)
	}
	end, err := parseDate(contract.PeriodEnd)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("invalid period_end: %v", err)
	}
	if end.Before(start) {
		return domain.Contract{}, fmt.Errorf("period_end must be >= period_start")
	}
	if contract.PreTaxAmount < 0 {
		return domain.Contract{}, fmt.Errorf("pre_tax_amount must be >= 0")
	}
	if strings.TrimSpace(contract.Supplier) == "" {
		return domain.Contract{}, fmt.Errorf("supplier is required")
	}
	if strings.TrimSpace(contract.BusinessContact) == "" {
		return domain.Contract{}, fmt.Errorf("business_contact is required")
	}
	if strings.TrimSpace(contract.TechContact) == "" {
		return domain.Contract{}, fmt.Errorf("tech_contact is required")
	}
	if isCreate {
		contract.CreatedAt = time.Now().Format(time.RFC3339)
	}
	contract.PeriodStart = start.Format("2006-01-02")
	contract.PeriodEnd = end.Format("2006-01-02")
	contract.UpdatedAt = time.Now().Format(time.RFC3339)
	return contract, nil
}

func (s *ContractService) attachmentPath(contractID, attachmentID string) string {
	return filepath.Join(s.uploadDir, contractID, attachmentID)
}

func (s *ContractService) deleteAttachmentFile(contractID, attachmentID string) error {
	if err := os.Remove(s.attachmentPath(contractID, attachmentID)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
