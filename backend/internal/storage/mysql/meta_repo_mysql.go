package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type MetaRepo struct{ db *sql.DB }

func NewMetaRepo(dsn string) *MetaRepo {
	db, err := getDB(dsn)
	if err != nil {
		panic(err)
	}
	return &MetaRepo{db: db}
}

func (r *MetaRepo) CreateModel(ctx context.Context, model domain.MetaModel) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO md_model (id, model_code, model_name, description, status, current_version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, model.ID, model.ModelCode, model.ModelName, model.Description, string(model.Status), model.CurrentVersion, model.CreatedAt, model.UpdatedAt)
	return err
}
func (r *MetaRepo) GetModel(ctx context.Context, modelID string) (domain.MetaModel, error) {
	var m domain.MetaModel
	var status string
	err := r.db.QueryRowContext(ctx, `SELECT id, model_code, model_name, description, status, current_version, created_at, updated_at FROM md_model WHERE id=?`, modelID).Scan(&m.ID, &m.ModelCode, &m.ModelName, &m.Description, &status, &m.CurrentVersion, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaModel{}, fmt.Errorf("model %s not found", modelID)
		}
		return domain.MetaModel{}, err
	}
	m.Status = domain.MetaModelStatus(status)
	return m, nil
}
func (r *MetaRepo) GetModelByCode(ctx context.Context, modelCode string) (domain.MetaModel, error) {
	var m domain.MetaModel
	var status string
	err := r.db.QueryRowContext(ctx, `SELECT id, model_code, model_name, description, status, current_version, created_at, updated_at FROM md_model WHERE model_code=?`, modelCode).Scan(&m.ID, &m.ModelCode, &m.ModelName, &m.Description, &status, &m.CurrentVersion, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaModel{}, fmt.Errorf("model_code %s not found", modelCode)
		}
		return domain.MetaModel{}, err
	}
	m.Status = domain.MetaModelStatus(status)
	return m, nil
}
func (r *MetaRepo) ListModels(ctx context.Context, status string) ([]domain.MetaModel, error) {
	query := `SELECT id, model_code, model_name, description, status, current_version, created_at, updated_at FROM md_model`
	args := []any{}
	if strings.TrimSpace(status) != "" {
		query += ` WHERE status=?`
		args = append(args, status)
	}
	query += ` ORDER BY updated_at DESC`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.MetaModel, 0)
	for rows.Next() {
		var m domain.MetaModel
		var st string
		if err := rows.Scan(&m.ID, &m.ModelCode, &m.ModelName, &m.Description, &st, &m.CurrentVersion, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.Status = domain.MetaModelStatus(st)
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r *MetaRepo) UpdateModel(ctx context.Context, model domain.MetaModel) error {
	res, err := r.db.ExecContext(ctx, `UPDATE md_model SET model_name=?, description=?, status=?, current_version=?, updated_at=? WHERE id=?`, model.ModelName, model.Description, string(model.Status), model.CurrentVersion, model.UpdatedAt, model.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("model %s not found", model.ID)
	}
	return nil
}
func (r *MetaRepo) DeleteModel(ctx context.Context, modelID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM md_model_ref WHERE model_id=?`, modelID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM md_model_field WHERE model_id=?`, modelID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM md_model WHERE id=?`, modelID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("model %s not found", modelID)
	}
	return tx.Commit()
}

func (r *MetaRepo) ListFields(ctx context.Context, modelID string) ([]domain.MetaField, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, model_id, field_code, field_name, category, value_type, required_flag, unique_flag, filterable_flag, sortable_flag, visible_flag, default_value, validation_rule, enum_options_json, sort_no, created_at, updated_at FROM md_model_field WHERE model_id=? ORDER BY sort_no ASC`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.MetaField, 0)
	for rows.Next() {
		var f domain.MetaField
		var req, uniq, fil, sor, vis int
		var enumJSON sql.NullString
		if err := rows.Scan(&f.ID, &f.ModelID, &f.FieldCode, &f.FieldName, &f.Category, &f.ValueType, &req, &uniq, &fil, &sor, &vis, &f.DefaultValue, &f.ValidationRule, &enumJSON, &f.SortNo, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Required = req == 1
		f.Unique = uniq == 1
		f.Filterable = fil == 1
		f.Sortable = sor == 1
		f.Visible = vis == 1
		if enumJSON.Valid && strings.TrimSpace(enumJSON.String) != "" {
			_ = json.Unmarshal([]byte(enumJSON.String), &f.EnumOptions)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
func (r *MetaRepo) CreateField(ctx context.Context, field domain.MetaField) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO md_model_field (id, model_id, field_code, field_name, category, value_type, required_flag, unique_flag, filterable_flag, sortable_flag, visible_flag, default_value, validation_rule, enum_options_json, sort_no, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		field.ID, field.ModelID, field.FieldCode, field.FieldName, field.Category, field.ValueType, boolToInt(field.Required), boolToInt(field.Unique), boolToInt(field.Filterable), boolToInt(field.Sortable), boolToInt(field.Visible), field.DefaultValue, field.ValidationRule, mustJSON(field.EnumOptions), field.SortNo, field.CreatedAt, field.UpdatedAt)
	return err
}
func (r *MetaRepo) GetField(ctx context.Context, modelID, fieldID string) (domain.MetaField, error) {
	var f domain.MetaField
	var req, uniq, fil, sor, vis int
	var enumJSON sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, model_id, field_code, field_name, category, value_type, required_flag, unique_flag, filterable_flag, sortable_flag, visible_flag, default_value, validation_rule, enum_options_json, sort_no, created_at, updated_at FROM md_model_field WHERE model_id=? AND id=?`, modelID, fieldID).Scan(&f.ID, &f.ModelID, &f.FieldCode, &f.FieldName, &f.Category, &f.ValueType, &req, &uniq, &fil, &sor, &vis, &f.DefaultValue, &f.ValidationRule, &enumJSON, &f.SortNo, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaField{}, fmt.Errorf("field %s not found", fieldID)
		}
		return domain.MetaField{}, err
	}
	f.Required = req == 1
	f.Unique = uniq == 1
	f.Filterable = fil == 1
	f.Sortable = sor == 1
	f.Visible = vis == 1
	if enumJSON.Valid && strings.TrimSpace(enumJSON.String) != "" {
		_ = json.Unmarshal([]byte(enumJSON.String), &f.EnumOptions)
	}
	return f, nil
}
func (r *MetaRepo) GetFieldByCode(ctx context.Context, modelID, fieldCode string) (domain.MetaField, error) {
	var f domain.MetaField
	var req, uniq, fil, sor, vis int
	var enumJSON sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, model_id, field_code, field_name, category, value_type, required_flag, unique_flag, filterable_flag, sortable_flag, visible_flag, default_value, validation_rule, enum_options_json, sort_no, created_at, updated_at FROM md_model_field WHERE model_id=? AND field_code=?`, modelID, fieldCode).Scan(&f.ID, &f.ModelID, &f.FieldCode, &f.FieldName, &f.Category, &f.ValueType, &req, &uniq, &fil, &sor, &vis, &f.DefaultValue, &f.ValidationRule, &enumJSON, &f.SortNo, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaField{}, fmt.Errorf("field_code %s not found", fieldCode)
		}
		return domain.MetaField{}, err
	}
	f.Required = req == 1
	f.Unique = uniq == 1
	f.Filterable = fil == 1
	f.Sortable = sor == 1
	f.Visible = vis == 1
	if enumJSON.Valid && strings.TrimSpace(enumJSON.String) != "" {
		_ = json.Unmarshal([]byte(enumJSON.String), &f.EnumOptions)
	}
	return f, nil
}
func (r *MetaRepo) UpdateField(ctx context.Context, field domain.MetaField) error {
	res, err := r.db.ExecContext(ctx, `UPDATE md_model_field SET field_name=?, category=?, value_type=?, required_flag=?, unique_flag=?, filterable_flag=?, sortable_flag=?, visible_flag=?, default_value=?, validation_rule=?, enum_options_json=?, updated_at=? WHERE model_id=? AND id=?`,
		field.FieldName, field.Category, field.ValueType, boolToInt(field.Required), boolToInt(field.Unique), boolToInt(field.Filterable), boolToInt(field.Sortable), boolToInt(field.Visible), field.DefaultValue, field.ValidationRule, mustJSON(field.EnumOptions), field.UpdatedAt, field.ModelID, field.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("field %s not found", field.ID)
	}
	return nil
}
func (r *MetaRepo) DeleteField(ctx context.Context, modelID, fieldID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM md_model_ref WHERE model_id=? AND source_field_id=?`, modelID, fieldID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM md_model_field WHERE model_id=? AND id=?`, modelID, fieldID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("field %s not found", fieldID)
	}
	return tx.Commit()
}
func (r *MetaRepo) ReorderFields(ctx context.Context, modelID string, order []repository.FieldOrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range order {
		if _, err := tx.ExecContext(ctx, `UPDATE md_model_field SET sort_no=?, updated_at=? WHERE model_id=? AND id=?`, item.SortNo, time.Now(), modelID, item.FieldID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MetaRepo) ListReferences(ctx context.Context, modelID string) ([]domain.MetaReference, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, model_id, source_field_id, target_model_id, target_field_id, display_fields_json, on_delete_action, created_at, updated_at FROM md_model_ref WHERE model_id=? ORDER BY created_at ASC`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.MetaReference, 0)
	for rows.Next() {
		var ref domain.MetaReference
		var display string
		if err := rows.Scan(&ref.ID, &ref.ModelID, &ref.SourceFieldID, &ref.TargetModelID, &ref.TargetFieldID, &display, &ref.OnDeleteAction, &ref.CreatedAt, &ref.UpdatedAt); err != nil {
			return nil, err
		}
		if strings.TrimSpace(display) != "" {
			_ = json.Unmarshal([]byte(display), &ref.DisplayFields)
		}
		out = append(out, ref)
	}
	return out, rows.Err()
}
func (r *MetaRepo) CreateReference(ctx context.Context, ref domain.MetaReference) error {
	display, _ := json.Marshal(ref.DisplayFields)
	_, err := r.db.ExecContext(ctx, `INSERT INTO md_model_ref (id, model_id, source_field_id, target_model_id, target_field_id, display_fields_json, on_delete_action, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, ref.ID, ref.ModelID, ref.SourceFieldID, ref.TargetModelID, ref.TargetFieldID, string(display), ref.OnDeleteAction, ref.CreatedAt, ref.UpdatedAt)
	return err
}
func (r *MetaRepo) GetReference(ctx context.Context, modelID, refID string) (domain.MetaReference, error) {
	var ref domain.MetaReference
	var display string
	err := r.db.QueryRowContext(ctx, `SELECT id, model_id, source_field_id, target_model_id, target_field_id, display_fields_json, on_delete_action, created_at, updated_at FROM md_model_ref WHERE model_id=? AND id=?`, modelID, refID).Scan(&ref.ID, &ref.ModelID, &ref.SourceFieldID, &ref.TargetModelID, &ref.TargetFieldID, &display, &ref.OnDeleteAction, &ref.CreatedAt, &ref.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaReference{}, fmt.Errorf("reference %s not found", refID)
		}
		return domain.MetaReference{}, err
	}
	if strings.TrimSpace(display) != "" {
		_ = json.Unmarshal([]byte(display), &ref.DisplayFields)
	}
	return ref, nil
}
func (r *MetaRepo) UpdateReference(ctx context.Context, ref domain.MetaReference) error {
	display, _ := json.Marshal(ref.DisplayFields)
	res, err := r.db.ExecContext(ctx, `UPDATE md_model_ref SET source_field_id=?, target_model_id=?, target_field_id=?, display_fields_json=?, on_delete_action=?, updated_at=? WHERE model_id=? AND id=?`, ref.SourceFieldID, ref.TargetModelID, ref.TargetFieldID, string(display), ref.OnDeleteAction, ref.UpdatedAt, ref.ModelID, ref.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("reference %s not found", ref.ID)
	}
	return nil
}
func (r *MetaRepo) DeleteReference(ctx context.Context, modelID, refID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM md_model_ref WHERE model_id=? AND id=?`, modelID, refID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("reference %s not found", refID)
	}
	return nil
}
func (r *MetaRepo) CountRecords(ctx context.Context, modelID string) (int64, error) {
	var cnt int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM md_record WHERE model_id=? AND deleted_flag=0`, modelID).Scan(&cnt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return 0, nil
		}
		return 0, err
	}
	return cnt, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (r *MetaRepo) CreateVersion(ctx context.Context, version domain.MetaModelVersion) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO md_model_version (id, model_id, version_no, snapshot_json, published_at, published_by, change_summary) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		version.ID, version.ModelID, version.VersionNo, version.SnapshotJSON, version.PublishedAt, version.PublishedBy, version.ChangeSummary)
	return err
}

