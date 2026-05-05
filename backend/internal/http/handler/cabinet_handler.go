package handler

import (
	"strings"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type CabinetHandler struct {
	service *service.CabinetService
}

func NewCabinetHandler(s *service.CabinetService) *CabinetHandler { return &CabinetHandler{service: s} }

func (h *CabinetHandler) GetUtilization(c *gin.Context) {
	c.Set("audit_action", "cabinet.utilization.get")
	data, err := h.service.GetUtilization(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, data)
}

func (h *CabinetHandler) UpdateUtilization(c *gin.Context) {
	c.Set("audit_action", "cabinet.utilization.update")
	var req struct {
		Utilization float64 `json:"utilization" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	data, err := h.service.UpdateUtilization(c.Request.Context(), req.Utilization)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, data)
}

func (h *CabinetHandler) List(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.list")
	rows, err := h.service.ListCabinetConfigs(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *CabinetHandler) Create(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.create")
	var req domain.CabinetConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	req.IDC = strings.TrimSpace(req.IDC)
	row, err := h.service.CreateCabinetConfig(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			fail(c, 40001, "机房+额定功率已存在")
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, row)
}

func (h *CabinetHandler) Update(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.update")
	var req domain.CabinetConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	id := atoiDefault(c.Param("id"), 0)
	req.ID = int64(id)
	req.IDC = strings.TrimSpace(req.IDC)
	row, err := h.service.UpdateCabinetConfig(c.Request.Context(), req)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "duplicate") {
			fail(c, 40001, "机房+额定功率已存在")
			return
		}
		if strings.Contains(lower, "not found") {
			fail(c, 40401, "记录不存在")
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, row)
}

func (h *CabinetHandler) Delete(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.delete")
	id := atoiDefault(c.Param("id"), 0)
	if id <= 0 {
		fail(c, 40001, "id无效")
		return
	}
	if err := h.service.DeleteCabinetConfig(c.Request.Context(), int64(id)); err != nil {
		fail(c, 50001, "删除失败")
		return
	}
	ok(c, gin.H{"deleted": true, "id": id})
}
