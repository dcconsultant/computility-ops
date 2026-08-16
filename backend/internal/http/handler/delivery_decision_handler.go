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
