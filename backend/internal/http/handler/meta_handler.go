package handler

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type MetaHandler struct {
	service *service.MetaService
}

func NewMetaHandler(s *service.MetaService) *MetaHandler { return &MetaHandler{service: s} }

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
	if strings.TrimSpace(sheet) == "" {
		fail(c, 40003, "Excel内容为空")
		return
	}
	iter, err := xf.Rows(sheet)
	if err != nil {
		fail(c, 40003, "Excel内容为空")
		return
	}
	defer iter.Close()
	if !iter.Next() {
		fail(c, 40003, "Excel内容为空")
		return
	}
	headers, err := iter.Columns()
	if err != nil {
		fail(c, 40003, "Excel内容为空")
		return
	}
	fieldMap := map[int]domain.MetaField{}
	for i, hname := range headers {
		for _, field := range fields {
			if strings.TrimSpace(hname) == field.FieldCode {
				fieldMap[i] = field
				break
			}
		}
	}
	total := 0
	for iter.Next() {
		total++
	}
	if total == 0 {
		fail(c, 40003, "Excel内容为空")
		return
	}

	jobID := fmt.Sprintf("meta-import-%d", time.Now().UnixNano())
	job := domain.MetaImportJob{
		JobID:     jobID,
		ModelID:   c.Param("model_id"),
		Status:    "running",
		Total:     total,
		Processed: 0,
		Success:   0,
		Failed:    0,
		Errors:    make([]map[string]any, 0),
		StartedAt: time.Now(),
	}
	if err := h.service.CreateImportJob(c.Request.Context(), job); err != nil {
		fail(c, 50001, "创建导入任务失败")
		return
	}

	ctx := context.Background()
	importUniqueMode := strings.ToLower(strings.TrimSpace(c.PostForm("unique_mode")))
	if importUniqueMode != "off" && importUniqueMode != "strict" {
		importUniqueMode = ""
	}
	primaryFieldCode := ""
	if len(fields) > 0 {
		primaryFieldCode = strings.TrimSpace(fields[0].FieldCode)
	}
	go h.runImportJob(ctx, jobID, c.Param("model_id"), xf, sheet, fieldMap, primaryFieldCode, importUniqueMode)
	ok(c, gin.H{"job_id": jobID, "status": "running", "total": total})
}

func (h *MetaHandler) runImportJob(ctx context.Context, jobID, modelID string, xf *excelize.File, sheet string, fieldMap map[int]domain.MetaField, primaryFieldCode, importUniqueMode string) {
	iter, err := xf.Rows(sheet)
	if err != nil {
		h.failJob(ctx, jobID, "读取Excel失败")
		return
	}
	defer iter.Close()
	if !iter.Next() {
		h.failJob(ctx, jobID, "Excel内容为空")
		return
	}
	_, _ = iter.Columns()

	chunkSize := 2000
	chunk := make([]map[string]any, 0, chunkSize)
	chunkKeys := make([]string, 0, chunkSize)
	processed := 0
	for iter.Next() {
		cols, colErr := iter.Columns()
		if colErr != nil {
			h.failJob(ctx, jobID, "读取Excel失败")
			return
		}
		data := map[string]any{}
		for ci, fv := range cols {
			if field, ok := fieldMap[ci]; ok {
				data[field.FieldCode] = strings.TrimSpace(fv)
			}
		}
		displayKey := ""
		if len(cols) > 0 {
			displayKey = strings.TrimSpace(cols[0])
		}
		if displayKey == "" && primaryFieldCode != "" {
			displayKey = strings.TrimSpace(fmt.Sprintf("%v", data[primaryFieldCode]))
		}
		chunk = append(chunk, data)
		chunkKeys = append(chunkKeys, displayKey)
		if len(chunk) >= chunkSize {
			if err := h.flushImportChunk(ctx, jobID, modelID, chunk, chunkKeys, processed, importUniqueMode); err != nil {
				h.failJob(ctx, jobID, err.Error())
				return
			}
			processed += len(chunk)
			chunk = chunk[:0]
			chunkKeys = chunkKeys[:0]
		}
	}
	if len(chunk) > 0 {
		if err := h.flushImportChunk(ctx, jobID, modelID, chunk, chunkKeys, processed, importUniqueMode); err != nil {
			h.failJob(ctx, jobID, err.Error())
			return
		}
	}
	job, err := h.service.GetImportJob(ctx, jobID)
	if err != nil {
		return
	}
	now := time.Now()
	job.Status = "done"
	job.FinishedAt = &now
	job.Message = ""
	_ = h.service.UpdateImportJob(ctx, job)
}

func (h *MetaHandler) flushImportChunk(ctx context.Context, jobID, modelID string, chunk []map[string]any, chunkKeys []string, processedBefore int, importUniqueMode string) error {
	res, err := h.service.ImportRecordsBatchWithMode(ctx, modelID, chunk, importUniqueMode)
	if err != nil {
		return err
	}
	job, err := h.service.GetImportJob(ctx, jobID)
	if err != nil {
		return err
	}
	for _, e := range res.Errors {
		rowNo, _ := strconv.Atoi(fmt.Sprintf("%v", e["row"]))
		errorIndex := rowNo - 2
		e["row"] = rowNo + processedBefore
		if errorIndex >= 0 && errorIndex < len(chunkKeys) {
			e["key"] = chunkKeys[errorIndex]
		} else {
			e["key"] = ""
		}
		job.Errors = append(job.Errors, e)
	}
	job.Processed = processedBefore + len(chunk)
	job.Success += res.Success
	job.Failed += res.Failed
	if job.Errors == nil {
		job.Errors = make([]map[string]any, 0)
	}
	return h.service.UpdateImportJob(ctx, job)
}

func (h *MetaHandler) failJob(ctx context.Context, jobID, msg string) {
	job, err := h.service.GetImportJob(ctx, jobID)
	if err != nil {
		return
	}
	now := time.Now()
	job.Status = "failed"
	job.Message = msg
	job.FinishedAt = &now
	if job.Processed < job.Total {
		job.Processed = job.Total
	}
	_ = h.service.UpdateImportJob(ctx, job)
}

func (h *MetaHandler) GetImportJob(c *gin.Context) {
	c.Set("audit_action", "meta.records.import.job")
	jobID := strings.TrimSpace(c.Param("job_id"))
	job, err := h.service.GetImportJob(c.Request.Context(), jobID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, "import job not found")
			return
		}
		fail(c, 50001, "查询导入任务失败")
		return
	}
	ok(c, job)
}

func (h *MetaHandler) ExportImportJobErrorsCSV(c *gin.Context) {
	c.Set("audit_action", "meta.records.import.errors.export")
	jobID := strings.TrimSpace(c.Param("job_id"))
	job, err := h.service.GetImportJob(c.Request.Context(), jobID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, "import job not found")
			return
		}
		fail(c, 50001, "查询导入任务失败")
		return
	}
	if job.Status == "cleaned" {
		fail(c, 40001, "导入任务明细已自动清理，无法导出")
		return
	}
	buf := &strings.Builder{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"row", "key", "error"})
	for _, it := range job.Errors {
		_ = w.Write([]string{fmt.Sprintf("%v", it["row"]), fmt.Sprintf("%v", it["key"]), fmt.Sprintf("%v", it["error"])})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filename := fmt.Sprintf("meta-import-errors-%s.csv", time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.String(200, "\uFEFF%s", buf.String())
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
