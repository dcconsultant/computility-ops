package handler

import (
	"strings"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ImportHandler struct {
	service *service.ImportService
}

func NewImportHandler(s *service.ImportService) *ImportHandler { return &ImportHandler{service: s} }

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
		fail(c, 50001, "导入失败，请稍后重试")
		return
	}
	ok(c, result)
}

func (h *ImportHandler) ListServers(c *gin.Context) {
	c.Set("audit_action", "servers.list")
	rows, err := h.service.ListServers(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *ImportHandler) ImportHostPackages(c *gin.Context) {
	c.Set("audit_action", "host_packages.import")
	headers, rows, okRead := h.readRows(c)
	if !okRead {
		return
	}
	headers = service.MapHeaders(headers, serviceHostPackageHeaderMap())
	if err := service.ValidateRequiredHeaders(headers, "config_type", "cpu_logical_cores", "arch_standardized_factor"); err != nil {
		fail(c, 40004, err.Error())
		return
	}
	result, err := h.service.ValidateAndReplaceHostPackages(c.Request.Context(), mapRows(headers, rows))
	if err != nil {
		fail(c, 50001, "导入失败，请稍后重试")
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
		fail(c, 50001, "导入失败，请稍后重试")
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
		fail(c, 50001, "导入失败，请稍后重试")
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
		fail(c, 50001, "导入失败，请稍后重试")
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
		fail(c, 50001, "导入失败，请稍后重试")
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
	return map[string]string{"sn": "sn", "序列号": "sn", "制造商": "manufacturer", "厂商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "model": "model", "psa": "psa", "机房": "idc", "idc": "idc", "环境": "environment", "env": "environment", "environment": "environment", "配置类型": "config_type", "套餐": "config_type", "configtype": "config_type", "保修结束日期": "warranty_end_date", "保修截止日期": "warranty_end_date", "warrantyenddate": "warranty_end_date", "投产日期": "launch_date", "launchdate": "launch_date"}
}
func serviceHostPackageHeaderMap() map[string]string {
	return map[string]string{"配置类型": "config_type", "套餐": "config_type", "configtype": "config_type", "场景大类": "scene_category", "scenecategory": "scene_category", "cpu逻辑核数": "cpu_logical_cores", "cpulogicalcores": "cpu_logical_cores", "磁盘类型": "disk_type", "disktype": "disk_type", "存储容量(tb)": "storage_capacity_tb", "存储容量": "storage_capacity_tb", "storagecapacitytb": "storage_capacity_tb", "架构标准化系数": "arch_standardized_factor", "archstandardizedfactor": "arch_standardized_factor"}
}
func serviceSpecialHeaderMap() map[string]string {
	return map[string]string{"sn": "sn", "序列号": "sn", "制造商": "manufacturer", "厂商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "model": "model", "psa": "psa", "机房": "idc", "idc": "idc", "套餐": "package_type", "配置类型": "package_type", "保修结束日期": "warranty_end_date", "投产日期": "launch_date", "策略": "policy", "标签": "policy", "黑白": "policy"}
}
func serviceModelFailureHeaderMap() map[string]string {
	return map[string]string{"厂商": "manufacturer", "制造商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "model": "model", "故障率": "failure_rate", "failurerate": "failure_rate"}
}
func servicePackageFailureHeaderMap() map[string]string {
	return map[string]string{"配置类型": "config_type", "套餐": "config_type", "configtype": "config_type", "故障率": "failure_rate", "failurerate": "failure_rate"}
}
func servicePackageModelFailureHeaderMap() map[string]string {
	return map[string]string{"套餐": "config_type", "配置类型": "config_type", "configtype": "config_type", "厂商": "manufacturer", "制造商": "manufacturer", "manufacturer": "manufacturer", "型号": "model", "model": "model", "故障率": "failure_rate", "failurerate": "failure_rate"}
}
