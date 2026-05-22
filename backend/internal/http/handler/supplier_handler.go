package handler

import (
	"fmt"
	"strings"
	"time"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type SupplierHandler struct {
	service *service.SupplierService
}

func NewSupplierHandler(s *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{service: s}
}

func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.create")
	var req UpsertSupplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查供应商字段")
		return
	}
	item, err := h.service.CreateSupplier(c.Request.Context(), service.UpsertSupplierInput(req))
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, item)
}

func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.update")
	var req UpsertSupplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查供应商字段")
		return
	}
	item, err := h.service.UpdateSupplier(c.Request.Context(), c.Param("supplier_id"), service.UpsertSupplierInput(req))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, item)
}

func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.get")
	item, err := h.service.GetSupplier(c.Request.Context(), c.Param("supplier_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, item)
}

func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	c.Set("audit_action", "suppliers.list")
	list, err := h.service.ListSuppliers(c.Request.Context(), c.Query("q"))
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": len(list), "page": 1, "page_size": len(list)})
}

func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	c.Set("audit_action", "suppliers.delete")
	if err := h.service.DeleteSupplier(c.Request.Context(), c.Param("supplier_id")); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") {
			fail(c, 40401, err.Error())
			return
		}
		if strings.Contains(msg, "referenced by contract") {
			fail(c, 40001, "该供应商已被合同引用，请先解除合同关联后再删除")
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "supplier_id": c.Param("supplier_id")})
}

func (h *SupplierHandler) ExportSuppliers(c *gin.Context) {
	c.Set("audit_action", "suppliers.export")
	list, err := h.service.ListSuppliers(c.Request.Context(), c.Query("q"))
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"公司全名", "税号", "项目负责人", "项目负责人电话", "技术接口人", "技术接口人电话", "业务范围"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for i, item := range list {
		row := i + 2
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.CompanyFullName)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.TaxNumber)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.ProjectOwner)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.ProjectOwnerPhone)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.TechContact)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item.TechContactPhone)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item.BusinessScope)
	}
	_ = f.SetColWidth(sheet, "A", "A", 34)
	_ = f.SetColWidth(sheet, "B", "B", 24)
	_ = f.SetColWidth(sheet, "C", "F", 18)
	_ = f.SetColWidth(sheet, "G", "G", 48)

	filename := fmt.Sprintf("suppliers-%s.xlsx", time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := f.Write(c.Writer); err != nil {
		fail(c, 50001, err.Error())
		return
	}
}

func (h *SupplierHandler) ExportSupplierTemplate(c *gin.Context) {
	c.Set("audit_action", "suppliers.template_export")
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"公司全名", "税号", "项目负责人", "项目负责人电话", "技术接口人", "技术接口人电话", "业务范围"}
	example := []string{"示例科技有限公司", "91310000MA1K12345X", "张三", "13800138000", "李四", "021-12345678", "云资源采购、服务器维保、IDC运维"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for i, v := range example {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheet, cell, v)
	}
	_ = f.SetColWidth(sheet, "A", "A", 34)
	_ = f.SetColWidth(sheet, "B", "B", 24)
	_ = f.SetColWidth(sheet, "C", "F", 18)
	_ = f.SetColWidth(sheet, "G", "G", 48)

	filename := fmt.Sprintf("supplier-import-template-%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := f.Write(c.Writer); err != nil {
		fail(c, 50001, err.Error())
		return
	}
}

func (h *SupplierHandler) ImportSuppliers(c *gin.Context) {
	c.Set("audit_action", "suppliers.import")
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传 Excel 文件")
		return
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		fail(c, 40002, "仅支持 .xlsx 文件")
		return
	}
	fp, err := file.Open()
	if err != nil {
		fail(c, 40003, "文件读取失败")
		return
	}
	defer fp.Close()

	xf, err := excelize.OpenReader(fp)
	if err != nil {
		fail(c, 40003, "文件格式无效，请确认是标准 .xlsx")
		return
	}
	sheet := xf.GetSheetName(0)
	rows, err := xf.GetRows(sheet)
	if err != nil {
		fail(c, 40003, "读取 Excel 行失败")
		return
	}
	inputs := make([]service.UpsertSupplierInput, 0)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isEmptySupplierRow(row) {
			continue
		}
		inputs = append(inputs, service.UpsertSupplierInput{
			CompanyFullName: cell(row, 0),
			TaxNumber:       cell(row, 1),
			ProjectOwner:    cell(row, 2),
			ProjectOwnerPhone: cell(row,
				3),
			TechContact:      cell(row, 4),
			TechContactPhone: cell(row, 5),
			BusinessScope:    cell(row, 6),
		})
	}
	count, err := h.service.ImportSuppliers(c.Request.Context(), inputs)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, gin.H{"imported": count})
}

func cell(row []string, idx int) string {
	if idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func isEmptySupplierRow(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
