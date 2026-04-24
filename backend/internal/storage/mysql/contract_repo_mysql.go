package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
)

type ContractRepo struct {
	db *sql.DB
}

func NewContractRepo(dsn string) *ContractRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &ContractRepo{db: db}
}

func (r *ContractRepo) SaveContract(ctx context.Context, contract domain.Contract) error {
	payload, err := json.Marshal(contract)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO ops_contracts (contract_id, payload_json)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE payload_json = VALUES(payload_json), updated_at = CURRENT_TIMESTAMP
	`, contract.ContractID, string(payload))
	return err
}

func (r *ContractRepo) GetContract(ctx context.Context, contractID string) (domain.Contract, error) {
	var payload string
	if err := r.db.QueryRowContext(ctx, `SELECT payload_json FROM ops_contracts WHERE contract_id = ?`, contractID).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domain.Contract{}, fmt.Errorf("contract %s not found", contractID)
		}
		return domain.Contract{}, err
	}
	var out domain.Contract
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return domain.Contract{}, err
	}
	return out, nil
}

func (r *ContractRepo) ListContracts(ctx context.Context) ([]domain.Contract, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT payload_json FROM ops_contracts ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Contract, 0)
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var c domain.Contract
		if err := json.Unmarshal([]byte(payload), &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortContracts(out)
	return out, nil
}

func (r *ContractRepo) DeleteContract(ctx context.Context, contractID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM ops_contracts WHERE contract_id = ?`, contractID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("contract %s not found", contractID)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_contract_attachments WHERE contract_id = ?`, contractID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ContractRepo) SaveAttachment(ctx context.Context, attachment domain.ContractAttachmentBlob) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ops_contract_attachments (attachment_id, contract_id, file_name, storage_path, file_size, mime_type)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE file_name = VALUES(file_name), storage_path = VALUES(storage_path), file_size = VALUES(file_size), mime_type = VALUES(mime_type)
	`, attachment.AttachmentID, attachment.ContractID, attachment.FileName, attachment.StoragePath, attachment.FileSize, attachment.MimeType)
	return err
}

func (r *ContractRepo) GetAttachment(ctx context.Context, contractID, attachmentID string) (domain.ContractAttachmentBlob, error) {
	var out domain.ContractAttachmentBlob
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT contract_id, attachment_id, file_name, storage_path, file_size, mime_type, created_at
		FROM ops_contract_attachments
		WHERE contract_id = ? AND attachment_id = ?
	`, contractID, attachmentID).Scan(&out.ContractID, &out.AttachmentID, &out.FileName, &out.StoragePath, &out.FileSize, &out.MimeType, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ContractAttachmentBlob{}, fmt.Errorf("attachment %s not found", attachmentID)
		}
		return domain.ContractAttachmentBlob{}, err
	}
	out.CreatedAt = createdAt
	return out, nil
}

func (r *ContractRepo) DeleteAttachment(ctx context.Context, contractID, attachmentID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM ops_contract_attachments WHERE contract_id = ? AND attachment_id = ?`, contractID, attachmentID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("attachment %s not found", attachmentID)
	}
	return nil
}

func sortContracts(list []domain.Contract) {
	sort.Slice(list, func(i, j int) bool {
		ia, ea := strconv.ParseInt(strings.TrimSpace(list[i].ContractID), 10, 64)
		ib, eb := strconv.ParseInt(strings.TrimSpace(list[j].ContractID), 10, 64)
		if ea == nil && eb == nil {
			return ia > ib
		}
		return strings.TrimSpace(list[i].UpdatedAt) > strings.TrimSpace(list[j].UpdatedAt)
	})
}
