package handler

import (
	"strings"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DeliveryDecisionHandler struct {
	svc *service.DeliveryDecisionService
}

func NewDeliveryDecisionHandler(svc *service.DeliveryDecisionService) *DeliveryDecisionHandler {
	return &DeliveryDecisionHandler{svc: svc}
}

func (h *DeliveryDecisionHandler) GetDefaults(c *gin.Context) {
	c.Set("audit_action", "delivery_decision.defaults.get")
	country := c.Query("country")
	v := h.svc.Defaults(c.Request.Context(), country)
	ok(c, v)
}

func (h *DeliveryDecisionHandler) GetConfig(c *gin.Context) {
	c.Set("audit_action", "delivery_decision.config.get")
	state, found, err := h.svc.GetConfig(c.Request.Context())
	if err != nil {
		fail(c, 50001, "读取交付方式决策配置失败")
		return
	}
	if !found {
		ok(c, gin.H{"found": false})
		return
	}
	ok(c, gin.H{"found": true, "state": state})
}

func (h *DeliveryDecisionHandler) SaveConfig(c *gin.Context) {
	c.Set("audit_action", "delivery_decision.config.save")
	var req domain.DeliveryDecisionCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	req.Input.Country = strings.TrimSpace(req.Input.Country)
	req.Input.Currency = strings.TrimSpace(req.Input.Currency)
	state, err := h.svc.SaveConfig(c.Request.Context(), req.Input)
	if err != nil {
		fail(c, 40004, err.Error())
		return
	}
	ok(c, gin.H{"saved": true, "state": state})
}

func (h *DeliveryDecisionHandler) Calculate(c *gin.Context) {
	c.Set("audit_action", "delivery_decision.calculate")
	var req domain.DeliveryDecisionCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	req.Input.Country = strings.TrimSpace(req.Input.Country)
	req.Input.Currency = strings.TrimSpace(req.Input.Currency)
	v, err := h.svc.Calculate(c.Request.Context(), req)
	if err != nil {
		fail(c, 40004, err.Error())
		return
	}
	c.Header("Cache-Control", "no-store")
	ok(c, v)
}
