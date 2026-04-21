package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"computility-ops/backend/internal/domain"
)

type RenewalRepo struct {
	db *sql.DB
}

func NewRenewalRepo(dsn string) *RenewalRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &RenewalRepo{db: db}
}

func (r *RenewalRepo) SavePlan(ctx context.Context, plan domain.RenewalPlan) error {
	payload, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO ops_renewal_plans (plan_id, payload_json)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE payload_json = VALUES(payload_json), updated_at = CURRENT_TIMESTAMP
	`, plan.PlanID, string(payload))
	return err
}

func (r *RenewalRepo) GetPlan(ctx context.Context, planID string) (domain.RenewalPlan, error) {
	var payload string
	if err := r.db.QueryRowContext(ctx, `SELECT payload_json FROM ops_renewal_plans WHERE plan_id = ?`, planID).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domain.RenewalPlan{}, fmt.Errorf("plan %s not found", planID)
		}
		return domain.RenewalPlan{}, err
	}
	var plan domain.RenewalPlan
	if err := json.Unmarshal([]byte(payload), &plan); err != nil {
		return domain.RenewalPlan{}, err
	}
	return plan, nil
}

func (r *RenewalRepo) ListPlans(ctx context.Context) ([]domain.RenewalPlan, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT payload_json FROM ops_renewal_plans ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.RenewalPlan, 0)
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var p domain.RenewalPlan
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		ii, _ := strconv.ParseInt(out[i].PlanID, 10, 64)
		jj, _ := strconv.ParseInt(out[j].PlanID, 10, 64)
		return ii > jj
	})
	return out, nil
}

func (r *RenewalRepo) DeletePlan(ctx context.Context, planID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM ops_renewal_plans WHERE plan_id = ?`, planID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("plan %s not found", planID)
	}
	return nil
}

func (r *RenewalRepo) ListUnitPrices(ctx context.Context) ([]domain.RenewalUnitPrice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT country, scene_category, unit_price
		FROM ops_renewal_unit_prices
		ORDER BY country ASC, scene_category ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.RenewalUnitPrice, 0)
	for rows.Next() {
		var x domain.RenewalUnitPrice
		if err := rows.Scan(&x.Country, &x.SceneCategory, &x.UnitPrice); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *RenewalRepo) ReplaceUnitPrices(ctx context.Context, prices []domain.RenewalUnitPrice) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM ops_renewal_unit_prices`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ops_renewal_unit_prices (country, scene_category, unit_price)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, p := range prices {
		if _, err := stmt.ExecContext(ctx, p.Country, p.SceneCategory, p.UnitPrice); err != nil {
			return err
		}
	}
	return tx.Commit()
}
