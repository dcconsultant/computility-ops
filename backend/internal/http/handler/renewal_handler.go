package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type RenewalHandler struct {
	service *service.RenewalService
}

func NewRenewalHandler(s *service.RenewalService) *RenewalHandler { return &RenewalHandler{service: s} }

func (h *RenewalHandler) CreatePlan(c *gin.Context) {
	c.Set("audit_action", "renewals.create_plan")
	var req CreatePlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查 target_date/target_cores/存储目标")
		return
	}
	plan, err := h.service.CreatePlan(c.Request.Context(), service.CreatePlanInput{
		TargetDate:           req.TargetDate,
		ExcludedEnvironments: req.ExcludedEnvironments,
		ExcludedPSAs:         req.ExcludedPSAs,
		TargetCores:          req.TargetCores,
		WarmTargetStorageTB:  req.WarmTargetStorageTB,
		HotTargetStorageTB:   req.HotTargetStorageTB,
	})
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, plan)
}

func (h *RenewalHandler) GetPlan(c *gin.Context) {
	c.Set("audit_action", "renewals.get_plan")
	plan, err := h.service.GetPlan(c.Request.Context(), c.Param("plan_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, plan)
}

func (h *RenewalHandler) ExportPlan(c *gin.Context) {
	c.Set("audit_action", "renewals.export_plan")
	planID := c.Param("plan_id")
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "xlsx")))
	if format != "xlsx" && format != "csv" {
		fail(c, 40001, "format must be xlsx or csv")
		return
	}

	plan, err := h.service.GetPlan(c.Request.Context(), planID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}

	filename := fmt.Sprintf("renewal_plan_%s.%s", planID, format)
	if format == "csv" {
		buf, err := buildCSV(plan)
		if err != nil {
			fail(c, 50001, err.Error())
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Data(http.StatusOK, "text/csv", buf.Bytes())
		return
	}

	buf, err := buildXLSX(plan)
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func buildCSV(plan domain.RenewalPlan) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write([]string{"plan_id", "target_date", "target_cores", "warm_target_storage_tb", "hot_target_storage_tb", "unmatched_config_count", "unmatched_config_types", "selected_cores", "selected_storage_tb", "selected_count"}); err != nil {
		return nil, err
	}
	if err := w.Write([]string{plan.PlanID, plan.TargetDate, fmt.Sprint(plan.TargetCores), fmt.Sprintf("%.4f", plan.WarmTargetStorageTB), fmt.Sprintf("%.4f", plan.HotTargetStorageTB), fmt.Sprint(plan.UnmatchedConfigCount), strings.Join(plan.UnmatchedConfigTypes, "|"), fmt.Sprint(plan.SelectedCores), fmt.Sprintf("%.4f", plan.SelectedStorageTB), fmt.Sprint(plan.SelectedCount)}); err != nil {
		return nil, err
	}
	if err := w.Write([]string{}); err != nil {
		return nil, err
	}
	if err := w.Write([]string{"rank", "bucket", "sn", "manufacturer", "model", "environment", "config_type", "cpu_logical_cores", "storage_capacity_tb", "psa", "arch_standardized_factor", "base_score", "afr_old", "afr_avg", "failure_adjust_factor", "final_score", "special_policy"}); err != nil {
		return nil, err
	}
	for _, item := range plan.Items {
		if err := w.Write([]string{
			fmt.Sprint(item.Rank),
			item.Bucket,
			item.SN,
			item.Manufacturer,
			item.Model,
			item.Environment,
			item.ConfigType,
			fmt.Sprint(item.CPULogicalCores),
			fmt.Sprintf("%.4f", item.StorageCapacityTB),
			fmt.Sprintf("%.4f", item.PSA),
			fmt.Sprintf("%.4f", item.ArchStandardizedFactor),
			fmt.Sprintf("%.4f", item.BaseScore),
			fmt.Sprintf("%.4f", item.AFROld),
			fmt.Sprintf("%.4f", item.AFRAvg),
			fmt.Sprintf("%.6f", item.FailureAdjustFactor),
			fmt.Sprintf("%.4f", item.FinalScore),
			item.SpecialPolicy,
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf, nil
}

func buildXLSX(plan domain.RenewalPlan) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)
	if err := f.SetSheetRow(sheet, "A1", &[]string{"plan_id", "target_date", "target_cores", "warm_target_storage_tb", "hot_target_storage_tb", "unmatched_config_count", "unmatched_config_types", "selected_cores", "selected_storage_tb", "selected_count"}); err != nil {
		return nil, err
	}
	if err := f.SetSheetRow(sheet, "A2", &[]any{plan.PlanID, plan.TargetDate, plan.TargetCores, plan.WarmTargetStorageTB, plan.HotTargetStorageTB, plan.UnmatchedConfigCount, strings.Join(plan.UnmatchedConfigTypes, "|"), plan.SelectedCores, plan.SelectedStorageTB, plan.SelectedCount}); err != nil {
		return nil, err
	}
	if err := f.SetSheetRow(sheet, "A4", &[]string{"rank", "bucket", "sn", "manufacturer", "model", "environment", "config_type", "cpu_logical_cores", "storage_capacity_tb", "psa", "arch_standardized_factor", "base_score", "afr_old", "afr_avg", "failure_adjust_factor", "final_score", "special_policy"}); err != nil {
		return nil, err
	}
	for i, item := range plan.Items {
		cell, _ := excelize.CoordinatesToCellName(1, i+5)
		if err := f.SetSheetRow(sheet, cell, &[]any{item.Rank, item.Bucket, item.SN, item.Manufacturer, item.Model, item.Environment, item.ConfigType, item.CPULogicalCores, item.StorageCapacityTB, item.PSA, item.ArchStandardizedFactor, item.BaseScore, item.AFROld, item.AFRAvg, item.FailureAdjustFactor, item.FinalScore, item.SpecialPolicy}); err != nil {
			return nil, err
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf, nil
}
