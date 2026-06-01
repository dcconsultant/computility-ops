package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (h *ValueScoreSetupHandler) GetCostParams(c *gin.Context) {
	c.Set("audit_action", "value_score.cost_params.get")
	v, err := h.svc.GetCostParams(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, v)
}

func (h *ValueScoreSetupHandler) UpdateCostParams(c *gin.Context) {
	c.Set("audit_action", "value_score.cost_params.update")
	var req domain.ValueScoreCostParams
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	v, err := h.svc.UpdateCostParams(c.Request.Context(), req)
	if err != nil {
		fail(c, 40004, err.Error())
		return
	}
	ok(c, v)
}

func (h *ValueScoreSetupHandler) ListOriginalValues(c *gin.Context) {
	c.Set("audit_action", "value_score.original_values.list")
	rows, err := h.svc.ListOriginalValues(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows})
}

func (h *ValueScoreSetupHandler) ImportOriginalValues(c *gin.Context) {
	c.Set("audit_action", "value_score.original_values.import")
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传文件")
		return
	}
	defer file.Close()
	xf, err := excelize.OpenReader(file)
	if err != nil {
		fail(c, 40003, "文件格式无效，请确认是标准 .xlsx")
		return
	}
	defer func() { _ = xf.Close() }()
	sheets := xf.GetSheetList()
	if len(sheets) == 0 {
		fail(c, 40003, "Excel 中没有可用工作表")
		return
	}
	raw, err := xf.GetRows(sheets[0])
	if err != nil || len(raw) < 2 {
		fail(c, 40003, "模板至少需要表头+一行数据")
		return
	}
	headers := raw[0]
	idxConfig, idxOriginal := -1, -1
	for i, h := range headers {
		n := strings.TrimSpace(strings.ToLower(h))
		if n == "配置类型" || n == "config_type" {
			idxConfig = i
		}
		if n == "原值(cny)" || n == "原值（cny）" || n == "server_original_cny" || n == "原值" {
			idxOriginal = i
		}
	}
	if idxConfig < 0 || idxOriginal < 0 {
		fail(c, 40004, "模板缺少必填列：配置类型、原值(CNY)")
		return
	}
	rows := make([]domain.ValueScoreOriginalValue, 0, len(raw)-1)
	for i := 1; i < len(raw); i++ {
		row := raw[i]
		if idxConfig >= len(row) && idxOriginal >= len(row) {
			continue
		}
		cfg := ""
		if idxConfig < len(row) {
			cfg = strings.TrimSpace(row[idxConfig])
		}
		if cfg == "" {
			continue
		}
		v := "0"
		if idxOriginal < len(row) {
			v = strings.TrimSpace(row[idxOriginal])
			if v == "" {
				v = "0"
			}
		}
		num, e := strconv.ParseFloat(v, 64)
		if e != nil {
			fail(c, 40004, fmt.Sprintf("第%d行原值(CNY)必须是数字", i+1))
			return
		}
		rows = append(rows, domain.ValueScoreOriginalValue{ConfigType: cfg, ServerOriginalCNY: num})
	}
	if err := h.svc.ReplaceOriginalValues(c.Request.Context(), rows); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	ok(c, gin.H{"imported": len(rows)})
}

