package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"computility-ops/backend/internal/domain"
)

type ServerRepo struct {
	db *sql.DB
}

func NewServerRepo(dsn string) *ServerRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &ServerRepo{db: db}
}

func (r *ServerRepo) ReplaceAll(ctx context.Context, servers []domain.Server) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_servers`); err != nil {
		return fmt.Errorf("clear ops_servers failed: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_servers (
			sn, manufacturer, model, psa, psa_hash, idc, environment, config_type, warranty_end_date, launch_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range servers {
		if _, err := stmt.ExecContext(ctx,
			s.SN,
			s.Manufacturer,
			s.Model,
			s.PSA,
			psaHash(s.PSA),
			nullIfEmpty(s.IDC),
			nullIfEmpty(s.Environment),
			s.ConfigType,
			nullIfEmpty(s.WarrantyEndDate),
			nullIfEmpty(s.LaunchDate),
		); err != nil {
			return fmt.Errorf("insert server %s failed: %w", s.SN, err)
		}
	}

	return tx.Commit()
}

func (r *ServerRepo) List(ctx context.Context) ([]domain.Server, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sn, manufacturer, model, psa, idc, environment, config_type,
			COALESCE(warranty_end_date, ''), COALESCE(launch_date, '')
		FROM ops_servers
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Server, 0)
	for rows.Next() {
		var s domain.Server
		if err := rows.Scan(
			&s.SN,
			&s.Manufacturer,
			&s.Model,
			&s.PSA,
			&s.IDC,
			&s.Environment,
			&s.ConfigType,
			&s.WarrantyEndDate,
			&s.LaunchDate,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ServerRepo) Clear(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ops_servers`)
	return err
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
