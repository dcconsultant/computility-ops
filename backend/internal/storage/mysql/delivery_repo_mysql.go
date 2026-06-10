package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"computility-ops/backend/internal/domain"
)

type DeliveryRepo struct {
	db *sql.DB
}

func NewDeliveryRepo(dsn string) *DeliveryRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &DeliveryRepo{db: db}
}

func (r *DeliveryRepo) SaveArrivalPlan(ctx context.Context, item domain.ArrivalPlan) error {
	return r.savePayload(ctx, "ops_delivery_arrival_plans", "plan_id", item.PlanID, item)
}

func (r *DeliveryRepo) GetArrivalPlan(ctx context.Context, planID string) (domain.ArrivalPlan, error) {
	var out domain.ArrivalPlan
	if err := r.getPayload(ctx, "ops_delivery_arrival_plans", "plan_id", planID, &out); err != nil {
		return domain.ArrivalPlan{}, fmt.Errorf("arrival plan %s not found", planID)
	}
	return out, nil
}

func (r *DeliveryRepo) ListArrivalPlans(ctx context.Context) ([]domain.ArrivalPlan, error) {
	var out []domain.ArrivalPlan
	if err := r.listPayloads(ctx, "ops_delivery_arrival_plans", &out); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *DeliveryRepo) DeleteArrivalPlan(ctx context.Context, planID string) error {
	return r.deletePayload(ctx, "ops_delivery_arrival_plans", "plan_id", planID, "arrival plan")
}

func (r *DeliveryRepo) SaveDeviceArrival(ctx context.Context, item domain.DeviceArrivalRecord) error {
	return r.savePayload(ctx, "ops_delivery_device_arrivals", "record_id", item.RecordID, item)
}

func (r *DeliveryRepo) GetDeviceArrival(ctx context.Context, recordID string) (domain.DeviceArrivalRecord, error) {
	var out domain.DeviceArrivalRecord
	if err := r.getPayload(ctx, "ops_delivery_device_arrivals", "record_id", recordID, &out); err != nil {
		return domain.DeviceArrivalRecord{}, fmt.Errorf("device arrival %s not found", recordID)
	}
	return out, nil
}

func (r *DeliveryRepo) ListDeviceArrivals(ctx context.Context) ([]domain.DeviceArrivalRecord, error) {
	var out []domain.DeviceArrivalRecord
	if err := r.listPayloads(ctx, "ops_delivery_device_arrivals", &out); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *DeliveryRepo) DeleteDeviceArrival(ctx context.Context, recordID string) error {
	return r.deletePayload(ctx, "ops_delivery_device_arrivals", "record_id", recordID, "device arrival")
}

func (r *DeliveryRepo) SaveAccessoryArrival(ctx context.Context, item domain.AccessoryArrivalRecord) error {
	return r.savePayload(ctx, "ops_delivery_accessory_arrivals", "record_id", item.RecordID, item)
}

func (r *DeliveryRepo) GetAccessoryArrival(ctx context.Context, recordID string) (domain.AccessoryArrivalRecord, error) {
	var out domain.AccessoryArrivalRecord
	if err := r.getPayload(ctx, "ops_delivery_accessory_arrivals", "record_id", recordID, &out); err != nil {
		return domain.AccessoryArrivalRecord{}, fmt.Errorf("accessory arrival %s not found", recordID)
	}
	return out, nil
}

func (r *DeliveryRepo) ListAccessoryArrivals(ctx context.Context) ([]domain.AccessoryArrivalRecord, error) {
	var out []domain.AccessoryArrivalRecord
	if err := r.listPayloads(ctx, "ops_delivery_accessory_arrivals", &out); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func (r *DeliveryRepo) DeleteAccessoryArrival(ctx context.Context, recordID string) error {
	return r.deletePayload(ctx, "ops_delivery_accessory_arrivals", "record_id", recordID, "accessory arrival")
}

func (r *DeliveryRepo) savePayload(ctx context.Context, table, idColumn, id string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (%s, payload_json)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE payload_json = VALUES(payload_json), updated_at = CURRENT_TIMESTAMP
	`, table, idColumn)
	_, err = r.db.ExecContext(ctx, query, id, string(body))
	return err
}

func (r *DeliveryRepo) getPayload(ctx context.Context, table, idColumn, id string, out any) error {
	query := fmt.Sprintf(`SELECT payload_json FROM %s WHERE %s = ?`, table, idColumn)
	var payload string
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&payload); err != nil {
		return err
	}
	return json.Unmarshal([]byte(payload), out)
}

func (r *DeliveryRepo) listPayloads(ctx context.Context, table string, out any) error {
	query := fmt.Sprintf(`SELECT payload_json FROM %s ORDER BY updated_at DESC`, table)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	switch target := out.(type) {
	case *[]domain.ArrivalPlan:
		for rows.Next() {
			var payload string
			if err := rows.Scan(&payload); err != nil {
				return err
			}
			var item domain.ArrivalPlan
			if err := json.Unmarshal([]byte(payload), &item); err != nil {
				return err
			}
			*target = append(*target, item)
		}
	case *[]domain.DeviceArrivalRecord:
		for rows.Next() {
			var payload string
			if err := rows.Scan(&payload); err != nil {
				return err
			}
			var item domain.DeviceArrivalRecord
			if err := json.Unmarshal([]byte(payload), &item); err != nil {
				return err
			}
			*target = append(*target, item)
		}
	case *[]domain.AccessoryArrivalRecord:
		for rows.Next() {
			var payload string
			if err := rows.Scan(&payload); err != nil {
				return err
			}
			var item domain.AccessoryArrivalRecord
			if err := json.Unmarshal([]byte(payload), &item); err != nil {
				return err
			}
			*target = append(*target, item)
		}
	default:
		return fmt.Errorf("unsupported payload list target")
	}
	return rows.Err()
}

func (r *DeliveryRepo) deletePayload(ctx context.Context, table, idColumn, id, label string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, table, idColumn)
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%s %s not found", label, id)
	}
	return nil
}
