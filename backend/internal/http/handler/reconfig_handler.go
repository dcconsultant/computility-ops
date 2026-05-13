package handler

import (
	"strings"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ReconfigHandler struct {
	svc *service.ReconfigService
}

func NewReconfigHandler(svc *service.ReconfigService) *ReconfigHandler { return &ReconfigHandler{svc: svc} }

type CalculateReconfigPlanReq struct {
	Target service.ReconfigTargetConfig `json:"target"`
	Scope  struct {
		PSAList     []string `json:"psa_list"`
		ConfigTypes []string `json:"config_types"`
		SNInput     string   `json:"sn_input"`
		SNs         []string `json:"sns"`
	} `json:"scope"`
	GoalValueScore float64 `json:"goal_value_score"`
}

func (h *ReconfigHandler) CalculatePlan(c *gin.Context) {
	c.Set("audit_action", "reconfig.calculate_plan")
	var req CalculateReconfigPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查改配配置")
		return
	}

	sns := append([]string(nil), req.Scope.SNs...)
	if strings.TrimSpace(req.Scope.SNInput) != "" {
		sns = append(sns, strings.Fields(req.Scope.SNInput)...)
	}
	if len(sns) > 1000 {
		sns = sns[:1000]
	}

	out, err := h.svc.CalculatePlan(c.Request.Context(), service.ReconfigPlanCalculateRequest{
		Target: req.Target,
		Scope: service.ReconfigScopeConfig{
			PSAList:     req.Scope.PSAList,
			ConfigTypes: req.Scope.ConfigTypes,
			SNs:         sns,
		},
		GoalValueScore: req.GoalValueScore,
	})
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, out)
}
