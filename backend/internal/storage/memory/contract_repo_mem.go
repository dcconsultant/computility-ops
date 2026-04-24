package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"computility-ops/backend/internal/domain"
)

type ContractRepo struct {
	mu          sync.RWMutex
	contracts   map[string]domain.Contract
	attachments map[string]domain.ContractAttachmentBlob
}

func NewContractRepo() *ContractRepo {
	return &ContractRepo{
		contracts:   map[string]domain.Contract{},
		attachments: map[string]domain.ContractAttachmentBlob{},
	}
}

func (r *ContractRepo) SaveContract(_ context.Context, contract domain.Contract) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contracts[contract.ContractID] = contract
	return nil
}

func (r *ContractRepo) GetContract(_ context.Context, contractID string) (domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.contracts[contractID]
	if !ok {
		return domain.Contract{}, fmt.Errorf("contract %s not found", contractID)
	}
	return v, nil
}

func (r *ContractRepo) ListContracts(_ context.Context) ([]domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Contract, 0, len(r.contracts))
	for _, c := range r.contracts {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *ContractRepo) DeleteContract(_ context.Context, contractID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.contracts[contractID]; !ok {
		return fmt.Errorf("contract %s not found", contractID)
	}
	delete(r.contracts, contractID)
	for k, v := range r.attachments {
		if v.ContractID == contractID {
			delete(r.attachments, k)
		}
	}
	return nil
}

func (r *ContractRepo) SaveAttachment(_ context.Context, attachment domain.ContractAttachmentBlob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.contracts[attachment.ContractID]; !ok {
		return fmt.Errorf("contract %s not found", attachment.ContractID)
	}
	r.attachments[attachment.ContractID+"|"+attachment.AttachmentID] = attachment
	return nil
}

func (r *ContractRepo) GetAttachment(_ context.Context, contractID, attachmentID string) (domain.ContractAttachmentBlob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.attachments[contractID+"|"+attachmentID]
	if !ok {
		return domain.ContractAttachmentBlob{}, fmt.Errorf("attachment %s not found", attachmentID)
	}
	return v, nil
}

func (r *ContractRepo) DeleteAttachment(_ context.Context, contractID, attachmentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := contractID + "|" + attachmentID
	if _, ok := r.attachments[key]; !ok {
		return fmt.Errorf("attachment %s not found", attachmentID)
	}
	delete(r.attachments, key)
	return nil
}
