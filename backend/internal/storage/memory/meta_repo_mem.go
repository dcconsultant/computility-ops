package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

type MetaRepo struct {
	mu          sync.RWMutex
	models      map[string]domain.MetaModel
	modelByCode map[string]string
	fields      map[string]map[string]domain.MetaField
	fieldByCode map[string]map[string]string
	references  map[string]map[string]domain.MetaReference
	versions    map[string]map[int]domain.MetaModelVersion
	records     map[string]map[string]domain.MetaRecord
	importJobs  map[string]domain.MetaImportJob
}

func NewMetaRepo() *MetaRepo {
	return &MetaRepo{
		models:      map[string]domain.MetaModel{},
		modelByCode: map[string]string{},
		fields:      map[string]map[string]domain.MetaField{},
		fieldByCode: map[string]map[string]string{},
		references:  map[string]map[string]domain.MetaReference{},
		versions:    map[string]map[int]domain.MetaModelVersion{},
		records:     map[string]map[string]domain.MetaRecord{},
		importJobs:  map[string]domain.MetaImportJob{},
	}
}

func (r *MetaRepo) CreateModel(_ context.Context, model domain.MetaModel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.modelByCode[model.ModelCode]; ok {
		return fmt.Errorf("model_code already exists")
	}
	r.models[model.ID] = model
	r.modelByCode[model.ModelCode] = model.ID
	return nil
}
func (r *MetaRepo) GetModel(_ context.Context, modelID string) (domain.MetaModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.models[modelID]
	if !ok {
		return domain.MetaModel{}, fmt.Errorf("model %s not found", modelID)
	}
	return m, nil
}
func (r *MetaRepo) GetModelByCode(_ context.Context, modelCode string) (domain.MetaModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.modelByCode[modelCode]
	if !ok {
		return domain.MetaModel{}, fmt.Errorf("model_code %s not found", modelCode)
	}
	return r.models[id], nil
}
func (r *MetaRepo) ListModels(_ context.Context, status string) ([]domain.MetaModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.MetaModel, 0, len(r.models))
	for _, m := range r.models {
		if status != "" && string(m.Status) != status {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (r *MetaRepo) UpdateModel(_ context.Context, model domain.MetaModel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.models[model.ID]
	if !ok {
		return fmt.Errorf("model %s not found", model.ID)
	}
	if old.ModelCode != model.ModelCode {
		if _, exists := r.modelByCode[model.ModelCode]; exists {
			return fmt.Errorf("model_code already exists")
		}
		delete(r.modelByCode, old.ModelCode)
		r.modelByCode[model.ModelCode] = model.ID
	}
	r.models[model.ID] = model
	return nil
}
func (r *MetaRepo) DeleteModel(_ context.Context, modelID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.models[modelID]
	if !ok {
		return fmt.Errorf("model %s not found", modelID)
	}
	delete(r.models, modelID)
	delete(r.modelByCode, m.ModelCode)
	delete(r.fields, modelID)
	delete(r.fieldByCode, modelID)
	delete(r.references, modelID)
	delete(r.records, modelID)
	return nil
}

func (r *MetaRepo) ListFields(_ context.Context, modelID string) ([]domain.MetaField, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fm := r.fields[modelID]
	out := make([]domain.MetaField, 0, len(fm))
	for _, f := range fm {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortNo < out[j].SortNo })
	return out, nil
}
func (r *MetaRepo) CreateField(_ context.Context, field domain.MetaField) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.models[field.ModelID]; !ok {
		return fmt.Errorf("model %s not found", field.ModelID)
	}
	if r.fields[field.ModelID] == nil {
		r.fields[field.ModelID] = map[string]domain.MetaField{}
	}
	if r.fieldByCode[field.ModelID] == nil {
		r.fieldByCode[field.ModelID] = map[string]string{}
	}
	if _, exists := r.fieldByCode[field.ModelID][field.FieldCode]; exists {
		return fmt.Errorf("field_code already exists in model")
	}
	r.fields[field.ModelID][field.ID] = field
	r.fieldByCode[field.ModelID][field.FieldCode] = field.ID
	return nil
}
func (r *MetaRepo) GetField(_ context.Context, modelID, fieldID string) (domain.MetaField, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fm := r.fields[modelID]
	f, ok := fm[fieldID]
	if !ok {
		return domain.MetaField{}, fmt.Errorf("field %s not found", fieldID)
	}
	return f, nil
}
func (r *MetaRepo) GetFieldByCode(_ context.Context, modelID, fieldCode string) (domain.MetaField, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.fieldByCode[modelID][fieldCode]
	if !ok {
		return domain.MetaField{}, fmt.Errorf("field_code %s not found", fieldCode)
	}
	return r.fields[modelID][id], nil
}
func (r *MetaRepo) UpdateField(_ context.Context, field domain.MetaField) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	fm := r.fields[field.ModelID]
	old, ok := fm[field.ID]
	if !ok {
		return fmt.Errorf("field %s not found", field.ID)
	}
	if old.FieldCode != field.FieldCode {
		return fmt.Errorf("field_code cannot be changed")
	}
	fm[field.ID] = field
	return nil
}
func (r *MetaRepo) DeleteField(_ context.Context, modelID, fieldID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	fm := r.fields[modelID]
	f, ok := fm[fieldID]
	if !ok {
		return fmt.Errorf("field %s not found", fieldID)
	}
	delete(fm, fieldID)
	delete(r.fieldByCode[modelID], f.FieldCode)
	for rid, ref := range r.references[modelID] {
		if ref.SourceFieldID == fieldID {
			delete(r.references[modelID], rid)
		}
	}
	return nil
}
func (r *MetaRepo) ReorderFields(_ context.Context, modelID string, order []repository.FieldOrderItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	fm := r.fields[modelID]
	for _, item := range order {
		f, ok := fm[item.FieldID]
		if !ok {
			return fmt.Errorf("field %s not found", item.FieldID)
		}
		f.SortNo = item.SortNo
		fm[item.FieldID] = f
	}
	return nil
}

