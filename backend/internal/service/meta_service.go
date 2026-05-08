package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
)

var modelCodeRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

type MetaService struct {
	repo repository.MetaRepo
}

func NewMetaService(repo repository.MetaRepo) *MetaService { return &MetaService{repo: repo} }

type CreateModelInput struct {
	ModelCode   string
	ModelName   string
	Description string
}

type UpdateModelInput struct {
	ModelName   string
	Description string
}

func (s *MetaService) CreateModel(ctx context.Context, in CreateModelInput) (domain.MetaModel, error) {
	code := strings.TrimSpace(in.ModelCode)
	name := strings.TrimSpace(in.ModelName)
	if !modelCodeRegexp.MatchString(code) {
		return domain.MetaModel{}, fmt.Errorf("model_code must match ^[A-Za-z][A-Za-z0-9_]*$")
	}
	if name == "" {
		return domain.MetaModel{}, fmt.Errorf("model_name is required")
	}
	if _, err := s.repo.GetModelByCode(ctx, code); err == nil {
		return domain.MetaModel{}, fmt.Errorf("model_code already exists")
	}
	now := time.Now()
	m := domain.MetaModel{ID: fmt.Sprintf("%d", now.UnixNano()), ModelCode: code, ModelName: name, Description: strings.TrimSpace(in.Description), Status: domain.MetaModelStatusDraft, CurrentVersion: 0, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateModel(ctx, m); err != nil {
		return domain.MetaModel{}, err
	}
	return m, nil
}

func (s *MetaService) ListModels(ctx context.Context, status string) ([]domain.MetaModel, error) {
	models, err := s.repo.ListModels(ctx, strings.TrimSpace(status))
	if err != nil {
		return nil, err
	}
	sort.Slice(models, func(i, j int) bool { return models[i].UpdatedAt.After(models[j].UpdatedAt) })
	return models, nil
}

func (s *MetaService) GetModel(ctx context.Context, modelID string) (domain.MetaModel, []domain.MetaField, error) {
	m, err := s.repo.GetModel(ctx, strings.TrimSpace(modelID))
	if err != nil {
		return domain.MetaModel{}, nil, err
	}
	fields, err := s.repo.ListFields(ctx, m.ID)
	if err != nil {
		return domain.MetaModel{}, nil, err
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].SortNo < fields[j].SortNo })
	return m, fields, nil
}

func (s *MetaService) UpdateModel(ctx context.Context, modelID string, in UpdateModelInput) (domain.MetaModel, error) {
	m, err := s.repo.GetModel(ctx, strings.TrimSpace(modelID))
	if err != nil {
		return domain.MetaModel{}, err
	}
	name := strings.TrimSpace(in.ModelName)
	if name == "" {
		return domain.MetaModel{}, fmt.Errorf("model_name is required")
	}
	m.ModelName = name
	m.Description = strings.TrimSpace(in.Description)
	m.UpdatedAt = time.Now()
	if err := s.repo.UpdateModel(ctx, m); err != nil {
		return domain.MetaModel{}, err
	}
	return m, nil
}

func (s *MetaService) ArchiveModel(ctx context.Context, modelID string) (domain.MetaModel, error) {
	m, err := s.repo.GetModel(ctx, strings.TrimSpace(modelID))
	if err != nil {
		return domain.MetaModel{}, err
	}
	m.Status = domain.MetaModelStatusArchived
	m.UpdatedAt = time.Now()
	if err := s.repo.UpdateModel(ctx, m); err != nil {
		return domain.MetaModel{}, err
	}
	return m, nil
}

func (s *MetaService) DeleteModel(ctx context.Context, modelID string) error {
	m, err := s.repo.GetModel(ctx, strings.TrimSpace(modelID))
	if err != nil {
		return err
	}
	if m.Status != domain.MetaModelStatusDraft {
		return fmt.Errorf("only draft model can be deleted")
	}
	cnt, err := s.repo.CountRecords(ctx, m.ID)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf("model has records and cannot be deleted")
	}
	return s.repo.DeleteModel(ctx, m.ID)
}

type CreateFieldInput struct {
	FieldCode      string
	FieldName      string
	Category       string
	ValueType      string
	Required       bool
	Unique         bool
	Filterable     bool
	Sortable       bool
	Visible        bool
	DefaultValue   string
	ValidationRule string
}

type UpdateFieldInput struct {
	FieldName      string
	Category       string
	ValueType      string
	Required       bool
	Unique         bool
	Filterable     bool
	Sortable       bool
	Visible        bool
	DefaultValue   string
	ValidationRule string
}

