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

func (r *MetaRepo) ListRecords(ctx context.Context, modelID string) ([]domain.MetaRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, model_id, data_json, created_at, updated_at FROM md_record WHERE model_id=? AND deleted_flag=0 ORDER BY updated_at DESC`, modelID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return []domain.MetaRecord{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.MetaRecord, 0)
	for rows.Next() {
		var rec domain.MetaRecord
		var dataJSON string
		if err := rows.Scan(&rec.ID, &rec.ModelID, &dataJSON, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		rec.Data = map[string]any{}
		if strings.TrimSpace(dataJSON) != "" {
			_ = json.Unmarshal([]byte(dataJSON), &rec.Data)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *MetaRepo) GetRecord(ctx context.Context, modelID, recordID string) (domain.MetaRecord, error) {
	var rec domain.MetaRecord
	var dataJSON string
	err := r.db.QueryRowContext(ctx, `SELECT id, model_id, data_json, created_at, updated_at FROM md_record WHERE model_id=? AND id=? AND deleted_flag=0`, modelID, recordID).Scan(&rec.ID, &rec.ModelID, &dataJSON, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaRecord{}, fmt.Errorf("record %s not found", recordID)
		}
		return domain.MetaRecord{}, err
	}
	rec.Data = map[string]any{}
	if strings.TrimSpace(dataJSON) != "" {
		_ = json.Unmarshal([]byte(dataJSON), &rec.Data)
	}
	return rec, nil
}

func (r *MetaRepo) CreateRecord(ctx context.Context, record domain.MetaRecord) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO md_record (id, model_id, data_json, deleted_flag, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?)`, record.ID, record.ModelID, mustJSON(record.Data), record.CreatedAt, record.UpdatedAt)
	return err
}

func (r *MetaRepo) CreateRecordsBatch(ctx context.Context, records []domain.MetaRecord) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO md_record (id, model_id, data_json, deleted_flag, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, rec := range records {
		if _, err := stmt.ExecContext(ctx, rec.ID, rec.ModelID, mustJSON(rec.Data), rec.CreatedAt, rec.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MetaRepo) UpdateRecord(ctx context.Context, record domain.MetaRecord) error {
	res, err := r.db.ExecContext(ctx, `UPDATE md_record SET data_json=?, updated_at=? WHERE model_id=? AND id=? AND deleted_flag=0`, mustJSON(record.Data), record.UpdatedAt, record.ModelID, record.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("record %s not found", record.ID)
	}
	return nil
}

func (r *MetaRepo) DeleteRecord(ctx context.Context, modelID, recordID string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE md_record SET deleted_flag=1, updated_at=? WHERE model_id=? AND id=? AND deleted_flag=0`, time.Now(), modelID, recordID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("record %s not found", recordID)
	}
	return nil
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

func (r *MetaRepo) CreateImportJob(ctx context.Context, job domain.MetaImportJob) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO md_import_job (job_id, model_id, status, total, processed, success, failed, errors_json, started_at, finished_at, message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.JobID, job.ModelID, job.Status, job.Total, job.Processed, job.Success, job.Failed, mustJSON(job.Errors), job.StartedAt, job.FinishedAt, job.Message, time.Now(), time.Now())
	return err
}

func (r *MetaRepo) UpdateImportJob(ctx context.Context, job domain.MetaImportJob) error {
	res, err := r.db.ExecContext(ctx, `UPDATE md_import_job SET status=?, total=?, processed=?, success=?, failed=?, errors_json=?, started_at=?, finished_at=?, message=?, updated_at=? WHERE job_id=?`,
		job.Status, job.Total, job.Processed, job.Success, job.Failed, mustJSON(job.Errors), job.StartedAt, job.FinishedAt, job.Message, time.Now(), job.JobID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("import job %s not found", job.JobID)
	}
	return nil
}

func (r *MetaRepo) GetImportJob(ctx context.Context, jobID string) (domain.MetaImportJob, error) {
	var job domain.MetaImportJob
	var errorsJSON sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT job_id, model_id, status, total, processed, success, failed, errors_json, started_at, finished_at, message FROM md_import_job WHERE job_id=?`, jobID).
		Scan(&job.JobID, &job.ModelID, &job.Status, &job.Total, &job.Processed, &job.Success, &job.Failed, &errorsJSON, &job.StartedAt, &job.FinishedAt, &job.Message)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.MetaImportJob{}, fmt.Errorf("import job %s not found", jobID)
		}
		return domain.MetaImportJob{}, err
	}
	job.Errors = make([]map[string]any, 0)
	if errorsJSON.Valid && strings.TrimSpace(errorsJSON.String) != "" {
		_ = json.Unmarshal([]byte(errorsJSON.String), &job.Errors)
	}
	return job, nil
}

func (r *MetaRepo) CleanupImportJobs(ctx context.Context, finishedBefore time.Time, keepLatest int) (int64, error) {
	if keepLatest < 0 {
		keepLatest = 0
	}
	rows, err := r.db.QueryContext(ctx, `SELECT job_id FROM md_import_job WHERE status IN ('done','failed') AND finished_at IS NOT NULL AND finished_at < ? ORDER BY finished_at DESC`, finishedBefore)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(ids) <= keepLatest {
		return 0, nil
	}
	toClean := ids[keepLatest:]
	cleaned := int64(0)
	for _, id := range toClean {
		res, err := r.db.ExecContext(ctx, `UPDATE md_import_job SET status='cleaned', errors_json='[]', message='导入任务明细已自动清理', updated_at=? WHERE job_id=?`, time.Now(), id)
		if err != nil {
			return cleaned, err
		}
		n, _ := res.RowsAffected()
		cleaned += n
	}
	return cleaned, nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
