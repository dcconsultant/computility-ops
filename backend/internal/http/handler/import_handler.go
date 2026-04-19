package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/diagnose"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ImportHandler struct {
	service *service.ImportService
}

type serverPackageStandardizedItem struct {
	SN                  string `json:"sn"`
	Manufacturer        string `json:"manufacturer"`
	Model               string `json:"model"`
	PSA                 string `json:"psa"`
	IDC                 string `json:"idc,omitempty"`
	Environment         string `json:"environment,omitempty"`
	ConfigType          string `json:"config_type"`
	ConfigTypeStandardized string `json:"config_type_standardized"`
	PackageMatched      bool   `json:"package_standardized_matched"`
	WarrantyEndDate     string `json:"warranty_end_date,omitempty"`
	LaunchDate          string `json:"launch_date,omitempty"`
}

func NewImportHandler(s *service.ImportService) *ImportHandler { return &ImportHandler{service: s} }

func (h *ImportHandler) failImport(c *gin.Context, action string, err error) {
	requestID, _ := c.Get("request_id")
	rid, _ := requestID.(string)
	diagnose.RecordImportError(action, rid, err)
	fail(c, 50001, fmt.Sprintf("导入失败：%s", diagnose.AnalyzeReason(err.Error())))
}

func (h *ImportHandler) ImportServers(c *gin.Context) {
	c.Set("audit_action", "servers.import")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, serviceServerHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "sn", "psa", "config_type"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	mapped := mapRows(headers, rows)
	result, err := h.service.ValidateAndReplaceServers(c.Request.Context(), mapped)
	if err != nil {
		h.failImport(c, "servers.import", err)
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ListServers(c *gin.Context) {
	c.Set("audit_action", "servers.list")
	rows, err := h.buildServerPackageStandardizedRows(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ExportServerPackageAnomalies(c *gin.Context) {
	c.Set("audit_action", "servers.package_anomaly.export")
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "xlsx")))
	if format != "xlsx" && format != "csv" {
		fail(c, 40001, "format must be xlsx or csv")
		return
	}
	rows, err := h.buildServerPackageStandardizedRows(c.Request.Context())
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	anomaly := make([]serverPackageStandardizedItem, 0, len(rows))
	for _, r := range rows {
		if !r.PackageMatched {
			anomaly = append(anomaly, r)
		}
	}
	exportRows := anomaly
	if len(exportRows) == 0 {
		// 兜底：避免仅表头，输出全量并让使用者核查标准化状态。
		exportRows = rows
	}
	filename := fmt.Sprintf("server-package-anomaly-%s.%s", time.Now().Format("20060102-150405"), format)
	if format == "csv" {
		buf := &bytes.Buffer{}
		w := csv.NewWriter(buf)
		header := []string{"SN", "制造商", "服务器型号", "PSA", "机房", "环境", "配置类型", "配置类型标准化", "保修结束日期", "投产日期"}
		if err := w.Write(header); err != nil {
			fail(c, 50001, "导出失败")
			return
		}
		for _, r := range exportRows {
			record := []string{r.SN, r.Manufacturer, r.Model, r.PSA, r.IDC, r.Environment, r.ConfigType, r.ConfigTypeStandardized, r.WarrantyEndDate, r.LaunchDate}
			if err := w.Write(record); err != nil {
				fail(c, 50001, "导出失败")
				return
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			fail(c, 50001, "导出失败")
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
		return
	}

	xf := excelize.NewFile()
	sheet := xf.GetSheetName(0)
	header := []string{"SN", "制造商", "服务器型号", "PSA", "机房", "环境", "配置类型", "配置类型标准化", "保修结束日期", "投产日期"}
	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	for idx, r := range exportRows {
		row := idx + 2
		_ = xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.SN)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.Manufacturer)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.Model)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.PSA)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.IDC)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("F%d", row), r.Environment)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("G%d", row), r.ConfigType)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("H%d", row), r.ConfigTypeStandardized)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("I%d", row), r.WarrantyEndDate)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("J%d", row), r.LaunchDate)
	}
	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *ImportHandler) ImportHostPackages(c *gin.Context) {
	c.Set("audit_action", "host_packages.import")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, serviceHostPackageHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "config_type", "cpu_logical_cores", "arch_standardized_factor", "data_disk_count", "server_value_score"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	result, err := h.service.ValidateAndReplaceHostPackages(c.Request.Context(), mapRows(headers, rows))
	if err != nil {
		h.failImport(c, "host_packages.import", err)
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ListHostPackages(c *gin.Context) {
	c.Set("audit_action", "host_packages.list")
	rows, err := h.service.ListHostPackages(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ImportSpecialRules(c *gin.Context) {
	c.Set("audit_action", "special_rules.import")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, serviceSpecialHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "sn", "policy"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	result, err := h.service.ValidateAndReplaceSpecialRules(c.Request.Context(), mapRows(headers, rows))
	if err != nil {
		h.failImport(c, "special_rules.import", err)
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ListSpecialRules(c *gin.Context) {
	c.Set("audit_action", "special_rules.list")
	rows, err := h.service.ListSpecialRules(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ImportModelFailureRates(c *gin.Context) {
	c.Set("audit_action", "failure_model.import")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, serviceModelFailureHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "manufacturer", "model", "failure_rate"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	result, err := h.service.ValidateAndReplaceModelFailureRates(c.Request.Context(), mapRows(headers, rows))
	if err != nil {
		h.failImport(c, "failure_model.import", err)
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ListModelFailureRates(c *gin.Context) {
	c.Set("audit_action", "failure_model.list")
	rows, err := h.service.ListModelFailureRates(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ImportPackageFailureRates(c *gin.Context) {
	c.Set("audit_action", "failure_package.import")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, servicePackageFailureHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "config_type", "failure_rate"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	result, err := h.service.ValidateAndReplacePackageFailureRates(c.Request.Context(), mapRows(headers, rows))
	if err != nil {
		h.failImport(c, "failure_package.import", err)
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ListPackageFailureRates(c *gin.Context) {
	c.Set("audit_action", "failure_package.list")
	rows, err := h.service.ListPackageFailureRates(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ImportPackageModelFailureRates(c *gin.Context) {
	c.Set("audit_action", "failure_package_model.import")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, servicePackageModelFailureHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "config_type", "manufacturer", "model", "failure_rate"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	result, err := h.service.ValidateAndReplacePackageModelFailureRates(c.Request.Context(), mapRows(headers, rows))
	if err != nil {
		h.failImport(c, "failure_package_model.import", err)
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ListPackageModelFailureRates(c *gin.Context) {
	c.Set("audit_action", "failure_package_model.list")
	rows, err := h.service.ListPackageModelFailureRates(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ListOverallFailureRates(c *gin.Context) {
	c.Set("audit_action", "failure_rates.overall.list")
	rows, err := h.service.ListOverallFailureRates(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ListFailureOverviewCards(c *gin.Context) {
	c.Set("audit_action", "failure_rates.overview_cards.list")
	rows, err := h.service.ListFailureOverviewCards(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ListFailureAgeTrendPoints(c *gin.Context) {
	c.Set("audit_action", "failure_rates.age_trend.list")
	rows, err := h.service.ListFailureAgeTrendPoints(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ListFailureFeatureFacts(c *gin.Context) {
	c.Set("audit_action", "failure_rates.features.list")
	rows, err := h.service.ListFailureFeatureFacts(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ListStorageTopServerRates(c *gin.Context) {
	c.Set("audit_action", "failure_rates.storage_top_servers.list")
	bucket := parseStorageTopBucket(c.Query("bucket"))
	rows, err := h.service.ListStorageTopServerRatesByBucket(c.Request.Context(), bucket)
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ExportWarmStorageServers(c *gin.Context) {
	c.Set("audit_action", "failure_rates.storage_top_servers.export")
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "xlsx")))
	if format != "xlsx" && format != "csv" {
		fail(c, 40001, "format must be xlsx or csv")
		return
	}
	bucket := parseStorageTopBucket(c.Query("bucket"))
	rows, err := h.service.ListStorageServerRatesByBucketForExport(c.Request.Context(), bucket)
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filenamePrefix := "warm-storage-servers"
	if bucket == "hot_storage" {
		filenamePrefix = "hot-storage-servers"
	}
	filename := fmt.Sprintf("%s-%s.%s", filenamePrefix, time.Now().Format("20060102-150405"), format)
	if format == "csv" {
		buf := &bytes.Buffer{}
		w := csv.NewWriter(buf)
		header := []string{"SN", "厂商", "型号", "配置类型", "环境", "机房", "保修截止日期", "数据盘数", "单盘容量(TB)", "单台总容量(TB)", "最近1年故障次数", "分母(1+盘数)", "故障率"}
		if err := w.Write(header); err != nil {
			fail(c, 50001, "导出失败")
			return
		}
		for _, r := range rows {
			record := []string{
				r.SN,
				r.Manufacturer,
				r.Model,
				r.ConfigType,
				r.Environment,
				r.IDC,
				r.WarrantyEndDate,
				strconv.Itoa(r.DataDiskCount),
				fmt.Sprintf("%.4f", r.SingleDiskCapacityTB),
				fmt.Sprintf("%.4f", r.TotalCapacityTB),
				strconv.Itoa(r.FaultCount),
				fmt.Sprintf("%.4f", r.Denominator),
				fmt.Sprintf("%.8f", r.FaultRate),
			}
			if err := w.Write(record); err != nil {
				fail(c, 50001, "导出失败")
				return
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			fail(c, 50001, "导出失败")
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
		return
	}

	xf := excelize.NewFile()
	sheet := xf.GetSheetName(0)
	header := []string{"SN", "厂商", "型号", "配置类型", "环境", "机房", "保修截止日期", "数据盘数", "单盘容量(TB)", "单台总容量(TB)", "最近1年故障次数", "分母(1+盘数)", "故障率"}
	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	for idx, r := range rows {
		row := idx + 2
		_ = xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.SN)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.Manufacturer)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.Model)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.ConfigType)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.Environment)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("F%d", row), r.IDC)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("G%d", row), r.WarrantyEndDate)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("H%d", row), r.DataDiskCount)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("I%d", row), r.SingleDiskCapacityTB)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("J%d", row), r.TotalCapacityTB)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("K%d", row), r.FaultCount)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("L%d", row), r.Denominator)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("M%d", row), r.FaultRate)
	}
	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *ImportHandler) AnalyzeFaultRates(c *gin.Context) {
	c.Set("audit_action", "failure_rates.analyze")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, serviceFaultListHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "sn"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	excludeOverWarranty := parseBoolLoose(c.PostForm("exclude_over_warranty"))
	result, err := h.service.AnalyzeFaultRates(c.Request.Context(), mapRows(headers, rows), excludeOverWarranty)
	if err != nil {
		h.failImport(c, "failure_rates.analyze", err)
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ExportYearFaultAnalysis(c *gin.Context) {
	c.Set("audit_action", "failure_rates.year_fault_analysis.export")
	var req ExportYearFaultAnalysisReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数有误，请检查后重试")
		return
	}
	if len(req.Rows) == 0 {
		fail(c, 40001, "rows 不能为空")
		return
	}
	year := req.Year
	if year <= 0 {
		year = time.Now().Year()
	}

	xf := excelize.NewFile()
	sheet := xf.GetSheetName(0)
	header := []string{"行号", "SN", "创建时间", "范围", "分类", "命中", "备注"}
	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xf.SetCellValue(sheet, cell, h)
	}
	for idx, r := range req.Rows {
		row := idx + 2
		_ = xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.RowNo)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.SN)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.CreatedAt)
		_ = xf.SetCellValue(sheet, fmt.Sprintf("D%d", row), scopeLabelCN(r.Scope))
		_ = xf.SetCellValue(sheet, fmt.Sprintf("E%d", row), segmentLabelCN(r.Segment))
		if r.Matched {
			_ = xf.SetCellValue(sheet, fmt.Sprintf("F%d", row), "是")
		} else {
			_ = xf.SetCellValue(sheet, fmt.Sprintf("F%d", row), "否")
		}
		_ = xf.SetCellValue(sheet, fmt.Sprintf("G%d", row), r.Remark)
	}
	buf, err := xf.WriteToBuffer()
	if err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	filename := fmt.Sprintf("year-fault-analysis-%d-%s.xlsx", year, time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *ImportHandler) readRows(c *gin.Context) ([]string, [][]string, bool) {
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传 Excel 文件")
		return nil, nil, false
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		fail(c, 40002, "仅支持 .xlsx 文件")
		return nil, nil, false
	}
	f, err := file.Open()
	if err != nil {
		fail(c, 40003, "文件读取失败，请重试")
		return nil, nil, false
	}
	defer f.Close()
	xf, err := excelize.OpenReader(f)
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
	return rows[0], rows[1:], true
}

func parseBoolLoose(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func parseStorageTopBucket(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "hot_storage", "hotstorage", "hot":
		return "hot_storage"
	default:
		return "warm_storage"
	}
}

func normalizeConfigTypeKey(v string) string {
	n := strings.ToLower(strings.TrimSpace(v))
	n = strings.ReplaceAll(n, " ", "")
	n = strings.ReplaceAll(n, "_", "")
	n = strings.ReplaceAll(n, "-", "")
	return n
}

func (h *ImportHandler) buildServerPackageStandardizedRows(ctx context.Context) ([]serverPackageStandardizedItem, error) {
	servers, err := h.service.ListServers(ctx)
	if err != nil {
		return nil, err
	}
	packages, err := h.service.ListHostPackages(ctx)
	if err != nil {
		return nil, err
	}
	pkgSet := map[string]struct{}{}
	for _, p := range packages {
		n := normalizeConfigTypeKey(p.ConfigType)
		if n == "" {
			continue
		}
		pkgSet[n] = struct{}{}
	}
	out := make([]serverPackageStandardizedItem, 0, len(servers))
	for _, s := range servers {
		n := normalizeConfigTypeKey(s.ConfigType)
		_, matched := pkgSet[n]
		status := "否"
		if matched {
			status = "是"
		}
		out = append(out, serverPackageStandardizedItem{
			SN:                     s.SN,
			Manufacturer:           s.Manufacturer,
			Model:                  s.Model,
			PSA:                    s.PSA,
			IDC:                    s.IDC,
			Environment:            s.Environment,
			ConfigType:             s.ConfigType,
			ConfigTypeStandardized: status,
			PackageMatched:         matched,
			WarrantyEndDate:        s.WarrantyEndDate,
			LaunchDate:             s.LaunchDate,
		})
	}
	return out, nil
}

func scopeLabelCN(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "all":
		return "整体"
	case "product":
		return "生产"
	case "devtest":
		return "开测"
	default:
		return strings.TrimSpace(v)
	}
}

func segmentLabelCN(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "storage":
		return "存储"
	case "non_storage":
		return "非存储"
	default:
		return strings.TrimSpace(v)
	}
}

func mapRows(headers []string, rows [][]string) []map[string]string {
	out := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		m := map[string]string{}
		for i, h := range headers {
			if i < len(row) {
				m[h] = strings.TrimSpace(row[i])
			} else {
				m[h] = ""
			}
		}
		out = append(out, m)
	}
	return out
}

func serviceServerHeaderMap() map[string]string {
	return map[string]string{"sn": "sn", "序列号": "sn", "制造商": "manufacturer", "厂商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "服务器型号": "model", "model": "model", "psa": "psa", "机房": "idc", "idc": "idc", "环境": "environment", "env": "environment", "environment": "environment", "配置类型": "config_type", "套餐": "config_type", "configtype": "config_type", "保修结束日期": "warranty_end_date", "保修截止日期": "warranty_end_date", "warrantyenddate": "warranty_end_date", "投产日期": "launch_date", "launchdate": "launch_date"}
}
func serviceHostPackageHeaderMap() map[string]string {
	return map[string]string{"配置类型": "config_type", "套餐": "config_type", "configtype": "config_type", "场景大类": "scene_category", "scenecategory": "scene_category", "cpu逻辑核数": "cpu_logical_cores", "cpulogicalcores": "cpu_logical_cores", "gpu卡数": "gpu_card_count", "卡数": "gpu_card_count", "gpu_card_count": "gpu_card_count", "gpucardcount": "gpu_card_count", "数据盘类型": "data_disk_type", "数据盘种类": "data_disk_type", "datadisktype": "data_disk_type", "磁盘类型": "data_disk_type", "disktype": "data_disk_type", "数据盘数量": "data_disk_count", "datadiskcount": "data_disk_count", "存储容量(tb)": "storage_capacity_tb", "存储容量": "storage_capacity_tb", "storagecapacitytb": "storage_capacity_tb", "服务器价值分": "server_value_score", "价值分": "server_value_score", "servervaluescore": "server_value_score", "架构标准化系数": "arch_standardized_factor", "archstandardizedfactor": "arch_standardized_factor"}
}
func serviceSpecialHeaderMap() map[string]string {
	return map[string]string{"sn": "sn", "序列号": "sn", "制造商": "manufacturer", "厂商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "model": "model", "psa": "psa", "机房": "idc", "idc": "idc", "套餐": "package_type", "配置类型": "package_type", "保修结束日期": "warranty_end_date", "投产日期": "launch_date", "策略": "policy", "标签": "policy", "黑白": "policy", "原因": "reason", "备注": "reason", "reason": "reason"}
}
func serviceModelFailureHeaderMap() map[string]string {
	return map[string]string{"厂商": "manufacturer", "制造商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "服务器型号": "model", "model": "model", "故障率": "failure_rate", "failurerate": "failure_rate", "过保故障率": "over_warranty_failure_rate", "overwarrantyfailurerate": "over_warranty_failure_rate"}
}
func servicePackageFailureHeaderMap() map[string]string {
	return map[string]string{"配置类型": "config_type", "套餐": "config_type", "configtype": "config_type", "故障率": "failure_rate", "failurerate": "failure_rate", "过保故障率": "over_warranty_failure_rate", "overwarrantyfailurerate": "over_warranty_failure_rate"}
}
func servicePackageModelFailureHeaderMap() map[string]string {
	return map[string]string{"套餐": "config_type", "配置类型": "config_type", "configtype": "config_type", "厂商": "manufacturer", "制造商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "服务器型号": "model", "model": "model", "故障率": "failure_rate", "failurerate": "failure_rate", "过保故障率": "over_warranty_failure_rate", "overwarrantyfailurerate": "over_warranty_failure_rate"}
}
func serviceFaultListHeaderMap() map[string]string {
	return map[string]string{"类型": "type", "主机名": "hostname", "业务": "business", "机房": "idc", "机柜": "rack", "厂商": "manufacturer", "制造商": "manufacturer", "型号": "model", "sn": "sn", "序列号": "sn", "ip": "ip", "ipmi": "ipmi", "过保日期": "warranty_end_date", "上报故障": "reported_fault", "故障描述": "fault_desc", "故障来源": "fault_source", "业务对接人": "business_owner", "处理环节": "process_stage", "工单状态": "ticket_status", "真实故障": "real_fault", "创建时间": "created_at", "提单人": "creator", "更新时间": "updated_at", "结束时间": "ended_at", "工单链接": "ticket_link", "日志链接": "log_link"}
}
