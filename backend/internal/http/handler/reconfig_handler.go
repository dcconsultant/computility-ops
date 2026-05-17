package handler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ReconfigHandler struct {
	svc  *service.ReconfigService
	mu   sync.RWMutex
	jobs map[string]*reconfigPlanJob
}

type reconfigPlanJob struct {
	JobID      string                                `json:"job_id"`
	Status     string                                `json:"status"`
	Progress   service.ReconfigPlanProgress          `json:"progress"`
	Result     *service.ReconfigPlanCalculateResponse `json:"result,omitempty"`
	Error      string                                `json:"error,omitempty"`
	StartedAt  time.Time                             `json:"started_at"`
	FinishedAt *time.Time                            `json:"finished_at,omitempty"`
}

func NewReconfigHandler(svc *service.ReconfigService) *ReconfigHandler {
	return &ReconfigHandler{svc: svc, jobs: map[string]*reconfigPlanJob{}}
}

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

func (h *ReconfigHandler) normalizeReq(req CalculateReconfigPlanReq) service.ReconfigPlanCalculateRequest {
	sns := append([]string(nil), req.Scope.SNs...)
	if strings.TrimSpace(req.Scope.SNInput) != "" {
		sns = append(sns, strings.Fields(req.Scope.SNInput)...)
	}
	if len(sns) > 1000 {
		sns = sns[:1000]
	}
	return service.ReconfigPlanCalculateRequest{
		Target: req.Target,
		Scope: service.ReconfigScopeConfig{PSAList: req.Scope.PSAList, ConfigTypes: req.Scope.ConfigTypes, SNs: sns},
		GoalValueScore: req.GoalValueScore,
	}
}

func (h *ReconfigHandler) CalculatePlan(c *gin.Context) {
	c.Set("audit_action", "reconfig.calculate_plan")
	var req CalculateReconfigPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查改配配置")
		return
	}
	out, err := h.svc.CalculatePlan(c.Request.Context(), h.normalizeReq(req))
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, out)
}

func (h *ReconfigHandler) StartPlan(c *gin.Context) {
	c.Set("audit_action", "reconfig.calculate_plan.start")
	var req CalculateReconfigPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查改配配置")
		return
	}
	jobID := fmt.Sprintf("reconfig-%d", time.Now().UnixNano())
	job := &reconfigPlanJob{JobID: jobID, Status: "running", StartedAt: time.Now(), Progress: service.ReconfigPlanProgress{Stage: "queued", Percent: 0, Message: "任务排队中"}}
	h.mu.Lock()
	h.jobs[jobID] = job
	h.mu.Unlock()

	go func() {
		ctx := context.Background()
		out, err := h.svc.CalculatePlanWithProgress(ctx, h.normalizeReq(req), func(p service.ReconfigPlanProgress) {
			h.mu.Lock()
			if j, ok := h.jobs[jobID]; ok {
				j.Progress = p
			}
			h.mu.Unlock()
		})
		h.mu.Lock()
		defer h.mu.Unlock()
		if j, ok := h.jobs[jobID]; ok {
			now := time.Now()
			j.FinishedAt = &now
			if err != nil {
				j.Status = "failed"
				j.Error = err.Error()
				return
			}
			j.Status = "done"
			j.Result = &out
			j.Progress.Stage = "done"
			j.Progress.Percent = 100
			j.Progress.Message = "计算完成"
		}
	}()

	ok(c, gin.H{"job_id": jobID, "status": "running"})
}

func (h *ReconfigHandler) GetPlanProgress(c *gin.Context) {
	c.Set("audit_action", "reconfig.calculate_plan.progress")
	jobID := strings.TrimSpace(c.Param("job_id"))
	h.mu.RLock()
	job, found := h.jobs[jobID]
	h.mu.RUnlock()
	if !found {
		fail(c, 40401, "任务不存在")
		return
	}
	ok(c, gin.H{"job_id": job.JobID, "status": job.Status, "progress": job.Progress, "error": job.Error, "started_at": job.StartedAt, "finished_at": job.FinishedAt})
}

func (h *ReconfigHandler) GetPlanResult(c *gin.Context) {
	c.Set("audit_action", "reconfig.calculate_plan.result")
	jobID := strings.TrimSpace(c.Param("job_id"))
	h.mu.RLock()
	job, found := h.jobs[jobID]
	h.mu.RUnlock()
	if !found {
		fail(c, 40401, "任务不存在")
		return
	}
	if job.Status == "failed" {
		fail(c, 50001, job.Error)
		return
	}
	if job.Status != "done" || job.Result == nil {
		ok(c, gin.H{"job_id": jobID, "status": job.Status})
		return
	}
	ok(c, gin.H{"job_id": jobID, "status": job.Status, "result": job.Result})
}