func (h *ValueScoreSetupHandler) ExportOriginalValuesTemplate(c *gin.Context) {
	c.Set("audit_action", "value_score.original_values.template.export")
	xf := excelize.NewFile()
	sheet := xf.GetSheetName(0)
	headers := []string{"配置类型", "原值(CNY)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	_ = xf.SetCellValue(sheet, "A2", "")
	_ = xf.SetCellValue(sheet, "B2", 0)
	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filename := fmt.Sprintf("value-score-original-value-template-%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
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

func (h *ValueScoreSetupHandler) ImportCostParams(c *gin.Context) {
	c.Set("audit_action", "value_score.cost_params.import")
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传文件")
		return
	}
	defer file.Close()
	xf, err := excelize.OpenReader(file)
	if err != nil {
		fail(c, 40003, "文件格式无效，请确认是标准 .xlsx")
		return
	}
	defer func() { _ = xf.Close() }()
	sheets := xf.GetSheetList()
	if len(sheets) == 0 {
		fail(c, 40003, "Excel 中没有可用工作表")
		return
	}
	rows, err := xf.GetRows(sheets[0])
	if err != nil || len(rows) < 2 {
		fail(c, 40003, "模板至少需要表头+一行数据")
		return
	}
	headers := rows[0]
	values := rows[1]
	get := func(name string) string {
		for i, h := range headers {
			if strings.EqualFold(strings.TrimSpace(h), name) {
				if i < len(values) {
					return strings.TrimSpace(values[i])
				}
				return ""
			}
		}
		return ""
	}
	dep := 60
	if s := get("折旧月数"); s != "" {
		v, e := strconv.Atoi(s)
		if e != nil {
			fail(c, 40004, "折旧月数必须是整数")
			return
		}
		dep = v
	}
	network := 0.0
	if s := get("网络设备分摊成本(CNY/月)"); s != "" {
		v, e := strconv.ParseFloat(s, 64)
		if e != nil {
			fail(c, 40004, "网络设备分摊成本(CNY/月) 必须是数字")
			return
		}
		network = v
	}
	renewal := 0.0
	if s := get("服务器续保费(CNY/月)"); s != "" {
		v, e := strconv.ParseFloat(s, 64)
		if e != nil {
			fail(c, 40004, "服务器续保费(CNY/月) 必须是数字")
			return
		}
		renewal = v
	}
	cabinetUtilization := 1.0
	if s := get("机柜利用率"); s != "" {
		v, e := strconv.ParseFloat(s, 64)
		if e != nil {
			fail(c, 40004, "机柜利用率必须是数字")
			return
		}
		cabinetUtilization = v
	}
	ratedPower := 0.0
	if s := get("额定功率(KW)"); s != "" {
		v, e := strconv.ParseFloat(s, 64)
		if e != nil {
			fail(c, 40004, "额定功率(KW) 必须是数字")
			return
		}
		ratedPower = v
	}
	monthlyRent := 0.0
	if s := get("机柜月租(CNY)"); s != "" {
		v, e := strconv.ParseFloat(s, 64)
		if e != nil {
			fail(c, 40004, "机柜月租(CNY) 必须是数字")
			return
		}
		monthlyRent = v
	}
	out, err := h.svc.UpdateCostParams(c.Request.Context(), domain.ValueScoreCostParams{
		DepreciationMonths:    dep,
		NetworkDeviceShareCNY: network,
		ServerRenewalFeeCNY:   renewal,
		CabinetUtilization:    cabinetUtilization,
		RatedPowerKW:          ratedPower,
		MonthlyRentCNY:        monthlyRent,
	})
	if err != nil {
		fail(c, 40004, err.Error())
		return
	}
	ok(c, out)
}

func (h *ValueScoreSetupHandler) ExportCostParamsTemplate(c *gin.Context) {
	c.Set("audit_action", "value_score.cost_params.template.export")
	xf := excelize.NewFile()
	sheet := xf.GetSheetName(0)
	headers := []string{"折旧月数", "网络设备分摊成本(CNY/月)", "服务器续保费(CNY/月)", "机柜利用率", "额定功率(KW)", "机柜月租(CNY)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	_ = xf.SetCellValue(sheet, "A2", 60)
	_ = xf.SetCellValue(sheet, "B2", 0)
	_ = xf.SetCellValue(sheet, "C2", 0)
	_ = xf.SetCellValue(sheet, "D2", 1)
	_ = xf.SetCellValue(sheet, "E2", 4)
	_ = xf.SetCellValue(sheet, "F2", 1000)
	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filename := fmt.Sprintf("value-score-cost-params-template-%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *ValueScoreSetupHandler) readPerformanceRows(c *gin.Context) ([]string, [][]string, bool) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传文件")
		return nil, nil, false
	}
	file, err := fileHeader.Open()
	if err != nil {
		fail(c, 40003, "文件读取失败，请重试")
		return nil, nil, false
	}
	defer file.Close()
	xf, err := excelize.OpenReader(file)
	if err != nil {
		fail(c, 40003, "文件格式无效，请确认是标准 .xlsx")
		return nil, nil, false
	}
	defer func() { _ = xf.Close() }()
	sheets := xf.GetSheetList()
	if len(sheets) == 0 {
		fail(c, 40003, "Excel 中没有可用工作表")
		return nil, nil, false
	}
	rows, err := xf.GetRows(sheets[0])
	if err != nil || len(rows) == 0 {
		fail(c, 40003, "读取工作表失败或无数据")
		return nil, nil, false
	}
	headers := service.MapHeaders(rows[0], performanceHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "config_type", "unavailable_cores", "unavailable_memory_gb", "performance_score", "server_original_cny"); err != nil {
		fail(c, 40004, err.Error())
		return nil, nil, false
	}
	return headers, rows[1:], true
}

func performanceHeaderMap() map[string]string {
	return map[string]string{
		"配置类型":                  "config_type",
		"套餐":                    "config_type",
		"config_type":           "config_type",
		"不可用核数":                 "unavailable_cores",
		"不可用cpu核数":              "unavailable_cores",
		"unavailable_cores":     "unavailable_cores",
		"不可用内存容量(gb)":           "unavailable_memory_gb",
		"不可用内存容量（gb）":           "unavailable_memory_gb",
		"unavailable_memory_gb": "unavailable_memory_gb",
		"性能跑分":                  "performance_score",
		"performance_score":     "performance_score",
		"原值(cny)":               "server_original_cny",
		"原值（cny）":               "server_original_cny",
		"原值":                    "server_original_cny",
		"server_original_cny":   "server_original_cny",
	}
}

func (h *ValueScoreSetupHandler) ListPerformanceParams(c *gin.Context) {
	c.Set("audit_action", "value_score.performance_params.list")
	rows, err := h.svc.ListPerformanceParams(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows})
}

func (h *ValueScoreSetupHandler) ImportPerformanceParams(c *gin.Context) {
	c.Set("audit_action", "value_score.performance_params.import")
	headers, rows, okRead := h.readPerformanceRows(c)
	if !okRead {
		return
	}
	mapped := mapRows(headers, rows)
	res, err := h.svc.ImportUnifiedConfigParams(c.Request.Context(), mapped)
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, res)
}

