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
