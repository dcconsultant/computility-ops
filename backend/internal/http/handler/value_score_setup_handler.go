package handler

import (
	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ValueScoreSetupHandler struct {
	svc *service.ValueScoreSetupService
}

func NewValueScoreSetupHandler(svc *service.ValueScoreSetupService) *ValueScoreSetupHandler {
	return &ValueScoreSetupHandler{svc: svc}
}

func (h *ValueScoreSetupHandler) GetCostSettings(c *gin.Context) {
	c.Set("audit_action", "value_score.cost_settings.get")
	v, err := h.svc.GetCostSettings(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, v)
}

func (h *ValueScoreSetupHandler) UpdateCostSettings(c *gin.Context) {
	c.Set("audit_action", "value_score.cost_settings.update")
	var req domain.ValueScoreCostSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	v, err := h.svc.UpdateCostSettings(c.Request.Context(), req)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, v)
}

func (h *ValueScoreSetupHandler) CheckPackageCabinetMapping(c *gin.Context) {
	c.Set("audit_action", "value_score.package_cabinet_mapping.check")
	rows, err := h.svc.CheckPackageCabinetMapping(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	unmatched := 0
	for _, r := range rows {
		if !r.Matched {
			unmatched++
		}
	}
	status := "ok"
	if unmatched > 0 {
		status = "warning"
	}
	ok(c, gin.H{"status": status, "unmatched": unmatched, "list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