func (r *MetaRepo) ListReferences(_ context.Context, modelID string) ([]domain.MetaReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rm := r.references[modelID]
	out := make([]domain.MetaReference, 0, len(rm))
	for _, ref := range rm {
		out = append(out, ref)
	}
	sort.Slice(out, func(i, j int) bool { return strings.Compare(out[i].CreatedAt.String(), out[j].CreatedAt.String()) < 0 })
	return out, nil
}
func (r *MetaRepo) CreateReference(_ context.Context, ref domain.MetaReference) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.models[ref.ModelID]; !ok {
		return fmt.Errorf("model %s not found", ref.ModelID)
	}
	if r.references[ref.ModelID] == nil {
		r.references[ref.ModelID] = map[string]domain.MetaReference{}
	}
	r.references[ref.ModelID][ref.ID] = ref
	return nil
}
func (r *MetaRepo) GetReference(_ context.Context, modelID, refID string) (domain.MetaReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ref, ok := r.references[modelID][refID]
	if !ok {
		return domain.MetaReference{}, fmt.Errorf("reference %s not found", refID)
	}
	return ref, nil
}
func (r *MetaRepo) UpdateReference(_ context.Context, ref domain.MetaReference) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.references[ref.ModelID][ref.ID]; !ok {
		return fmt.Errorf("reference %s not found", ref.ID)
	}
	r.references[ref.ModelID][ref.ID] = ref
	return nil
}
func (r *MetaRepo) DeleteReference(_ context.Context, modelID, refID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.references[modelID][refID]; !ok {
		return fmt.Errorf("reference %s not found", refID)
	}
	delete(r.references[modelID], refID)
	return nil
}
func (r *MetaRepo) CountRecords(_ context.Context, modelID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.records[modelID])), nil
}

