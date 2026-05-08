package handler

import (
	"fmt"
	"strings"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type MetaHandler struct{ service *service.MetaService }

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
