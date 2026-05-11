package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"computility-ops/backend/internal/domain"
)

func (r *DatasetRepo) GetCabinetUtilization(ctx context.Context) (domain.CabinetUtilizationSetting, error) {
	var v float64
	err := r.db.QueryRowContext(ctx, `SELECT utilization FROM ops_cabinet_settings WHERE id=1`).Scan(&v)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.CabinetUtilizationSetting{Utilization: 1}, nil
		}
		return domain.CabinetUtilizationSetting{}, err
	}
	if v <= 0 {
		v = 1
	}
	return domain.CabinetUtilizationSetting{Utilization: v}, nil
}

func (r *DatasetRepo) SetCabinetUtilization(ctx context.Context, setting domain.CabinetUtilizationSetting) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ops_cabinet_settings (id, utilization)
		VALUES (1, ?)
		ON DUPLICATE KEY UPDATE utilization = VALUES(utilization)
	`, setting.Utilization)
	return err
}

func (r *DatasetRepo) CreateCabinetConfig(ctx context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO ops_cabinet_configs (idc, rated_power_kw, monthly_rent)
		VALUES (?, ?, ?)
	`, row.IDC, row.RatedPowerKW, row.MonthlyRent)
	if err != nil {
		if isDuplicateErr(err) {
			return domain.CabinetConfig{}, fmt.Errorf("duplicate key")
		}
		return domain.CabinetConfig{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.CabinetConfig{}, err
	}
	row.ID = id
	return row, nil
}

func (r *DatasetRepo) UpdateCabinetConfig(ctx context.Context, row domain.CabinetConfig) (domain.CabinetConfig, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE ops_cabinet_configs
		SET idc=?, rated_power_kw=?, monthly_rent=?
		WHERE id=?
	`, row.IDC, row.RatedPowerKW, row.MonthlyRent, row.ID)
	if err != nil {
		if isDuplicateErr(err) {
			return domain.CabinetConfig{}, fmt.Errorf("duplicate key")
		}
		return domain.CabinetConfig{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return domain.CabinetConfig{}, fmt.Errorf("not found")
	}
	return row, nil
}

func (r *DatasetRepo) DeleteCabinetConfig(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ops_cabinet_configs WHERE id=?`, id)
	return err
}

func (r *DatasetRepo) ListCabinetConfigs(ctx context.Context) ([]domain.CabinetConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, idc, rated_power_kw, monthly_rent
		FROM ops_cabinet_configs
		ORDER BY idc ASC, rated_power_kw ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.CabinetConfig, 0)
	for rows.Next() {
		var item domain.CabinetConfig
		if err := rows.Scan(&item.ID, &item.IDC, &item.RatedPowerKW, &item.MonthlyRent); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) GetValueScoreCostParams(ctx context.Context) (domain.ValueScoreCostParams, error) {
	var out domain.ValueScoreCostParams
	err := r.db.QueryRowContext(ctx, `
		SELECT depreciation_months, network_device_share_cny, server_renewal_fee_cny,
			cabinet_utilization, rated_power_kw, monthly_rent_cny
		FROM ops_value_score_cost_params
		WHERE id=1
	`).Scan(&out.DepreciationMonths, &out.NetworkDeviceShareCNY, &out.ServerRenewalFeeCNY, &out.CabinetUtilization, &out.RatedPowerKW, &out.MonthlyRentCNY)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ValueScoreCostParams{DepreciationMonths: 60, CabinetUtilization: 1}, nil
		}
		return domain.ValueScoreCostParams{}, err
	}
	if out.DepreciationMonths <= 0 {
		out.DepreciationMonths = 60
	}
	if out.NetworkDeviceShareCNY < 0 {
		out.NetworkDeviceShareCNY = 0
	}
	if out.ServerRenewalFeeCNY < 0 {
		out.ServerRenewalFeeCNY = 0
	}
	if out.CabinetUtilization <= 0 {
		out.CabinetUtilization = 1
	}
	if out.RatedPowerKW < 0 {
		out.RatedPowerKW = 0
	}
	if out.MonthlyRentCNY < 0 {
		out.MonthlyRentCNY = 0
	}
	return out, nil
}

func (r *DatasetRepo) SetValueScoreCostParams(ctx context.Context, params domain.ValueScoreCostParams) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ops_value_score_cost_params (id, depreciation_months, network_device_share_cny, server_renewal_fee_cny, cabinet_utilization, rated_power_kw, monthly_rent_cny)
		VALUES (1, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			depreciation_months = VALUES(depreciation_months),
			network_device_share_cny = VALUES(network_device_share_cny),
			server_renewal_fee_cny = VALUES(server_renewal_fee_cny),
			cabinet_utilization = VALUES(cabinet_utilization),
			rated_power_kw = VALUES(rated_power_kw),
			monthly_rent_cny = VALUES(monthly_rent_cny)
	`, params.DepreciationMonths, params.NetworkDeviceShareCNY, params.ServerRenewalFeeCNY, params.CabinetUtilization, params.RatedPowerKW, params.MonthlyRentCNY)
	return err
}

func (r *DatasetRepo) ReplaceValueScoreOriginalValues(ctx context.Context, rows []domain.ValueScoreOriginalValue) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_value_score_original_values`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_value_score_original_values (config_type, server_original_cny)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx, x.ConfigType, x.ServerOriginalCNY); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListValueScoreOriginalValues(ctx context.Context) ([]domain.ValueScoreOriginalValue, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT config_type, server_original_cny
		FROM ops_value_score_original_values
		ORDER BY config_type ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ValueScoreOriginalValue, 0)
	for rows.Next() {
		var x domain.ValueScoreOriginalValue
		if err := rows.Scan(&x.ConfigType, &x.ServerOriginalCNY); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceValueScorePerformanceParams(ctx context.Context, rows []domain.ValueScorePerformanceParam) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_value_score_performance_params`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_value_score_performance_params (config_type, unavailable_cores, unavailable_memory_gb, performance_score)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx, x.ConfigType, x.UnavailableCores, x.UnavailableMemoryGB, x.PerformanceScore); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ReplaceValueScoreConfigParams(ctx context.Context, originals []domain.ValueScoreOriginalValue, performance []domain.ValueScorePerformanceParam) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_value_score_original_values`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_value_score_performance_params`); err != nil {
		return err
	}

	origStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_value_score_original_values (config_type, server_original_cny)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer origStmt.Close()
	for _, x := range originals {
		if _, err := origStmt.ExecContext(ctx, x.ConfigType, x.ServerOriginalCNY); err != nil {
			return err
		}
	}

	perfStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_value_score_performance_params (config_type, unavailable_cores, unavailable_memory_gb, performance_score)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer perfStmt.Close()
	for _, x := range performance {
		if _, err := perfStmt.ExecContext(ctx, x.ConfigType, x.UnavailableCores, x.UnavailableMemoryGB, x.PerformanceScore); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListValueScorePerformanceParams(ctx context.Context) ([]domain.ValueScorePerformanceParam, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT config_type, unavailable_cores, unavailable_memory_gb, performance_score
		FROM ops_value_score_performance_params
		ORDER BY config_type ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ValueScorePerformanceParam, 0)
	for rows.Next() {
		var x domain.ValueScorePerformanceParam
		if err := rows.Scan(&x.ConfigType, &x.UnavailableCores, &x.UnavailableMemoryGB, &x.PerformanceScore); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") || strings.Contains(s, "1062")
}