func (s *MetaService) CreateField(ctx context.Context, modelID string, in CreateFieldInput) (domain.MetaField, error) {
	modelID = strings.TrimSpace(modelID)
	if _, err := s.repo.GetModel(ctx, modelID); err != nil {
		return domain.MetaField{}, err
	}
	code := strings.TrimSpace(in.FieldCode)
	name := strings.TrimSpace(in.FieldName)
	if !modelCodeRegexp.MatchString(code) {
		return domain.MetaField{}, fmt.Errorf("field_code must match ^[A-Za-z][A-Za-z0-9_]*$")
	}
	if name == "" {
		return domain.MetaField{}, fmt.Errorf("field_name is required")
	}
	if strings.TrimSpace(in.ValueType) == "" {
		return domain.MetaField{}, fmt.Errorf("value_type is required")
	}
	if _, err := s.repo.GetFieldByCode(ctx, modelID, code); err == nil {
		return domain.MetaField{}, fmt.Errorf("field_code already exists in model")
	}
	fields, _ := s.repo.ListFields(ctx, modelID)
	now := time.Now()
	field := domain.MetaField{ID: fmt.Sprintf("%d", now.UnixNano()), ModelID: modelID, FieldCode: code, FieldName: name, Category: strings.TrimSpace(in.Category), ValueType: strings.TrimSpace(in.ValueType), Required: in.Required, Unique: in.Unique, Filterable: in.Filterable, Sortable: in.Sortable, Visible: in.Visible, DefaultValue: strings.TrimSpace(in.DefaultValue), ValidationRule: strings.TrimSpace(in.ValidationRule), SortNo: len(fields) + 1, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateField(ctx, field); err != nil {
		return domain.MetaField{}, err
	}
	return field, nil
}

func (s *MetaService) UpdateField(ctx context.Context, modelID, fieldID string, in UpdateFieldInput) (domain.MetaField, error) {
	f, err := s.repo.GetField(ctx, strings.TrimSpace(modelID), strings.TrimSpace(fieldID))
	if err != nil {
		return domain.MetaField{}, err
	}
	name := strings.TrimSpace(in.FieldName)
	if name == "" {
		return domain.MetaField{}, fmt.Errorf("field_name is required")
	}
	if strings.TrimSpace(in.ValueType) == "" {
		return domain.MetaField{}, fmt.Errorf("value_type is required")
	}
	f.FieldName = name
	f.Category = strings.TrimSpace(in.Category)
	f.ValueType = strings.TrimSpace(in.ValueType)
	f.Required = in.Required
	f.Unique = in.Unique
	f.Filterable = in.Filterable
	f.Sortable = in.Sortable
	f.Visible = in.Visible
	f.DefaultValue = strings.TrimSpace(in.DefaultValue)
	f.ValidationRule = strings.TrimSpace(in.ValidationRule)
	f.UpdatedAt = time.Now()
	if err := s.repo.UpdateField(ctx, f); err != nil {
		return domain.MetaField{}, err
	}
	return f, nil
}

func (s *MetaService) DeleteField(ctx context.Context, modelID, fieldID string) error {
	return s.repo.DeleteField(ctx, strings.TrimSpace(modelID), strings.TrimSpace(fieldID))
}

func (s *MetaService) ReorderFields(ctx context.Context, modelID string, fieldIDs []string) ([]domain.MetaField, error) {
	modelID = strings.TrimSpace(modelID)
	fields, err := s.repo.ListFields(ctx, modelID)
	if err != nil {
		return nil, err
	}
	if len(fieldIDs) != len(fields) {
		return nil, fmt.Errorf("reorder list size mismatch")
	}
	seen := map[string]struct{}{}
	items := make([]repository.FieldOrderItem, 0, len(fieldIDs))
	for i, id := range fieldIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("field_id is required")
		}
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("duplicate field_id in reorder list")
		}
		seen[id] = struct{}{}
		items = append(items, repository.FieldOrderItem{FieldID: id, SortNo: i + 1})
	}
	for _, f := range fields {
		if _, ok := seen[f.ID]; !ok {
			return nil, fmt.Errorf("reorder list missing field_id %s", f.ID)
		}
	}
	if err := s.repo.ReorderFields(ctx, modelID, items); err != nil {
		return nil, err
	}
	updated, err := s.repo.ListFields(ctx, modelID)
	if err != nil {
		return nil, err
	}
	sort.Slice(updated, func(i, j int) bool { return updated[i].SortNo < updated[j].SortNo })
	return updated, nil
}

type CreateReferenceInput struct {
	SourceFieldID  string
	TargetModelID  string
	TargetFieldID  string
	DisplayFields  []string
	OnDeleteAction string
}

type UpdateReferenceInput = CreateReferenceInput