func (r *MetaRepo) ListVersions(ctx context.Context, modelID string) ([]domain.MetaModelVersion, error) {
	if _, err := r.GetModel(ctx, modelID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, model_id, version_no, snapshot_json, published_at, published_by, change_summary FROM md_model_version WHERE model_id=? ORDER BY version_no DESC`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.MetaModelVersion, 0)
	for rows.Next() {
		var v domain.MetaModelVersion
		if err := rows.Scan(&v.ID, &v.ModelID, &v.VersionNo, &v.SnapshotJSON, &v.PublishedAt, &v.PublishedBy, &v.ChangeSummary); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *MetaRepo) GetVersion(ctx context.Context, modelID string, versionNo int) (domain.MetaModelVersion, error) {
	var v domain.MetaModelVersion
	err := r.db.QueryRowContext(ctx, `SELECT id, model_id, version_no, snapshot_json, published_at, published_by, change_summary FROM md_model_version WHERE model_id=? AND version_no=?`, modelID, versionNo).Scan(&v.ID, &v.ModelID, &v.VersionNo, &v.SnapshotJSON, &v.PublishedAt, &v.PublishedBy, &v.ChangeSummary)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaModelVersion{}, fmt.Errorf("version %d not found", versionNo)
		}
		return domain.MetaModelVersion{}, err
	}
	return v, nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
