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

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") || strings.Contains(s, "1062")
}
