package handler

import (
	"fmt"
	"net/http"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
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

func (h *ValueScoreSetupHandler) CalculateMonthlyTCO(c *gin.Context) {
	c.Set("audit_action", "value_score.tco.calculate")
	var req domain.ValueScoreTCOCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	v, err := h.svc.CalculateMonthlyTCO(c.Request.Context(), req)
	if err != nil {
		fail(c, 50001, "计算失败")
		return
	}
	ok(c, v)
}

func (h *ValueScoreSetupHandler) ExportMonthlyTCO(c *gin.Context) {
	c.Set("audit_action", "value_score.tco.export")
	res, err := h.svc.CalculateMonthlyTCO(c.Request.Context(), domain.ValueScoreTCOCalculateRequest{})
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}

	xf := excelize.NewFile()
	sheet := "TCO"
	xf.SetSheetName("Sheet1", sheet)
	_ = xf.SetCellValue(sheet, "A1", "目标机房")
	_ = xf.SetCellValue(sheet, "B1", res.IDC)
	_ = xf.SetCellValue(sheet, "A2", "机柜利用率")
	_ = xf.SetCellValue(sheet, "B2", res.CabinetUtilization)
	_ = xf.SetCellValue(sheet, "A3", "最低额定功率(KW)")
	_ = xf.SetCellValue(sheet, "B3", res.MinRatedPowerKW)
	_ = xf.SetCellValue(sheet, "A4", "机柜月租(CNY)")
	_ = xf.SetCellValue(sheet, "B4", res.MonthlyRentCNY)
	_ = xf.SetCellValue(sheet, "A5", "折旧月数")
	_ = xf.SetCellValue(sheet, "B5", res.DepreciationMonths)
	_ = xf.SetCellValue(sheet, "A6", "公式")
	_ = xf.SetCellValue(sheet, "B6", res.Formula)

	headers := []string{"配置类型", "功率(W)", "功率(KW)", "机柜费/月", "折旧/月", "月TCO"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 8)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	for idx, item := range res.Items {
		row := idx + 9
		_ = xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.ConfigType)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.PowerWatts)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.PowerKW)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.CabinetCostMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.DepreciationMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("F%d", row), item.TotalTCOMonthly)
	}

	if res.Note != "" {
		_ = xf.SetCellValue(sheet, "A7", "备注")
		_ = xf.SetCellValue(sheet, "B7", res.Note)
	}

	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filename := "value-score-monthly-tco.xlsx"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

