package handler

import (
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ValueScoreSetupHandler struct {
	svc *service.ValueScoreSetupService
}

func NewValueScoreSetupHandler(svc *service.ValueScoreSetupService) *ValueScoreSetupHandler {
	return &ValueScoreSetupHandler{svc: svc}
}

func (h *ValueScoreSetupHandler) GetCabinetBaseline(c *gin.Context) {
	c.Set("audit_action", "value_score.cabinet_baseline.get")
	v, err := h.svc.GetCabinetBaseline(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, v)
}

