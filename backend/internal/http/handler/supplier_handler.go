package handler

import (
	"strings"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SupplierHandler struct {
	service *service.SupplierService
}

func NewSupplierHandler(s *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{service: s}
}

func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.create")
	var req UpsertSupplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查供应商字段")
		return
	}
	item, err := h.service.CreateSupplier(c.Request.Context(), service.UpsertSupplierInput(req))
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, item)
}

func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.update")
	var req UpsertSupplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查供应商字段")
		return
	}
	item, err := h.service.UpdateSupplier(c.Request.Context(), c.Param("supplier_id"), service.UpsertSupplierInput(req))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, item)
}

func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.get")
	item, err := h.service.GetSupplier(c.Request.Context(), c.Param("supplier_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, item)
}

func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	c.Set("audit_action", "suppliers.list")
	list, err := h.service.ListSuppliers(c.Request.Context(), c.Query("q"))
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": len(list), "page": 1, "page_size": len(list)})
}

func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.delete")
	if err := h.service.DeleteSupplier(c.Request.Context(), c.Param("supplier_id")); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "supplier_id": c.Param("supplier_id")})
}
