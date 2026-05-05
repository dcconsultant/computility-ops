package handler

import (
	"fmt"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type CabinetHandler struct {
	service *service.CabinetService
}

func NewCabinetHandler(s *service.CabinetService) *CabinetHandler { return &CabinetHandler{service: s} }

func (h *CabinetHandler) GetUtilization(c *gin.Context) {
	c.Set("audit_action", "cabinet.utilization.get")
	data, err := h.service.GetUtilization(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, data)
}

func (h *CabinetHandler) UpdateUtilization(c *gin.Context) {
	c.Set("audit_action", "cabinet.utilization.update")
	var req struct {
		Utilization float64 `json:"utilization" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	data, err := h.service.UpdateUtilization(c.Request.Context(), req.Utilization)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, data)
}

func (h *CabinetHandler) List(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.list")
	rows, err := h.service.ListCabinetConfigs(c.Request.Context())
	if err != nil {
		fail(c, 50001, "查询失败")
		return
	}
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *CabinetHandler) Create(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.create")
	var req domain.CabinetConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	req.IDC = strings.TrimSpace(req.IDC)
	row, err := h.service.CreateCabinetConfig(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			fail(c, 40001, "机房+额定功率已存在")
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, row)
}

func (h *CabinetHandler) Update(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.update")
	var req domain.CabinetConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	id := atoiDefault(c.Param("id"), 0)
	req.ID = int64(id)
	req.IDC = strings.TrimSpace(req.IDC)
	row, err := h.service.UpdateCabinetConfig(c.Request.Context(), req)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "duplicate") {
			fail(c, 40001, "机房+额定功率已存在")
			return
		}
		if strings.Contains(lower, "not found") {
			fail(c, 40401, "记录不存在")
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, row)
}

func (h *CabinetHandler) Import(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.import")
	headers, rows, okRead := readSimpleXlsxRows(c)
	if !okRead {
		return
	}
	mapped := mapCabinetRows(headers, rows)
	res, err := h.service.ImportCabinetConfigs(c.Request.Context(), mapped)
	if err != nil {
		fail(c, 50001, "导入失败")
		return
	}
	ok(c, res)
}

func readSimpleXlsxRows(c *gin.Context) ([]string, [][]string, bool) {
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传 file")
		return nil, nil, false
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		fail(c, 40002, "仅支持 .xlsx 文件")
		return nil, nil, false
	}
	f, err := file.Open()
	if err != nil {
		fail(c, 50001, "读取文件失败")
		return nil, nil, false
	}
	defer f.Close()
	xf, err := excelize.OpenReader(f)
	if err != nil {
		fail(c, 40003, "文件格式无效，请确认是标准 .xlsx")
		return nil, nil, false
	}
	sheet := xf.GetSheetName(0)
	if sheet == "" {
		fail(c, 40001, "空文件")
		return nil, nil, false
	}
	all, err := xf.GetRows(sheet)
	if err != nil || len(all) == 0 {
		fail(c, 40001, "空文件")
		return nil, nil, false
	}
	headers := all[0]
	if len(headers) == 0 {
		fail(c, 40001, "缺少表头")
		return nil, nil, false
	}
	return headers, all[1:], true
}

func mapCabinetRows(headers []string, rows [][]string) []map[string]string {
	out := make([]map[string]string, 0, len(rows))
	norm := make([]string, len(headers))
	for i, h := range headers {
		norm[i] = normalizeCabinetHeaderName(h)
	}
	for _, r := range rows {
		item := make(map[string]string, len(headers))
		for i := range headers {
			v := ""
			if i < len(r) {
				v = strings.TrimSpace(r[i])
			}
			item[norm[i]] = v
		}
		out = append(out, item)
	}
	return out
}

func normalizeCabinetHeaderName(raw string) string {
	n := strings.TrimSpace(strings.ToLower(raw))
	n = strings.ReplaceAll(n, " ", "")
	n = strings.ReplaceAll(n, "-", "")
	n = strings.ReplaceAll(n, "_", "")
	return n
}

func (h *CabinetHandler) ExportTemplate(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.template.export")
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"机房", "额定功率(KW)", "机柜月租(CNY)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	_ = f.SetCellValue(sheet, "A2", "IDC-SG-01")
	_ = f.SetCellValue(sheet, "B2", 10)
	_ = f.SetCellValue(sheet, "C2", 3500)

	filename := fmt.Sprintf("cabinet-import-template-%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := f.Write(c.Writer); err != nil {
		fail(c, 50001, "导出失败")
		return
	}
	c.Set("audit_result", "ok")
}

func (h *CabinetHandler) Delete(c *gin.Context) {
	c.Set("audit_action", "cabinet.config.delete")
	id := atoiDefault(c.Param("id"), 0)
	if id <= 0 {
		fail(c, 40001, "id无效")
		return
	}
	if err := h.service.DeleteCabinetConfig(c.Request.Context(), int64(id)); err != nil {
		fail(c, 50001, "删除失败")
		return
	}
	ok(c, gin.H{"deleted": true, "id": id})
}
