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

	withDetailedConfigCol, err := hasServerDetailedConfigColumn(ctx, tx)
	if err != nil {
		return err
	}
	insertSQL := `
		INSERT INTO ops_servers (
			sn, manufacturer, model, psa, psa_hash, idc, environment, config_type, warranty_end_date, launch_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if withDetailedConfigCol {
		insertSQL = `
			INSERT INTO ops_servers (
				sn, manufacturer, model, detailed_config, psa, psa_hash, idc, environment, config_type, warranty_end_date, launch_date
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
	}
	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range servers {
		if withDetailedConfigCol {
			if _, err := stmt.ExecContext(ctx,
				s.SN,
				s.Manufacturer,
				s.Model,
				nullIfEmpty(s.DetailedConfig),
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
			continue
		}
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
	withDetailedConfigCol, err := hasServerDetailedConfigColumn(ctx, r.db)
	if err != nil {
		return nil, err
	}
	querySQL := `
		SELECT sn, manufacturer, model, psa, idc, environment, config_type,
			COALESCE(warranty_end_date, ''), COALESCE(launch_date, '')
		FROM ops_servers
		ORDER BY created_at DESC
	`
	if withDetailedConfigCol {
		querySQL = `
			SELECT sn, manufacturer, model, COALESCE(detailed_config, ''), psa, idc, environment, config_type,
				COALESCE(warranty_end_date, ''), COALESCE(launch_date, '')
			FROM ops_servers
			ORDER BY created_at DESC
		`
	}
	rows, err := r.db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Server, 0)
	for rows.Next() {
		var s domain.Server
		if withDetailedConfigCol {
			if err := rows.Scan(
				&s.SN,
				&s.Manufacturer,
				&s.Model,
				&s.DetailedConfig,
				&s.PSA,
				&s.IDC,
				&s.Environment,
				&s.ConfigType,
				&s.WarrantyEndDate,
				&s.LaunchDate,
			); err != nil {
				return nil, err
			}
		} else {
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
			s.DetailedConfig = ""
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

func hasServerDetailedConfigColumn(ctx context.Context, q rowQueryer) (bool, error) {
	var count int
	err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'ops_servers'
		  AND COLUMN_NAME = 'detailed_config'
	`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}