func (h *ValueScoreSetupHandler) PreviewPerformanceParams(c *gin.Context) {
	c.Set("audit_action", "value_score.performance_params.preview")
	headers, rows, okRead := h.readPerformanceRows(c)
	if !okRead {
		return
	}
	mapped := mapRows(headers, rows)
	res, err := h.svc.PreviewPerformanceParams(c.Request.Context(), mapped)
	if err != nil {
		fail(c, 50001, "预检失败")
		return
	}
	ok(c, res)
}

func (h *ValueScoreSetupHandler) ExportPerformanceParamsTemplate(c *gin.Context) {
	c.Set("audit_action", "value_score.performance_params.template.export")
	xf := excelize.NewFile()
	sheet := xf.GetSheetName(0)
	headers := []string{"配置类型", "原值(CNY)", "不可用核数", "不可用内存容量(GB)", "性能跑分"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	_ = xf.SetCellValue(sheet, "A2", "compute-a")
	_ = xf.SetCellValue(sheet, "B2", 100000)
	_ = xf.SetCellValue(sheet, "C2", 0)
	_ = xf.SetCellValue(sheet, "D2", 0)
	_ = xf.SetCellValue(sheet, "E2", 2277)
	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filename := fmt.Sprintf("value-score-performance-template-%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *ValueScoreSetupHandler) CalculatePerformance(c *gin.Context) {
	c.Set("audit_action", "value_score.performance.calculate")
	v, err := h.svc.CalculatePerformance(c.Request.Context())
	if err != nil {
		fail(c, 50001, "计算失败")
		return
	}
	ok(c, v)
}

func (h *ValueScoreSetupHandler) ExportMonthlyTCO(c *gin.Context) {
	c.Set("audit_action", "value_score.tco.export")
	rows, err := h.svc.CalculateTableItems(c.Request.Context())
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}

	xf := excelize.NewFile()
	sheet := "价值分"
	xf.SetSheetName("Sheet1", sheet)

	headers := []string{
		"配置类型",
		"场景",
		"CPU逻辑核数",
		"内存容量(GB)",
		"存储容量(TB)",
		"GPU卡数",
		"不可用核数",
		"不可用内存(GB)",
		"性能跑分",
		"可用核数",
		"可用内存(GB)",
		"标准跑分",
		"CPU性能折算系数",
		"内存配比",
		"内存配比系数",
		"整体性能折算比",
		"功率(W)",
		"功率(KW)",
		"机柜费/月",
		"原值(CNY)",
		"折旧/月",
		"网络设备分摊/月",
		"网络机柜等分摊/月",
		"服务器续保费/月",
		"其他固定成本/月",
		"月TCO",
		"单位月TCO",
		"价值分v1",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	for idx, item := range rows {
		row := idx + 2
		_ = xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.ConfigType)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.SceneCategory)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.CPULogicalCores)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.MemoryCapacityGB)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.CapacityStorageTB)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("F%d", row), item.CountGPU)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("G%d", row), item.UnavailableCores)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("H%d", row), item.UnavailableMemoryGB)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("I%d", row), item.PerformanceScore)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("J%d", row), item.AvailableCores)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("K%d", row), item.AvailableMemoryGB)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("L%d", row), item.StandardScore)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("M%d", row), item.CPUPerformanceFactor)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("N%d", row), item.MemoryRatio)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("O%d", row), item.MemoryRatioFactor)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("P%d", row), item.OverallPerformanceRatio)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("Q%d", row), item.PowerWatts)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("R%d", row), item.PowerKW)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("S%d", row), item.CabinetCostMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("T%d", row), item.ServerOriginalCNY)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("U%d", row), item.DepreciationMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("V%d", row), item.NetworkDeviceMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("W%d", row), item.NetworkCabinetMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("X%d", row), item.ServerRenewalMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("Y%d", row), item.OtherFixedCostMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("Z%d", row), item.TotalTCOMonthly)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("AA%d", row), fmt.Sprintf("%.2f %s", item.UnitTCO, valueScoreUnitLabel(item.SceneType)))
		_ = xf.SetCellValue(sheet, fmt.Sprintf("AB%d", row), item.ValueScoreV1)
	}

	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filename := "value-score.xlsx"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func valueScoreUnitLabel(sceneType string) string {
	switch sceneType {
	case "gpu":
		return "/GPU卡"
	case "warm_storage", "hot_storage":
		return "/TB"
	default:
		return "/核"
	}
}