func (r *MetaRepo) ListRecords(_ context.Context, modelID string) ([]domain.MetaRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rm := r.records[modelID]
	out := make([]domain.MetaRecord, 0, len(rm))
	for _, rec := range rm {
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (r *MetaRepo) GetRecord(_ context.Context, modelID, recordID string) (domain.MetaRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.records[modelID][recordID]
	if !ok {
		return domain.MetaRecord{}, fmt.Errorf("record %s not found", recordID)
	}
	return rec, nil
}

func (r *MetaRepo) CreateRecord(_ context.Context, record domain.MetaRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.models[record.ModelID]; !ok {
		return fmt.Errorf("model %s not found", record.ModelID)
	}
	if r.records[record.ModelID] == nil {
		r.records[record.ModelID] = map[string]domain.MetaRecord{}
	}
	r.records[record.ModelID][record.ID] = record
	return nil
}

func (r *MetaRepo) CreateRecordsBatch(_ context.Context, records []domain.MetaRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, record := range records {
		if _, ok := r.models[record.ModelID]; !ok {
			return fmt.Errorf("model %s not found", record.ModelID)
		}
		if r.records[record.ModelID] == nil {
			r.records[record.ModelID] = map[string]domain.MetaRecord{}
		}
		r.records[record.ModelID][record.ID] = record
	}
	return nil
}

func (r *MetaRepo) UpdateRecord(_ context.Context, record domain.MetaRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.records[record.ModelID][record.ID]; !ok {
		return fmt.Errorf("record %s not found", record.ID)
	}
	r.records[record.ModelID][record.ID] = record
	return nil
}

func (r *MetaRepo) DeleteRecord(_ context.Context, modelID, recordID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.records[modelID][recordID]; !ok {
		return fmt.Errorf("record %s not found", recordID)
	}
	delete(r.records[modelID], recordID)
	return nil
}

func (r *MetaRepo) DeleteRecordsByModel(_ context.Context, modelID string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.models[modelID]; !ok {
		return 0, fmt.Errorf("model %s not found", modelID)
	}
	cnt := int64(len(r.records[modelID]))
	r.records[modelID] = map[string]domain.MetaRecord{}
	return cnt, nil
}

func (r *MetaRepo) CreateVersion(_ context.Context, version domain.MetaModelVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.models[version.ModelID]; !ok {
		return fmt.Errorf("model %s not found", version.ModelID)
	}
	if r.versions[version.ModelID] == nil {
		r.versions[version.ModelID] = map[int]domain.MetaModelVersion{}
	}
	if _, exists := r.versions[version.ModelID][version.VersionNo]; exists {
		return fmt.Errorf("version %d already exists", version.VersionNo)
	}
	r.versions[version.ModelID][version.VersionNo] = version
	return nil
}

func (r *MetaRepo) ListVersions(_ context.Context, modelID string) ([]domain.MetaModelVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.models[modelID]; !ok {
		return nil, fmt.Errorf("model %s not found", modelID)
	}
	vm := r.versions[modelID]
	out := make([]domain.MetaModelVersion, 0, len(vm))
	for _, v := range vm {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].VersionNo > out[j].VersionNo })
	return out, nil
}

func (r *MetaRepo) GetVersion(_ context.Context, modelID string, versionNo int) (domain.MetaModelVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.models[modelID]; !ok {
		return domain.MetaModelVersion{}, fmt.Errorf("model %s not found", modelID)
	}
	v, ok := r.versions[modelID][versionNo]
	if !ok {
		return domain.MetaModelVersion{}, fmt.Errorf("version %d not found", versionNo)
	}
	return v, nil
}

func (r *MetaRepo) CreateImportJob(_ context.Context, job domain.MetaImportJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.importJobs[job.JobID] = job
	return nil
}

func (r *MetaRepo) UpdateImportJob(_ context.Context, job domain.MetaImportJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.importJobs[job.JobID]; !ok {
		return fmt.Errorf("import job %s not found", job.JobID)
	}
	r.importJobs[job.JobID] = job
	return nil
}

func (r *MetaRepo) GetImportJob(_ context.Context, jobID string) (domain.MetaImportJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.importJobs[jobID]
	if !ok {
		return domain.MetaImportJob{}, fmt.Errorf("import job %s not found", jobID)
	}
	return job, nil
}

func (r *MetaRepo) CleanupImportJobs(_ context.Context, finishedBefore time.Time, keepLatest int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	type kv struct {
		id  string
		job domain.MetaImportJob
	}
	finished := make([]kv, 0)
	for id, job := range r.importJobs {
		if (job.Status == "done" || job.Status == "failed") && job.FinishedAt != nil && job.FinishedAt.Before(finishedBefore) {
			finished = append(finished, kv{id: id, job: job})
		}
	}
	sort.Slice(finished, func(i, j int) bool {
		return finished[i].job.FinishedAt.After(*finished[j].job.FinishedAt)
	})
	var cleaned int64
	for i, item := range finished {
		if i < keepLatest {
			continue
		}
		job := item.job
		job.Status = "cleaned"
		job.Errors = []map[string]any{}
		job.Message = "导入任务明细已自动清理"
		r.importJobs[item.id] = job
		cleaned++
	}
	return cleaned, nil
}