func (s *MetaService) CreateReference(ctx context.Context, modelID string, in CreateReferenceInput) (domain.MetaReference, error) {
	modelID = strings.TrimSpace(modelID)
	if _, err := s.repo.GetModel(ctx, modelID); err != nil {
		return domain.MetaReference{}, err
	}
	if _, err := s.repo.GetField(ctx, modelID, strings.TrimSpace(in.SourceFieldID)); err != nil {
		return domain.MetaReference{}, fmt.Errorf("source_field_id invalid: %w", err)
	}
	if _, err := s.repo.GetModel(ctx, strings.TrimSpace(in.TargetModelID)); err != nil {
		return domain.MetaReference{}, fmt.Errorf("target_model_id invalid: %w", err)
	}
	if _, err := s.repo.GetField(ctx, strings.TrimSpace(in.TargetModelID), strings.TrimSpace(in.TargetFieldID)); err != nil {
		return domain.MetaReference{}, fmt.Errorf("target_field_id invalid: %w", err)
	}
	now := time.Now()
	action := strings.TrimSpace(in.OnDeleteAction)
	if action == "" {
		action = "restrict"
	}
	ref := domain.MetaReference{ID: fmt.Sprintf("%d", now.UnixNano()), ModelID: modelID, SourceFieldID: strings.TrimSpace(in.SourceFieldID), TargetModelID: strings.TrimSpace(in.TargetModelID), TargetFieldID: strings.TrimSpace(in.TargetFieldID), DisplayFields: sanitizeStringSlice(in.DisplayFields), OnDeleteAction: action, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateReference(ctx, ref); err != nil {
		return domain.MetaReference{}, err
	}
	return ref, nil
}

func (s *MetaService) UpdateReference(ctx context.Context, modelID, refID string, in UpdateReferenceInput) (domain.MetaReference, error) {
	ref, err := s.repo.GetReference(ctx, strings.TrimSpace(modelID), strings.TrimSpace(refID))
	if err != nil {
		return domain.MetaReference{}, err
	}
	if _, err := s.repo.GetField(ctx, strings.TrimSpace(modelID), strings.TrimSpace(in.SourceFieldID)); err != nil {
		return domain.MetaReference{}, fmt.Errorf("source_field_id invalid: %w", err)
	}
	if _, err := s.repo.GetModel(ctx, strings.TrimSpace(in.TargetModelID)); err != nil {
		return domain.MetaReference{}, fmt.Errorf("target_model_id invalid: %w", err)
	}
	if _, err := s.repo.GetField(ctx, strings.TrimSpace(in.TargetModelID), strings.TrimSpace(in.TargetFieldID)); err != nil {
		return domain.MetaReference{}, fmt.Errorf("target_field_id invalid: %w", err)
	}
	action := strings.TrimSpace(in.OnDeleteAction)
	if action == "" {
		action = "restrict"
	}
	ref.SourceFieldID = strings.TrimSpace(in.SourceFieldID)
	ref.TargetModelID = strings.TrimSpace(in.TargetModelID)
	ref.TargetFieldID = strings.TrimSpace(in.TargetFieldID)
	ref.DisplayFields = sanitizeStringSlice(in.DisplayFields)
	ref.OnDeleteAction = action
	ref.UpdatedAt = time.Now()
	if err := s.repo.UpdateReference(ctx, ref); err != nil {
		return domain.MetaReference{}, err
	}
	return ref, nil
}

func (s *MetaService) DeleteReference(ctx context.Context, modelID, refID string) error {
	return s.repo.DeleteReference(ctx, strings.TrimSpace(modelID), strings.TrimSpace(refID))
}

func (s *MetaService) ListReferences(ctx context.Context, modelID string) ([]domain.MetaReference, error) {
	return s.repo.ListReferences(ctx, strings.TrimSpace(modelID))
}

func sanitizeStringSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

type PublishModelInput struct {
	ChangeSummary string
	PublishedBy   string
}

func (s *MetaService) PublishModel(ctx context.Context, modelID string, in PublishModelInput) (domain.MetaModelVersion, error) {
	modelID = strings.TrimSpace(modelID)
	m, err := s.repo.GetModel(ctx, modelID)
	if err != nil {
		return domain.MetaModelVersion{}, err
	}
	fields, err := s.repo.ListFields(ctx, modelID)
	if err != nil {
		return domain.MetaModelVersion{}, err
	}
	if len(fields) == 0 {
		return domain.MetaModelVersion{}, fmt.Errorf("cannot publish model without fields")
	}
	if err := s.validateNoCycle(ctx, modelID); err != nil {
		return domain.MetaModelVersion{}, err
	}
	refs, err := s.repo.ListReferences(ctx, modelID)
	if err != nil {
		return domain.MetaModelVersion{}, err
	}
	snapshot := domain.MetaModelSnapshot{Model: m, Fields: fields, References: refs}
	b, err := json.Marshal(snapshot)
	if err != nil {
		return domain.MetaModelVersion{}, err
	}
	next := m.CurrentVersion + 1
	now := time.Now()
	v := domain.MetaModelVersion{
		ID:            fmt.Sprintf("%d", now.UnixNano()),
		ModelID:       modelID,
		VersionNo:     next,
		SnapshotJSON:  string(b),
		PublishedAt:   now,
		PublishedBy:   strings.TrimSpace(in.PublishedBy),
		ChangeSummary: strings.TrimSpace(in.ChangeSummary),
	}
	if err := s.repo.CreateVersion(ctx, v); err != nil {
		return domain.MetaModelVersion{}, err
	}
	m.CurrentVersion = next
	m.Status = domain.MetaModelStatusPublished
	m.UpdatedAt = now
	if err := s.repo.UpdateModel(ctx, m); err != nil {
		return domain.MetaModelVersion{}, err
	}
	return v, nil
}

func (s *MetaService) ListVersions(ctx context.Context, modelID string) ([]domain.MetaModelVersion, error) {
	return s.repo.ListVersions(ctx, strings.TrimSpace(modelID))
}

func (s *MetaService) GetVersion(ctx context.Context, modelID string, versionNo int) (domain.MetaModelVersion, domain.MetaModelSnapshot, error) {
	v, err := s.repo.GetVersion(ctx, strings.TrimSpace(modelID), versionNo)
	if err != nil {
		return domain.MetaModelVersion{}, domain.MetaModelSnapshot{}, err
	}
	var snap domain.MetaModelSnapshot
	if err := json.Unmarshal([]byte(v.SnapshotJSON), &snap); err != nil {
		return domain.MetaModelVersion{}, domain.MetaModelSnapshot{}, err
	}
	return v, snap, nil
}

func (s *MetaService) RollbackModel(ctx context.Context, modelID string, versionNo int) (domain.MetaModel, error) {
	modelID = strings.TrimSpace(modelID)
	m, err := s.repo.GetModel(ctx, modelID)
	if err != nil {
		return domain.MetaModel{}, err
	}
	v, err := s.repo.GetVersion(ctx, modelID, versionNo)
	if err != nil {
		return domain.MetaModel{}, err
	}
	var snap domain.MetaModelSnapshot
	if err := json.Unmarshal([]byte(v.SnapshotJSON), &snap); err != nil {
		return domain.MetaModel{}, err
	}
	currentFields, err := s.repo.ListFields(ctx, modelID)
	if err != nil {
		return domain.MetaModel{}, err
	}
	for _, f := range currentFields {
		if err := s.repo.DeleteField(ctx, modelID, f.ID); err != nil {
			return domain.MetaModel{}, err
		}
	}
	currentRefs, err := s.repo.ListReferences(ctx, modelID)
	if err == nil {
		for _, r := range currentRefs {
			_ = s.repo.DeleteReference(ctx, modelID, r.ID)
		}
	}
	for _, f := range snap.Fields {
		f.ModelID = modelID
		if err := s.repo.CreateField(ctx, f); err != nil {
			return domain.MetaModel{}, err
		}
	}
	for _, r := range snap.References {
		r.ModelID = modelID
		if err := s.repo.CreateReference(ctx, r); err != nil {
			return domain.MetaModel{}, err
		}
	}
	m.CurrentVersion = versionNo
	m.Status = domain.MetaModelStatusPublished
	m.UpdatedAt = time.Now()
	if err := s.repo.UpdateModel(ctx, m); err != nil {
		return domain.MetaModel{}, err
	}
	return m, nil
}

func (s *MetaService) validateNoCycle(ctx context.Context, modelID string) error {
	models, err := s.repo.ListModels(ctx, "")
	if err != nil {
		return err
	}
	graph := map[string][]string{}
	for _, m := range models {
		refs, err := s.repo.ListReferences(ctx, m.ID)
		if err != nil {
			return err
		}
		for _, r := range refs {
			graph[m.ID] = append(graph[m.ID], r.TargetModelID)
		}
	}
	vis := map[string]int{}
	var dfs func(string) bool
	dfs = func(n string) bool {
		vis[n] = 1
		for _, nxt := range graph[n] {
			if vis[nxt] == 1 {
				return true
			}
			if vis[nxt] == 0 && dfs(nxt) {
				return true
			}
		}
		vis[n] = 2
		return false
	}
	if vis[modelID] == 0 {
		if dfs(modelID) {
			return fmt.Errorf("cycle reference detected")
		}
	}
	return nil
}
