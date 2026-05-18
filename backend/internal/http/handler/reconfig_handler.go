package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/xuri/excelize/v2"

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
	PlanID               string                       `json:"plan_id"`
	CreatedAt            time.Time                    `json:"created_at"`
	Target               service.ReconfigTargetConfig `json:"target"`
	ScopeServerCount     int                          `json:"scope_server_count"`
	SuccessServerCount   int                          `json:"success_server_count"`
	SuccessCoreCount     float64                      `json:"success_core_count"`
	ReconfigServerCount  int                          `json:"reconfig_server_count"`
	DismantleServerCount int                          `json:"dismantle_server_count"`
	PlannedReconfigCount int                          `json:"planned_reconfig_count"`
	SuccessReconfigCount int                          `json:"success_reconfig_count"`
	ResourceEfficiency   float64                      `json:"resource_efficiency"`
	ReconfigFee          int                          `json:"reconfig_fee"`
	Hosts                []map[string]any             `json:"hosts"`
	Actions              []service.ReconfigActionRow  `json:"actions"`
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
		Target:         req.Target,
		Scope:          service.ReconfigScopeConfig{PSAList: req.Scope.PSAList, ConfigTypes: req.Scope.ConfigTypes, SNs: sns},
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

func (h *ReconfigHandler) ListSavedPlans(c *gin.Context) {
	c.Set("audit_action", "reconfig.plan.saved.list")
	base := filepath.Join("backend", "logs", "reconfig-plans")
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			ok(c, gin.H{"list": []any{}, "total": 0})
			return
		}
		fail(c, 50001, "读取改配方案列表失败")
		return
	}
	list := make([]reconfigPlanSnapshot, 0)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		payload, readErr := os.ReadFile(filepath.Join(base, e.Name()))
		if readErr != nil {
			continue
		}
		var row reconfigPlanSnapshot
		if jsonErr := json.Unmarshal(payload, &row); jsonErr != nil {
			continue
		}
		list = append(list, row)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *ReconfigHandler) GetSavedPlan(c *gin.Context) {
	c.Set("audit_action", "reconfig.plan.saved.get")
	row, err := loadSavedPlan(strings.TrimSpace(c.Param("plan_id")))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fail(c, 40401, "方案不存在")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			fail(c, 50001, "方案内容损坏")
			return
		}
		fail(c, 50001, "读取改配方案失败")
		return
	}
	ok(c, row)
}

func (h *ReconfigHandler) ExportPlanResultActions(c *gin.Context) {
	c.Set("audit_action", "reconfig.plan.result.actions.export")
	jobID := strings.TrimSpace(c.Param("job_id"))
	h.mu.RLock()
	job, found := h.jobs[jobID]
	h.mu.RUnlock()
	if !found {
		fail(c, 40401, "任务不存在")
		return
	}
	if job.Status != "done" || job.Result == nil {
		fail(c, 40001, "任务未完成，无法导出执行清单")
		return
	}
	exportActionsExcel(c, job.Result.Actions, fmt.Sprintf("reconfig-actions-%s", jobID))
}

func (h *ReconfigHandler) ExportSavedPlanActions(c *gin.Context) {
	c.Set("audit_action", "reconfig.plan.saved.actions.export")
	row, err := loadSavedPlan(strings.TrimSpace(c.Param("plan_id")))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fail(c, 40401, "方案不存在")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			fail(c, 50001, "方案内容损坏")
			return
		}
		fail(c, 50001, "读取改配方案失败")
		return
	}
	exportActionsExcel(c, row.Actions, fmt.Sprintf("reconfig-actions-%s", row.PlanID))
}

