package handler

import (
	"strings"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ResourcePlanningHandler struct {
	svc *service.ResourcePlanningService
}

func NewResourcePlanningHandler(svc *service.ResourcePlanningService) *ResourcePlanningHandler {
	return &ResourcePlanningHandler{svc: svc}
}

func (h *ResourcePlanningHandler) GetConfig(c *gin.Context) {
	c.Set("audit_action", "resource_planning.get_config")
	state, found, err := h.svc.GetConfig(c.Request.Context())
	if err != nil {
		fail(c, 50001, "读取资源规划配置失败")
		return
	}
	if !found {
		ok(c, gin.H{"found": false})
		return
	}
	ok(c, gin.H{"found": true, "state": state})
}

func (h *ResourcePlanningHandler) SaveConfig(c *gin.Context) {
	c.Set("audit_action", "resource_planning.save_config")
	var req service.ResourcePlanningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查资源规划输入")
		return
	}
	if err := h.svc.SaveConfig(c.Request.Context(), req); err != nil {
		fail(c, 50001, "保存资源规划配置失败")
		return
	}
	ok(c, gin.H{"saved": true})
}

func (h *ResourcePlanningHandler) Calculate(c *gin.Context) {
	c.Set("audit_action", "resource_planning.calculate")
	var req service.ResourcePlanningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查资源规划输入")
		return
	}
	if strings.TrimSpace(req.NonBusinessPSAs) == "" {
		fail(c, 40001, "非业务PSA不能为空")
		return
	}
	if strings.TrimSpace(req.DisposalPSAs) == "" {
		fail(c, 40001, "处置PSA不能为空")
		return
	}
	out, err := h.svc.Calculate(c.Request.Context(), req)
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, out)
}
