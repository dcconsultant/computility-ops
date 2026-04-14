package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"computility-ops/backend/internal/domain"
)

type DatasetRepo struct {
	db *sql.DB
}

func NewDatasetRepo(dsn string) *DatasetRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &DatasetRepo{db: db}
}

func (r *DatasetRepo) ReplaceHostPackages(ctx context.Context, rows []domain.HostPackageConfig) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_host_packages`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_host_packages (
			config_type, scene_category, cpu_logical_cores, gpu_card_count,
			data_disk_type, data_disk_count, storage_capacity_tb,
			server_value_score, arch_standardized_factor
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx,
			x.ConfigType,
			nullIfEmpty(x.SceneCategory),
			x.CPULogicalCores,
			x.GPUCardCount,
			nullIfEmpty(x.DataDiskType),
			x.DataDiskCount,
			x.StorageCapacityTB,
			x.ServerValueScore,
			x.ArchStandardizedFactor,
		); err != nil {
			return fmt.Errorf("insert host package %s failed: %w", x.ConfigType, err)
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListHostPackages(ctx context.Context) ([]domain.HostPackageConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT config_type, COALESCE(scene_category,''), cpu_logical_cores, gpu_card_count,
			COALESCE(data_disk_type,''), data_disk_count, storage_capacity_tb,
			server_value_score, arch_standardized_factor
		FROM ops_host_packages
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.HostPackageConfig, 0)
	for rows.Next() {
		var x domain.HostPackageConfig
		if err := rows.Scan(
			&x.ConfigType,
			&x.SceneCategory,
			&x.CPULogicalCores,
			&x.GPUCardCount,
			&x.DataDiskType,
			&x.DataDiskCount,
			&x.StorageCapacityTB,
			&x.ServerValueScore,
			&x.ArchStandardizedFactor,
		); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceSpecialRules(ctx context.Context, rows []domain.SpecialRule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_special_rules`); err != nil {
		return err
	}
	withReasonCol, err := hasSpecialRuleReasonColumn(ctx, tx)
	if err != nil {
		return err
	}
	insertSQL := `
		INSERT INTO ops_special_rules (
			sn, manufacturer, model, psa, psa_hash, idc, package_type, warranty_end_date, launch_date, policy
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if withReasonCol {
		insertSQL = `
			INSERT INTO ops_special_rules (
				sn, manufacturer, model, psa, psa_hash, idc, package_type, warranty_end_date, launch_date, policy, reason
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
	}
	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if withReasonCol {
			if _, err := stmt.ExecContext(ctx,
				x.SN,
				nullIfEmpty(x.Manufacturer),
				nullIfEmpty(x.Model),
				nullIfEmpty(x.PSA),
				nullPSAHash(x.PSA),
				nullIfEmpty(x.IDC),
				nullIfEmpty(x.PackageType),
				nullIfEmpty(x.WarrantyEndDate),
				nullIfEmpty(x.LaunchDate),
				x.Policy,
				nullIfEmpty(x.Reason),
			); err != nil {
				return fmt.Errorf("insert special rule %s failed: %w", x.SN, err)
			}
			continue
		}
		if _, err := stmt.ExecContext(ctx,
			x.SN,
			nullIfEmpty(x.Manufacturer),
			nullIfEmpty(x.Model),
			nullIfEmpty(x.PSA),
			nullPSAHash(x.PSA),
			nullIfEmpty(x.IDC),
			nullIfEmpty(x.PackageType),
			nullIfEmpty(x.WarrantyEndDate),
			nullIfEmpty(x.LaunchDate),
			x.Policy,
		); err != nil {
			return fmt.Errorf("insert special rule %s failed: %w", x.SN, err)
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListSpecialRules(ctx context.Context) ([]domain.SpecialRule, error) {
	withReasonCol, err := hasSpecialRuleReasonColumn(ctx, r.db)
	if err != nil {
		return nil, err
	}
	querySQL := `
		SELECT sn, COALESCE(manufacturer,''), COALESCE(model,''), COALESCE(psa,''),
			COALESCE(idc,''), COALESCE(package_type,''), COALESCE(warranty_end_date,''),
			COALESCE(launch_date,''), policy
		FROM ops_special_rules
		ORDER BY created_at DESC
	`
	if withReasonCol {
		querySQL = `
			SELECT sn, COALESCE(manufacturer,''), COALESCE(model,''), COALESCE(psa,''),
				COALESCE(idc,''), COALESCE(package_type,''), COALESCE(warranty_end_date,''),
				COALESCE(launch_date,''), policy, COALESCE(reason,'')
			FROM ops_special_rules
			ORDER BY created_at DESC
		`
	}
	rows, err := r.db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SpecialRule, 0)
	for rows.Next() {
		var x domain.SpecialRule
		if withReasonCol {
			if err := rows.Scan(
				&x.SN,
				&x.Manufacturer,
				&x.Model,
				&x.PSA,
				&x.IDC,
				&x.PackageType,
				&x.WarrantyEndDate,
				&x.LaunchDate,
				&x.Policy,
				&x.Reason,
			); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(
				&x.SN,
				&x.Manufacturer,
				&x.Model,
				&x.PSA,
				&x.IDC,
				&x.PackageType,
				&x.WarrantyEndDate,
				&x.LaunchDate,
				&x.Policy,
			); err != nil {
				return nil, err
			}
			x.Reason = ""
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceModelFailureRates(ctx context.Context, rows []domain.ModelFailureRate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_model_failure_rates`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_model_failure_rates (manufacturer, model, failure_rate, over_warranty_failure_rate)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx, x.Manufacturer, x.Model, x.FailureRate, x.OverWarrantyFailureRate); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListModelFailureRates(ctx context.Context) ([]domain.ModelFailureRate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT manufacturer, model, failure_rate, over_warranty_failure_rate
		FROM ops_model_failure_rates
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ModelFailureRate, 0)
	for rows.Next() {
		var x domain.ModelFailureRate
		if err := rows.Scan(&x.Manufacturer, &x.Model, &x.FailureRate, &x.OverWarrantyFailureRate); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplacePackageFailureRates(ctx context.Context, rows []domain.PackageFailureRate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_package_failure_rates`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_package_failure_rates (period, stat_year, config_type, failure_rate, over_warranty_failure_rate)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx, x.Period, x.Year, x.ConfigType, x.FailureRate, x.OverWarrantyFailureRate); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListPackageFailureRates(ctx context.Context) ([]domain.PackageFailureRate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT period, stat_year, config_type, failure_rate, over_warranty_failure_rate
		FROM ops_package_failure_rates
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.PackageFailureRate, 0)
	for rows.Next() {
		var x domain.PackageFailureRate
		if err := rows.Scan(&x.Period, &x.Year, &x.ConfigType, &x.FailureRate, &x.OverWarrantyFailureRate); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplacePackageModelFailureRates(ctx context.Context, rows []domain.PackageModelFailureRate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_package_model_failure_rates`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_package_model_failure_rates (
			period, stat_year, config_type, manufacturer, model, failure_rate, over_warranty_failure_rate
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx, x.Period, x.Year, x.ConfigType, x.Manufacturer, x.Model, x.FailureRate, x.OverWarrantyFailureRate); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListPackageModelFailureRates(ctx context.Context) ([]domain.PackageModelFailureRate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT period, stat_year, config_type, manufacturer, model, failure_rate, over_warranty_failure_rate
		FROM ops_package_model_failure_rates
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.PackageModelFailureRate, 0)
	for rows.Next() {
		var x domain.PackageModelFailureRate
		if err := rows.Scan(&x.Period, &x.Year, &x.ConfigType, &x.Manufacturer, &x.Model, &x.FailureRate, &x.OverWarrantyFailureRate); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceOverallFailureRates(ctx context.Context, rows []domain.FailureRateSummary) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_overall_failure_rates`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_overall_failure_rates (
			period, stat_year, scope, segment,
			full_cycle_failure_rate, over_warranty_failure_rate,
			fault_count, over_warranty_fault_count,
			server_years, over_warranty_years
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx,
			x.Period,
			x.Year,
			x.Scope,
			x.Segment,
			x.FullCycleFailureRate,
			x.OverWarrantyRate,
			x.FaultCount,
			x.OverWarrantyFaults,
			x.ServerYears,
			x.OverWarrantyYears,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListOverallFailureRates(ctx context.Context) ([]domain.FailureRateSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT period, stat_year, scope, segment,
			full_cycle_failure_rate, over_warranty_failure_rate,
			fault_count, over_warranty_fault_count,
			server_years, over_warranty_years
		FROM ops_overall_failure_rates
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FailureRateSummary, 0)
	for rows.Next() {
		var x domain.FailureRateSummary
		if err := rows.Scan(
			&x.Period,
			&x.Year,
			&x.Scope,
			&x.Segment,
			&x.FullCycleFailureRate,
			&x.OverWarrantyRate,
			&x.FaultCount,
			&x.OverWarrantyFaults,
			&x.ServerYears,
			&x.OverWarrantyYears,
		); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceFailureOverviewCards(ctx context.Context, rows []domain.FailureOverviewCard) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM failure_overview_cards`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO failure_overview_cards (
			segment, stat_year, current_year_fault_rate, history_avg_fault_rate,
			current_year_fault_count, current_year_denominator,
			history_fault_count, history_denominator
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx,
			x.Segment,
			x.Year,
			x.CurrentYearFaultRate,
			x.HistoryAvgFaultRate,
			x.CurrentYearFaultCount,
			x.CurrentYearDenominator,
			x.HistoryFaultCount,
			x.HistoryDenominator,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListFailureOverviewCards(ctx context.Context) ([]domain.FailureOverviewCard, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT segment, stat_year, current_year_fault_rate, history_avg_fault_rate,
			current_year_fault_count, current_year_denominator,
			history_fault_count, history_denominator
		FROM failure_overview_cards
		ORDER BY stat_year DESC, segment ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FailureOverviewCard, 0)
	for rows.Next() {
		var x domain.FailureOverviewCard
		if err := rows.Scan(
			&x.Segment,
			&x.Year,
			&x.CurrentYearFaultRate,
			&x.HistoryAvgFaultRate,
			&x.CurrentYearFaultCount,
			&x.CurrentYearDenominator,
			&x.HistoryFaultCount,
			&x.HistoryDenominator,
		); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceFailureAgeTrendPoints(ctx context.Context, rows []domain.FailureAgeTrendPoint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM failure_age_trend_points`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO failure_age_trend_points (
			segment, age_bucket, numerator_fault_count, denominator_exposure, fault_rate
		) VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx,
			x.Segment,
			x.AgeBucket,
			x.NumeratorFaultCount,
			x.DenominatorExposure,
			x.FaultRate,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListFailureAgeTrendPoints(ctx context.Context) ([]domain.FailureAgeTrendPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT segment, age_bucket, numerator_fault_count, denominator_exposure, fault_rate
		FROM failure_age_trend_points
		ORDER BY segment ASC, age_bucket ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FailureAgeTrendPoint, 0)
	for rows.Next() {
		var x domain.FailureAgeTrendPoint
		if err := rows.Scan(&x.Segment, &x.AgeBucket, &x.NumeratorFaultCount, &x.DenominatorExposure, &x.FaultRate); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceFailureFeatureFacts(ctx context.Context, rows []domain.FailureFeatureFact) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_failure_feature_facts`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_failure_feature_facts (
			record_year_index, record_year_start, record_year_end,
			scope, scene_group, age_bucket,
			denominator_weighted, fault_count, fault_rate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if _, err := stmt.ExecContext(ctx,
			x.RecordYearIndex,
			x.RecordYearStart,
			x.RecordYearEnd,
			x.Scope,
			x.SceneGroup,
			x.AgeBucket,
			x.DenominatorWeighted,
			x.FaultCount,
			x.FaultRate,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListFailureFeatureFacts(ctx context.Context) ([]domain.FailureFeatureFact, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT record_year_index, DATE_FORMAT(record_year_start, '%Y-%m-%d'), DATE_FORMAT(record_year_end, '%Y-%m-%d'),
			scope, scene_group, age_bucket, denominator_weighted, fault_count, fault_rate
		FROM ops_failure_feature_facts
		ORDER BY record_year_index ASC, scope ASC, scene_group ASC, age_bucket ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FailureFeatureFact, 0)
	for rows.Next() {
		var x domain.FailureFeatureFact
		if err := rows.Scan(
			&x.RecordYearIndex,
			&x.RecordYearStart,
			&x.RecordYearEnd,
			&x.Scope,
			&x.SceneGroup,
			&x.AgeBucket,
			&x.DenominatorWeighted,
			&x.FaultCount,
			&x.FaultRate,
		); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *DatasetRepo) ReplaceStorageTopServerRates(ctx context.Context, rows []domain.StorageTopServerRate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_storage_top_server_rates`); err != nil {
		return err
	}
	withCapacityCols, err := hasStorageTopCapacityColumns(ctx, tx)
	if err != nil {
		return err
	}
	insertSQL := `
		INSERT INTO ops_storage_top_server_rates (
			sn, manufacturer, model, config_type, environment, idc,
			data_disk_count, fault_count, denominator, fault_rate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if withCapacityCols {
		insertSQL = `
			INSERT INTO ops_storage_top_server_rates (
				sn, manufacturer, model, config_type, environment, idc,
				data_disk_count, single_disk_capacity_tb, total_capacity_tb,
				fault_count, denominator, fault_rate
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
	}
	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, x := range rows {
		if withCapacityCols {
			if _, err := stmt.ExecContext(ctx,
				x.SN, x.Manufacturer, x.Model, x.ConfigType, x.Environment, x.IDC,
				x.DataDiskCount, x.SingleDiskCapacityTB, x.TotalCapacityTB,
				x.FaultCount, x.Denominator, x.FaultRate,
			); err != nil {
				return err
			}
			continue
		}
		if _, err := stmt.ExecContext(ctx,
			x.SN, x.Manufacturer, x.Model, x.ConfigType, x.Environment, x.IDC,
			x.DataDiskCount, x.FaultCount, x.Denominator, x.FaultRate,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *DatasetRepo) ListStorageTopServerRates(ctx context.Context) ([]domain.StorageTopServerRate, error) {
	withCapacityCols, err := hasStorageTopCapacityColumns(ctx, r.db)
	if err != nil {
		return nil, err
	}
	querySQL := `
		SELECT sn, manufacturer, model, config_type, environment, idc,
			data_disk_count, fault_count, denominator, fault_rate
		FROM ops_storage_top_server_rates
		ORDER BY fault_rate DESC, fault_count DESC, sn ASC
	`
	if withCapacityCols {
		querySQL = `
			SELECT sn, manufacturer, model, config_type, environment, idc,
				data_disk_count, single_disk_capacity_tb, total_capacity_tb,
				fault_count, denominator, fault_rate
			FROM ops_storage_top_server_rates
			ORDER BY fault_rate DESC, fault_count DESC, sn ASC
		`
	}
	rows, err := r.db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.StorageTopServerRate, 0)
	for rows.Next() {
		var x domain.StorageTopServerRate
		if withCapacityCols {
			if err := rows.Scan(
				&x.SN, &x.Manufacturer, &x.Model, &x.ConfigType, &x.Environment, &x.IDC,
				&x.DataDiskCount, &x.SingleDiskCapacityTB, &x.TotalCapacityTB,
				&x.FaultCount, &x.Denominator, &x.FaultRate,
			); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(
				&x.SN, &x.Manufacturer, &x.Model, &x.ConfigType, &x.Environment, &x.IDC,
				&x.DataDiskCount, &x.FaultCount, &x.Denominator, &x.FaultRate,
			); err != nil {
				return nil, err
			}
			x.SingleDiskCapacityTB = 0
			x.TotalCapacityTB = 0
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

type rowQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func hasStorageTopCapacityColumns(ctx context.Context, q rowQueryer) (bool, error) {
	var count int
	err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'ops_storage_top_server_rates'
		  AND COLUMN_NAME IN ('single_disk_capacity_tb', 'total_capacity_tb')
	`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 2, nil
}

func hasSpecialRuleReasonColumn(ctx context.Context, q rowQueryer) (bool, error) {
	var count int
	err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'ops_special_rules'
		  AND COLUMN_NAME = 'reason'
	`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}