func saveReconfigPlanSnapshot(jobID string, target service.ReconfigTargetConfig, out service.ReconfigPlanCalculateResponse) error {
	summary := out.Summary
	scopeCnt := toInt(summary["scope_server_count"])
	successServerCnt := toInt(summary["success_server_count"])
	successCoreCnt := toFloat(summary["success_core_count"])
	plannedReconfigCount := toInt(summary["planned_reconfig_count"])
	successReconfigCount := toInt(summary["success_reconfig_count"])
	resourceEfficiency := toFloat(summary["resource_efficiency"])

	hosts := make([]map[string]any, 0)
	targetSet := map[string]struct{}{}
	donorSet := map[string]struct{}{}
	successSet := map[string]struct{}{}
	if arr, ok := summary["success_server_sns"].([]any); ok {
		for _, it := range arr {
			sn := strings.TrimSpace(fmt.Sprintf("%v", it))
			if sn != "" {
				successSet[sn] = struct{}{}
			}
		}
	}

	for _, a := range out.Actions {
		hosts = append(hosts, map[string]any{
			"target_sn":    a.TargetSN,
			"gap_type":     a.GapType,
			"gap_qty":      a.GapQty,
			"part_details": a.PartDetails,
		})
		if sn := strings.TrimSpace(a.TargetSN); sn != "" {
			targetSet[sn] = struct{}{}
		}
		for _, sn := range extractSourceSNs(a.Source) {
			donorSet[sn] = struct{}{}
		}
	}

	dismantleSet := map[string]struct{}{}
	for sn := range donorSet {
		if _, inbound := targetSet[sn]; inbound {
			continue
		}
		if _, success := successSet[sn]; success {
			continue
		}
		dismantleSet[sn] = struct{}{}
	}
	reconfigServerCount := len(targetSet)
	dismantleServerCount := len(dismantleSet)
	reconfigFee := (reconfigServerCount + dismantleServerCount) * 70

	snapshot := reconfigPlanSnapshot{
		PlanID:               fmt.Sprintf("%s-%d", jobID, time.Now().Unix()),
		CreatedAt:            time.Now(),
		Target:               target,
		ScopeServerCount:     scopeCnt,
		SuccessServerCount:   successServerCnt,
		SuccessCoreCount:     successCoreCnt,
		ReconfigServerCount:  reconfigServerCount,
		DismantleServerCount: dismantleServerCount,
		PlannedReconfigCount: plannedReconfigCount,
		SuccessReconfigCount: successReconfigCount,
		ResourceEfficiency:   resourceEfficiency,
		ReconfigFee:          reconfigFee,
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

func loadSavedPlan(planID string) (reconfigPlanSnapshot, error) {
	if planID == "" {
		return reconfigPlanSnapshot{}, fmt.Errorf("plan_id empty")
	}
	filename := filepath.Join("backend", "logs", "reconfig-plans", fmt.Sprintf("%s.json", planID))
	payload, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return reconfigPlanSnapshot{}, fmt.Errorf("not found")
		}
		return reconfigPlanSnapshot{}, err
	}
	var row reconfigPlanSnapshot
	if err := json.Unmarshal(payload, &row); err != nil {
		return reconfigPlanSnapshot{}, fmt.Errorf("invalid")
	}
	return row, nil
}

func exportActionsExcel(c *gin.Context, actions []service.ReconfigActionRow, filenamePrefix string) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetSheetName(sheet, "执行清单")
	headers := []string{"目标SN", "缺口类型", "缺口数量", "推荐来源", "配件明细", "跨机房搬迁", "执行动作", "规则命中说明"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue("执行清单", cell, h)
	}
	for i, a := range actions {
		row := i + 2
		_ = f.SetCellValue("执行清单", fmt.Sprintf("A%d", row), a.TargetSN)
		_ = f.SetCellValue("执行清单", fmt.Sprintf("B%d", row), a.GapType)
		_ = f.SetCellValue("执行清单", fmt.Sprintf("C%d", row), a.GapQty)
		_ = f.SetCellValue("执行清单", fmt.Sprintf("D%d", row), a.Source)
		_ = f.SetCellValue("执行清单", fmt.Sprintf("E%d", row), a.PartDetails)
		_ = f.SetCellValue("执行清单", fmt.Sprintf("F%d", row), a.CrossIDC)
		_ = f.SetCellValue("执行清单", fmt.Sprintf("G%d", row), a.Action)
		_ = f.SetCellValue("执行清单", fmt.Sprintf("H%d", row), a.RuleHit)
	}
	_ = f.SetColWidth("执行清单", "A", "H", 24)
	buf, err := f.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出执行清单失败")
		return
	}
	filename := fmt.Sprintf("%s-%s.xlsx", filenamePrefix, time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func extractSourceSNs(source string) []string {
	if !strings.Contains(source, "SN:") {
		return nil
	}
	parts := strings.Split(source, "SN:")
	out := make([]string, 0, len(parts)-1)
	seen := map[string]struct{}{}
	for i := 1; i < len(parts); i++ {
		raw := strings.TrimSpace(parts[i])
		if raw == "" {
			continue
		}
		tokens := strings.Fields(raw)
		if len(tokens) == 0 {
			continue
		}
		sn := strings.TrimFunc(tokens[0], func(r rune) bool {
			return unicode.IsPunct(r) || unicode.IsSpace(r)
		})
		if sn == "" {
			continue
		}
		if _, ok := seen[sn]; ok {
			continue
		}
		seen[sn] = struct{}{}
		out = append(out, sn)
	}
	return out
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
