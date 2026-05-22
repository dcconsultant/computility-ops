package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"computility-ops/backend/internal/domain"
)

type SupplierRepo struct {
	db *sql.DB
}

func NewSupplierRepo(dsn string) *SupplierRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &SupplierRepo{db: db}
}

func (r *SupplierRepo) SaveSupplier(ctx context.Context, supplier domain.Supplier) error {
	payload, err := json.Marshal(supplier)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO ops_suppliers (supplier_id, payload_json)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE payload_json = VALUES(payload_json), updated_at = CURRENT_TIMESTAMP
	`, supplier.SupplierID, string(payload))
	return err
}

func (r *SupplierRepo) GetSupplier(ctx context.Context, supplierID string) (domain.Supplier, error) {
	var payload string
	if err := r.db.QueryRowContext(ctx, `SELECT payload_json FROM ops_suppliers WHERE supplier_id = ?`, supplierID).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domain.Supplier{}, fmt.Errorf("supplier %s not found", supplierID)
		}
		return domain.Supplier{}, err
	}
	var out domain.Supplier
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return domain.Supplier{}, err
	}
	return out, nil
}

func (r *SupplierRepo) ListSuppliers(ctx context.Context) ([]domain.Supplier, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT payload_json FROM ops_suppliers ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Supplier, 0)
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var s domain.Supplier
		if err := json.Unmarshal([]byte(payload), &s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SupplierRepo) DeleteSupplier(ctx context.Context, supplierID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM ops_suppliers WHERE supplier_id = ?`, supplierID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("supplier %s not found", supplierID)
	}
	return nil
}
