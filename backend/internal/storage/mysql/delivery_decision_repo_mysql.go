package mysql

import (
	"context"
	"database/sql"
	"strings"

	"computility-ops/backend/internal/domain"
)

type DeliveryDecisionRepo struct {
	db *sql.DB
}

func NewDeliveryDecisionRepo(dsn string) *DeliveryDecisionRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &DeliveryDecisionRepo{db: db}
}

func (r *DeliveryDecisionRepo) GetConfig(ctx context.Context) (domain.DeliveryDecisionConfigState, bool, error) {
	var payload string
	err := r.db.QueryRowContext(ctx, `SELECT payload_json FROM ops_delivery_decision_config WHERE id = 1`).Scan(&payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.DeliveryDecisionConfigState{}, false, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return domain.DeliveryDecisionConfigState{}, false, nil
		}
		return domain.DeliveryDecisionConfigState{}, false, err
	}
	var out domain.DeliveryDecisionConfigState
	if err := unmarshalJSONPayload(payload, &out); err != nil {
		return domain.DeliveryDecisionConfigState{}, false, err
	}
	return out, true, nil
}

func (r *DeliveryDecisionRepo) SaveConfig(ctx context.Context, state domain.DeliveryDecisionConfigState) error {
	payload, err := marshalJSONPayload(state)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO ops_delivery_decision_config (id, payload_json)
		VALUES (1, ?)
		ON DUPLICATE KEY UPDATE payload_json = VALUES(payload_json), updated_at = CURRENT_TIMESTAMP
	`, string(payload))
	return err
}
