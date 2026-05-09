package handler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type MetaImportJob struct {
	JobID      string         `json:"job_id"`
	ModelID    string         `json:"model_id"`
	Status     string         `json:"status"`
	Total      int            `json:"total"`
	Processed  int            `json:"processed"`
	Success    int            `json:"success"`
	Failed     int            `json:"failed"`
	Errors     []map[string]any `json:"errors"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	Message    string         `json:"message,omitempty"`
}

type MetaHandler struct {
	service *service.MetaService
	jobs    map[string]*MetaImportJob
	mu      sync.RWMutex
}

func NewMetaHandler(s *service.MetaService) *MetaHandler { return &MetaHandler{service: s, jobs: map[string]*MetaImportJob{}} }

func (h *MetaHandler) CreateModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.create")
	var req CreateMetaModelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查模型字段")
		return
	}
	model, err := h.service.CreateModel(c.Request.Context(), service.CreateModelInput{ModelCode: req.ModelCode, ModelName: req.ModelName, Description: req.Description})
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, model)
}

func (h *MetaHandler) ListModels(c *gin.Context) {
	c.Set("audit_action", "meta.models.list")
	status := strings.TrimSpace(c.Query("status"))
	models, err := h.service.ListModels(c.Request.Context(), status)
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": models, "total": len(models), "page": 1, "page_size": len(models)})
}

func (h *MetaHandler) GetModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.get")
	model, fields, err := h.service.GetModel(c.Request.Context(), c.Param("model_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"model": model, "fields": fields})
}

func (h *MetaHandler) UpdateModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.update")
	var req UpdateMetaModelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查模型字段")
		return
	}
	model, err := h.service.UpdateModel(c.Request.Context(), c.Param("model_id"), service.UpdateModelInput{ModelName: req.ModelName, Description: req.Description})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, model)
}

func (h *MetaHandler) ArchiveModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.archive")
	model, err := h.service.ArchiveModel(c.Request.Context(), c.Param("model_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, model)
}

func (h *MetaHandler) CloneModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.clone")
	var req CloneMetaModelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查模型字段")
		return
	}
	model, err := h.service.CloneModel(c.Request.Context(), c.Param("model_id"), service.CreateModelInput{ModelCode: req.ModelCode, ModelName: req.ModelName, Description: req.Description})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, model)
}

func (h *MetaHandler) DeleteModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.delete")
	if err := h.service.DeleteModel(c.Request.Context(), c.Param("model_id")); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "model_id": c.Param("model_id")})
}

func (h *MetaHandler) CreateField(c *gin.Context) {
	c.Set("audit_action", "meta.fields.create")
	var req CreateMetaFieldReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查属性字段")
		return
	}
	field, err := h.service.CreateField(c.Request.Context(), c.Param("model_id"), service.CreateFieldInput{
		FieldCode: req.FieldCode, FieldName: req.FieldName, Category: req.Category, ValueType: req.ValueType,
		Required: req.Required, Unique: req.Unique, Filterable: req.Filterable, Sortable: req.Sortable, Visible: req.Visible,
		DefaultValue: req.DefaultValue, ValidationRule: req.ValidationRule,
		EnumOptions: toDomainEnumOptions(req.EnumOptions),
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, field)
}

func (h *MetaHandler) UpdateField(c *gin.Context) {
	c.Set("audit_action", "meta.fields.update")
	var req UpdateMetaFieldReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查属性字段")
		return
	}
	field, err := h.service.UpdateField(c.Request.Context(), c.Param("model_id"), c.Param("field_id"), service.UpdateFieldInput{
		FieldName: req.FieldName, Category: req.Category, ValueType: req.ValueType,
		Required: req.Required, Unique: req.Unique, Filterable: req.Filterable, Sortable: req.Sortable, Visible: req.Visible,
		DefaultValue: req.DefaultValue, ValidationRule: req.ValidationRule,
		EnumOptions: toDomainEnumOptions(req.EnumOptions),
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, field)
}

func (h *MetaHandler) DeleteField(c *gin.Context) {
	c.Set("audit_action", "meta.fields.delete")
	if err := h.service.DeleteField(c.Request.Context(), c.Param("model_id"), c.Param("field_id")); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "field_id": c.Param("field_id")})
}

func (h *MetaHandler) ReorderFields(c *gin.Context) {
	c.Set("audit_action", "meta.fields.reorder")
	var req ReorderMetaFieldsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查排序字段")
		return
	}
	fields, err := h.service.ReorderFields(c.Request.Context(), c.Param("model_id"), req.FieldIDs)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, fields)
}

func (h *MetaHandler) CreateReference(c *gin.Context) {
	c.Set("audit_action", "meta.references.create")
	var req CreateMetaReferenceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查关联字段")
		return
	}
	ref, err := h.service.CreateReference(c.Request.Context(), c.Param("model_id"), service.CreateReferenceInput{
		SourceFieldID: req.SourceFieldID, TargetModelID: req.TargetModelID, TargetFieldID: req.TargetFieldID,
		DisplayFields: req.DisplayFields, OnDeleteAction: req.OnDeleteAction,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, ref)
}

func (h *MetaHandler) UpdateReference(c *gin.Context) {
	c.Set("audit_action", "meta.references.update")
	var req UpdateMetaReferenceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查关联字段")
		return
	}
	ref, err := h.service.UpdateReference(c.Request.Context(), c.Param("model_id"), c.Param("ref_id"), service.UpdateReferenceInput{
		SourceFieldID: req.SourceFieldID, TargetModelID: req.TargetModelID, TargetFieldID: req.TargetFieldID,
		DisplayFields: req.DisplayFields, OnDeleteAction: req.OnDeleteAction,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, ref)
}

func (h *MetaHandler) DeleteReference(c *gin.Context) {
	c.Set("audit_action", "meta.references.delete")
	if err := h.service.DeleteReference(c.Request.Context(), c.Param("model_id"), c.Param("ref_id")); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "ref_id": c.Param("ref_id")})
}

func (h *MetaHandler) ListReferences(c *gin.Context) {
	c.Set("audit_action", "meta.references.list")
	refs, err := h.service.ListReferences(c.Request.Context(), c.Param("model_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": refs, "total": len(refs), "page": 1, "page_size": len(refs)})
}

func (h *MetaHandler) PublishModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.publish")
	var req PublishMetaModelReq
	_ = c.ShouldBindJSON(&req)
	v, err := h.service.PublishModel(c.Request.Context(), c.Param("model_id"), service.PublishModelInput{ChangeSummary: req.ChangeSummary, PublishedBy: req.PublishedBy})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, v)
}

func (h *MetaHandler) ListVersions(c *gin.Context) {
	c.Set("audit_action", "meta.models.versions.list")
	list, err := h.service.ListVersions(c.Request.Context(), c.Param("model_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": len(list), "page": 1, "page_size": len(list)})
}

func (h *MetaHandler) GetVersion(c *gin.Context) {
	c.Set("audit_action", "meta.models.versions.get")
	vno := 0
	if _, err := fmt.Sscanf(c.Param("version"), "%d", &vno); err != nil || vno <= 0 {
		fail(c, 40001, "invalid version")
		return
	}
	v, snap, err := h.service.GetVersion(c.Request.Context(), c.Param("model_id"), vno)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"version": v, "snapshot": snap})
}

func (h *MetaHandler) RollbackModel(c *gin.Context) {
	c.Set("audit_action", "meta.models.rollback")
	var req RollbackMetaModelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查回滚版本")
		return
	}
	m, err := h.service.RollbackModel(c.Request.Context(), c.Param("model_id"), req.Version)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, m)
}

func (h *MetaHandler) ListRecords(c *gin.Context) {
	c.Set("audit_action", "meta.records.list")
	m, fields, records, err := h.service.ListRecords(c.Request.Context(), c.Param("model_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, gin.H{"model": m, "fields": fields, "list": records, "total": len(records), "page": 1, "page_size": len(records)})
}

func (h *MetaHandler) CreateRecord(c *gin.Context) {
	c.Set("audit_action", "meta.records.create")
	var req UpsertMetaRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查记录字段")
		return
	}
	rec, err := h.service.CreateRecord(c.Request.Context(), c.Param("model_id"), service.UpsertRecordInput{Data: req.Data})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, rec)
}

func (h *MetaHandler) UpdateRecord(c *gin.Context) {
	c.Set("audit_action", "meta.records.update")
	var req UpsertMetaRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查记录字段")
		return
	}
	rec, err := h.service.UpdateRecord(c.Request.Context(), c.Param("model_id"), c.Param("record_id"), service.UpsertRecordInput{Data: req.Data})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, rec)
}

func (h *MetaHandler) DeleteRecord(c *gin.Context) {
	c.Set("audit_action", "meta.records.delete")
	if err := h.service.DeleteRecord(c.Request.Context(), c.Param("model_id"), c.Param("record_id")); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "record_id": c.Param("record_id")})
}

func (h *MetaHandler) ExportRecordTemplate(c *gin.Context) {
	c.Set("audit_action", "meta.records.template.export")
	_, fields, err := h.service.GetModel(c.Request.Context(), c.Param("model_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, field := range fields {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, field.FieldCode)
		_ = f.SetColWidth(sheet, cell[:1], cell[:1], maxFloat(float64(len([]rune(field.FieldName)))+4, 12))
	}
	for i, field := range fields {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		hint := field.FieldName
		if field.ValueType == "enum" && len(field.EnumOptions) > 0 {
			enums := make([]string, 0, len(field.EnumOptions))
			for _, op := range field.EnumOptions {
				if !op.Disabled {
					enums = append(enums, op.Value)
				}
			}
			hint = fmt.Sprintf("%s (enum: %s)", field.FieldName, strings.Join(enums, "/"))
		}
		_ = f.SetCellValue(sheet, cell, hint)
	}
	filename := fmt.Sprintf("meta-record-template-%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := f.Write(c.Writer); err != nil {
		fail(c, 50001, "模板导出失败")
		return
	}
	c.Set("audit_result", "ok")
}

func (h *MetaHandler) ImportRecords(c *gin.Context) {
	c.Set("audit_action", "meta.records.import")
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, 40002, "缺少上传文件")
		return
	}
	f, err := file.Open()
	if err != nil {
		fail(c, 50001, "读取上传文件失败")
		return
	}
	defer f.Close()
	xf, err := excelize.OpenReader(f)
	if err != nil {
		fail(c, 40003, "文件格式无效，请确认是标准 .xlsx")
		return
	}
	_, fields, err := h.service.GetModel(c.Request.Context(), c.Param("model_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	sheet := xf.GetSheetName(0)
	rows, err := xf.GetRows(sheet)
	if err != nil || len(rows) < 1 {
		fail(c, 40003, "Excel内容为空")
		return
	}
	headers := rows[0]
	fieldMap := map[int]domain.MetaField{}
	for i, hname := range headers {
		for _, field := range fields {
			if strings.TrimSpace(hname) == field.FieldCode {
				fieldMap[i] = field
				break
			}
		}
	}
	rowsData := make([]map[string]any, 0, len(rows)-1)
	for _, row := range rows[1:] {
		data := map[string]any{}
		for ci, fv := range row {
			if field, ok := fieldMap[ci]; ok {
				data[field.FieldCode] = strings.TrimSpace(fv)
			}
		}
		rowsData = append(rowsData, data)
	}

	jobID := fmt.Sprintf("meta-import-%d", time.Now().UnixNano())
	job := &MetaImportJob{
		JobID:     jobID,
		ModelID:   c.Param("model_id"),
		Status:    "running",
		Total:     len(rowsData),
		Processed: 0,
		Success:   0,
		Failed:    0,
		Errors:    make([]map[string]any, 0),
		StartedAt: time.Now(),
	}
	h.mu.Lock()
	h.jobs[jobID] = job
	h.mu.Unlock()

	ctx := context.Background()
	go func(modelID string, data []map[string]any) {
		result, runErr := h.service.ImportRecordsBatch(ctx, modelID, data)
		now := time.Now()
		h.mu.Lock()
		defer h.mu.Unlock()
		j, ok := h.jobs[jobID]
		if !ok {
			return
		}
		if runErr != nil {
			j.Status = "failed"
			j.Message = runErr.Error()
			j.FinishedAt = &now
			j.Processed = j.Total
			return
		}
		j.Status = "done"
		j.Success = result.Success
		j.Failed = result.Failed
		j.Errors = result.Errors
		j.Processed = result.Total
		j.FinishedAt = &now
	} (c.Param("model_id"), rowsData)

	ok(c, gin.H{"job_id": jobID, "status": "running", "total": len(rowsData)})
}

func (h *MetaHandler) GetImportJob(c *gin.Context) {
	c.Set("audit_action", "meta.records.import.job")
	jobID := strings.TrimSpace(c.Param("job_id"))
	h.mu.RLock()
	job, found := h.jobs[jobID]
	h.mu.RUnlock()
	if !found {
		fail(c, 40401, "import job not found")
		return
	}
	ok(c, job)
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func toDomainEnumOptions(in []MetaEnumOptionReq) []domain.MetaEnumOption {
	out := make([]domain.MetaEnumOption, 0, len(in))
	for _, v := range in {
		out = append(out, domain.MetaEnumOption{Value: v.Value, Label: v.Label, Disabled: v.Disabled})
	}
	return out
}
