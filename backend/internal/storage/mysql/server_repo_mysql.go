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
	for _, modelCode := range serverMetadataModelCodes() {
		out, err := r.listMetaServers(ctx, modelCode)
		if err != nil {
			return nil, err
		}
		if len(out) > 0 {
			return out, nil
		}
	}

	// Fallback source for rollback/compatibility: legacy ops_servers
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

func (r *ServerRepo) listMetaServers(ctx context.Context, modelCode string) ([]domain.Server, error) {
	metaRows, err := r.db.QueryContext(ctx, `
		SELECT
			COALESCE(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.sn')), '') AS sn,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.manufacture')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.manufacturer')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."制造商"')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."厂商"')), ''),
				''
			) AS manufacturer,
			COALESCE(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.model')), '') AS model,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.detail_configuration')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.detailed_config')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."详细配置"')), ''),
				''
			) AS detailed_config,
			COALESCE(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.psa')), '') AS psa,
			COALESCE(JSON_UNQUOTE(JSON_EXTRACT(rk.data_json, '$.datacenter')), '') AS idc,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.env')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.environment')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."环境"')), ''),
				''
			) AS environment,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.config_type')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.package_type')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."配置类型"')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."套餐"')), ''),
				''
			) AS config_type,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.server_warranty_last_date')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.warranty_end_date')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."保修结束日期"')), ''),
				''
			) AS warranty_end_date,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.install_date')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.launch_date')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$."投产日期"')), ''),
				''
			) AS launch_date
		FROM md_record s
		INNER JOIN md_model ms ON ms.id = s.model_id
		LEFT JOIN md_model mr ON mr.model_code = 'rack'
		LEFT JOIN md_record rk
			ON rk.model_id = mr.id
			AND rk.deleted_flag = 0
			AND COALESCE(JSON_UNQUOTE(JSON_EXTRACT(rk.data_json, '$.sn')), '') = COALESCE(JSON_UNQUOTE(JSON_EXTRACT(s.data_json, '$.rack')), '')
		WHERE ms.model_code = ?
			AND s.deleted_flag = 0
		ORDER BY s.updated_at DESC
	`, modelCode)
	if err != nil {
		return nil, err
	}
	defer metaRows.Close()

	out := make([]domain.Server, 0)
	for metaRows.Next() {
		var s domain.Server
		if err := metaRows.Scan(
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
		if s.SN == "" {
			continue
		}
		out = append(out, s)
	}
	if err := metaRows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func serverMetadataModelCodes() []string {
	return []string{"server", "sever"}
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
