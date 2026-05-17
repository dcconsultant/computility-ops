package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	JobID      string                                 `json:"job_id"`
	Status     string                                 `json:"status"`
	Progress   service.ReconfigPlanProgress           `json:"progress"`
	Result     *service.ReconfigPlanCalculateResponse `json:"result,omitempty"`
	Error      string                                 `json:"error,omitempty"`
	StartedAt  time.Time                              `json:"started_at"`
	FinishedAt *time.Time                             `json:"finished_at,omitempty"`
}

type reconfigPlanSnapshot struct {
	PlanID               string                         `json:"plan_id"`
	CreatedAt            time.Time                      `json:"created_at"`
	Target               service.ReconfigTargetConfig   `json:"target"`
	ScopeServerCount     int                            `json:"scope_server_count"`
	SuccessServerCount   int                            `json:"success_server_count"`
	SuccessCoreCount     float64                        `json:"success_core_count"`
	ReconfigServerCount  int                            `json:"reconfig_server_count"`
	DismantleServerCount int                            `json:"dismantle_server_count"`
	Hosts                []map[string]any               `json:"hosts"`
	Actions              []service.ReconfigActionRow    `json:"actions"`
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
			_ = saveReconfigPlanSnapshot(jobID, req.Target, out)
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

func saveReconfigPlanSnapshot(jobID string, target service.ReconfigTargetConfig, out service.ReconfigPlanCalculateResponse) error {
	summary := out.Summary
	scopeCnt := toInt(summary["scope_server_count"])
	successServerCnt := toInt(summary["success_server_count"])
	successCoreCnt := toFloat(summary["success_core_count"])

	hosts := make([]map[string]any, 0)
	reconfigSet := map[string]struct{}{}
	dismantleSet := map[string]struct{}{}
	for _, a := range out.Actions {
		hosts = append(hosts, map[string]any{
			"target_sn": a.TargetSN,
			"gap_type": a.GapType,
			"gap_qty": a.GapQty,
			"part_details": a.PartDetails,
		})
		if strings.Contains(a.Source, "SN:") {
			parts := strings.Split(a.Source, "SN:")
			for i := 1; i < len(parts); i++ {
				sn := strings.Fields(parts[i])[0]
				if sn != "" {
					dismantleSet[sn] = struct{}{}
				}
			}
		}
		reconfigSet[a.TargetSN] = struct{}{}
	}

	snapshot := reconfigPlanSnapshot{
		PlanID:               fmt.Sprintf("%s-%d", jobID, time.Now().Unix()),
		CreatedAt:            time.Now(),
		Target:               target,
		ScopeServerCount:     scopeCnt,
		SuccessServerCount:   successServerCnt,
		SuccessCoreCount:     successCoreCnt,
		ReconfigServerCount:  len(reconfigSet),
		DismantleServerCount: len(dismantleSet),
		Hosts:                hosts,
		Actions:              out.Actions,
	}

	base := filepath.Join("backend", "logs", "reconfig-plans")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	filename := filepath.Join(base, fmt.Sprintf("%s.json", snapshot.PlanID))
	return os.WriteFile(filename, payload, 0o644)
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	default:
		return 0
	}
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		n, _ := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return n
	default:
		return 0
	}
}
