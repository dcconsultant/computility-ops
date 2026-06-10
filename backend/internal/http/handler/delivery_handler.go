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

type DeliveryHandler struct {
	service *service.DeliveryService
}

func NewDeliveryHandler(s *service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{service: s}
}

func (h *DeliveryHandler) CreateArrivalPlan(c *gin.Context) {
	c.Set("audit_action", "delivery.arrival_plans.create")
	var req domain.ArrivalPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查到货计划字段")
		return
	}
	item, err := h.service.CreateArrivalPlan(c.Request.Context(), req)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, item)
}

func (h *DeliveryHandler) UpdateArrivalPlan(c *gin.Context) {
	c.Set("audit_action", "delivery.arrival_plans.update")
	var req domain.ArrivalPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查到货计划字段")
		return
	}
	item, err := h.service.UpdateArrivalPlan(c.Request.Context(), c.Param("plan_id"), req)
	if err != nil {
		h.failDeliveryErr(c, err)
		return
	}
	ok(c, item)
}

func (h *DeliveryHandler) ListArrivalPlans(c *gin.Context) {
	c.Set("audit_action", "delivery.arrival_plans.list")
	list, err := h.service.ListArrivalPlans(c.Request.Context(), service.ArrivalPlanFilter{
		Keyword:              c.Query("q"),
		Category:             c.Query("category"),
		Supplier:             c.Query("supplier"),
		OrderNo:              c.Query("order_no"),
		MaterialCode:         c.Query("material_code"),
		EstimatedArrivalFrom: c.Query("estimated_arrival_from"),
		EstimatedArrivalTo:   c.Query("estimated_arrival_to"),
	})
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": len(list), "page": 1, "page_size": len(list)})
}

func (h *DeliveryHandler) DeleteArrivalPlan(c *gin.Context) {
	c.Set("audit_action", "delivery.arrival_plans.delete")
	if err := h.service.DeleteArrivalPlan(c.Request.Context(), c.Param("plan_id")); err != nil {
		h.failDeliveryErr(c, err)
		return
	}
	ok(c, gin.H{"deleted": true, "plan_id": c.Param("plan_id")})
}

func (h *DeliveryHandler) ExportArrivalPlans(c *gin.Context) {
	c.Set("audit_action", "delivery.arrival_plans.export")
	list, err := h.service.ListArrivalPlans(c.Request.Context(), service.ArrivalPlanFilter{
		Keyword: c.Query("q"), Category: c.Query("category"), Supplier: c.Query("supplier"), OrderNo: c.Query("order_no"), MaterialCode: c.Query("material_code"),
	})
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"类别", "物料代码", "物料名称", "数量（台）", "收货地址", "供应商", "订单号", "资产编码区间", "预估到货时间", "备注"}
	writeHeaders(f, sheet, headers)
	for i, item := range list {
		row := i + 2
		values := []any{item.Category, item.MaterialCode, item.MaterialName, item.Quantity, item.ReceivingAddress, item.Supplier, item.OrderNo, item.AssetCodeRange, item.EstimatedArrivalTime, item.Remark}
		writeRow(f, sheet, row, values)
	}
	writeXLSX(c, f, "arrival-plans")
}

func (h *DeliveryHandler) ExportArrivalPlanTemplate(c *gin.Context) {
	c.Set("audit_action", "delivery.arrival_plans.template_export")
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	writeHeaders(f, sheet, []string{"类别", "物料代码", "物料名称", "数量（台）", "收货地址", "供应商", "订单号", "资产编码区间", "预估到货时间", "备注"})
	writeRow(f, sheet, 2, []any{"服务器", "MAT-001", "示例服务器", 10, "上海测试机房", "示例供应商", "ORD-001", "ASSET-001~ASSET-010", "2026-06-15 10:00:00", "示例"})
	writeXLSX(c, f, "arrival-plans-template")
}

func (h *DeliveryHandler) ImportArrivalPlans(c *gin.Context) {
	c.Set("audit_action", "delivery.arrival_plans.import")
	rows, fileOK := readXLSXRows(c)
	if !fileOK {
		return
	}
	result := importResult{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isEmptyRow(row) {
			continue
		}
		quantity, err := parsePositiveInt(cell(row, 3))
		if err != nil {
			result.fail(i+1, err.Error())
			continue
		}
		if _, err := h.service.CreateArrivalPlan(c.Request.Context(), domain.ArrivalPlan{
			Category: cell(row, 0), MaterialCode: cell(row, 1), MaterialName: cell(row, 2), Quantity: quantity, ReceivingAddress: cell(row, 4), Supplier: cell(row, 5), OrderNo: cell(row, 6), AssetCodeRange: cell(row, 7), EstimatedArrivalTime: cell(row, 8), Remark: cell(row, 9),
		}); err != nil {
			result.fail(i+1, err.Error())
			continue
		}
		result.Created++
	}
	ok(c, result)
}

func (h *DeliveryHandler) CreateDeviceArrival(c *gin.Context) {
	c.Set("audit_action", "delivery.device_arrivals.create")
	var req domain.DeviceArrivalRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查服务器&网络设备到货记录字段")
		return
	}
	item, err := h.service.CreateDeviceArrival(c.Request.Context(), req)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, item)
}

func (h *DeliveryHandler) UpdateDeviceArrival(c *gin.Context) {
	c.Set("audit_action", "delivery.device_arrivals.update")
	var req domain.DeviceArrivalRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查服务器&网络设备到货记录字段")
		return
	}
	item, err := h.service.UpdateDeviceArrival(c.Request.Context(), c.Param("record_id"), req)
	if err != nil {
		h.failDeliveryErr(c, err)
		return
	}
	ok(c, item)
}

func (h *DeliveryHandler) ListDeviceArrivals(c *gin.Context) {
	c.Set("audit_action", "delivery.device_arrivals.list")
	list, err := h.service.ListDeviceArrivals(c.Request.Context(), service.DeviceArrivalFilter{
		Keyword:           c.Query("q"),
		Category:          c.Query("category"),
		Manufacturer:      c.Query("manufacturer"),
		PackageCode:       c.Query("package_code"),
		PurchaseRequestNo: c.Query("purchase_request_no"),
		PONo:              c.Query("po_no"),
		ArrivalFrom:       c.Query("arrival_from"),
		ArrivalTo:         c.Query("arrival_to"),
	})
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": len(list), "page": 1, "page_size": len(list)})
}

func (h *DeliveryHandler) DeleteDeviceArrival(c *gin.Context) {
	c.Set("audit_action", "delivery.device_arrivals.delete")
	if err := h.service.DeleteDeviceArrival(c.Request.Context(), c.Param("record_id")); err != nil {
		h.failDeliveryErr(c, err)
		return
	}
	ok(c, gin.H{"deleted": true, "record_id": c.Param("record_id")})
}

func (h *DeliveryHandler) ExportDeviceArrivals(c *gin.Context) {
	c.Set("audit_action", "delivery.device_arrivals.export")
	list, err := h.service.ListDeviceArrivals(c.Request.Context(), service.DeviceArrivalFilter{
		Keyword: c.Query("q"), Category: c.Query("category"), Manufacturer: c.Query("manufacturer"), PackageCode: c.Query("package_code"), PurchaseRequestNo: c.Query("purchase_request_no"), PONo: c.Query("po_no"),
	})
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"类别", "套餐代号", "套餐类型", "物料/服务编码", "物料服务描述", "U数", "厂商", "数量", "收货地点", "采购申请编号", "SRM需求提交时间", "PO号", "实际到货时间"}
	writeHeaders(f, sheet, headers)
	for i, item := range list {
		writeRow(f, sheet, i+2, []any{item.Category, item.PackageCode, item.PackageType, item.MaterialServiceCode, item.MaterialServiceDescription, item.RackUnits, item.Manufacturer, item.Quantity, item.ReceivingLocation, item.PurchaseRequestNo, item.SRMRequirementSubmittedAt, item.PONo, item.ActualArrivalTime})
	}
	writeXLSX(c, f, "device-arrivals")
}

func (h *DeliveryHandler) ExportDeviceArrivalTemplate(c *gin.Context) {
	c.Set("audit_action", "delivery.device_arrivals.template_export")
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	writeHeaders(f, sheet, []string{"类别", "套餐代号", "套餐类型", "物料/服务编码", "物料服务描述", "U数", "厂商", "数量", "收货地点", "采购申请编号", "SRM需求提交时间", "PO号", "实际到货时间"})
	writeRow(f, sheet, 2, []any{"服务器", "PKG-001", "计算型", "MS-001", "示例服务器套餐", 2, "示例厂商", 10, "上海测试机房", "PR-001", "2026-06-01 10:00:00", "PO-001", "2026-06-15 10:00:00"})
	writeXLSX(c, f, "device-arrivals-template")
}

func (h *DeliveryHandler) ImportDeviceArrivals(c *gin.Context) {
	c.Set("audit_action", "delivery.device_arrivals.import")
	rows, fileOK := readXLSXRows(c)
	if !fileOK {
		return
	}
	result := importResult{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isEmptyRow(row) {
			continue
		}
		quantity, err := parsePositiveInt(cell(row, 7))
		if err != nil {
			result.fail(i+1, err.Error())
			continue
		}
		rackUnits, err := parseNonNegativeFloat(cell(row, 5))
		if err != nil {
			result.fail(i+1, err.Error())
			continue
		}
		if _, err := h.service.CreateDeviceArrival(c.Request.Context(), domain.DeviceArrivalRecord{
			Category: cell(row, 0), PackageCode: cell(row, 1), PackageType: cell(row, 2), MaterialServiceCode: cell(row, 3), MaterialServiceDescription: cell(row, 4), RackUnits: rackUnits, Manufacturer: cell(row, 6), Quantity: quantity, ReceivingLocation: cell(row, 8), PurchaseRequestNo: cell(row, 9), SRMRequirementSubmittedAt: cell(row, 10), PONo: cell(row, 11), ActualArrivalTime: cell(row, 12),
		}); err != nil {
			result.fail(i+1, err.Error())
			continue
		}
		result.Created++
	}
	ok(c, result)
}

func (h *DeliveryHandler) CreateAccessoryArrival(c *gin.Context) {
	c.Set("audit_action", "delivery.accessory_arrivals.create")
	var req domain.AccessoryArrivalRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查配件&耗材到货记录字段")
		return
	}
	item, err := h.service.CreateAccessoryArrival(c.Request.Context(), req)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, item)
}

func (h *DeliveryHandler) UpdateAccessoryArrival(c *gin.Context) {
	c.Set("audit_action", "delivery.accessory_arrivals.update")
	var req domain.AccessoryArrivalRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查配件&耗材到货记录字段")
		return
	}
	item, err := h.service.UpdateAccessoryArrival(c.Request.Context(), c.Param("record_id"), req)
	if err != nil {
		h.failDeliveryErr(c, err)
		return
	}
	ok(c, item)
}

func (h *DeliveryHandler) ListAccessoryArrivals(c *gin.Context) {
	c.Set("audit_action", "delivery.accessory_arrivals.list")
	list, err := h.service.ListAccessoryArrivals(c.Request.Context(), service.AccessoryArrivalFilter{
		Keyword:           c.Query("q"),
		Supplier:          c.Query("supplier"),
		IDCRoom:           c.Query("idc_room"),
		PurchaseRequestNo: c.Query("purchase_request_no"),
		PONo:              c.Query("po_no"),
		ArrivalFrom:       c.Query("arrival_from"),
		ArrivalTo:         c.Query("arrival_to"),
	})
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": len(list), "page": 1, "page_size": len(list)})
}

func (h *DeliveryHandler) DeleteAccessoryArrival(c *gin.Context) {
	c.Set("audit_action", "delivery.accessory_arrivals.delete")
	if err := h.service.DeleteAccessoryArrival(c.Request.Context(), c.Param("record_id")); err != nil {
		h.failDeliveryErr(c, err)
		return
	}
	ok(c, gin.H{"deleted": true, "record_id": c.Param("record_id")})
}

func (h *DeliveryHandler) ExportAccessoryArrivals(c *gin.Context) {
	c.Set("audit_action", "delivery.accessory_arrivals.export")
	list, err := h.service.ListAccessoryArrivals(c.Request.Context(), service.AccessoryArrivalFilter{
		Keyword: c.Query("q"), Supplier: c.Query("supplier"), IDCRoom: c.Query("idc_room"), PurchaseRequestNo: c.Query("purchase_request_no"), PONo: c.Query("po_no"),
	})
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"采购申请编号", "物料/服务代码", "物料/服务描述", "数量", "供应商", "机房", "请购背景", "SRM需求提交时间", "PO单号", "到货时间"}
	writeHeaders(f, sheet, headers)
	for i, item := range list {
		writeRow(f, sheet, i+2, []any{item.PurchaseRequestNo, item.MaterialServiceCode, item.MaterialServiceDescription, item.Quantity, item.Supplier, item.IDCRoom, item.PurchaseBackground, item.SRMRequirementSubmittedAt, item.PONo, item.ArrivalTime})
	}
	writeXLSX(c, f, "accessory-arrivals")
}

func (h *DeliveryHandler) ExportAccessoryArrivalTemplate(c *gin.Context) {
	c.Set("audit_action", "delivery.accessory_arrivals.template_export")
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	writeHeaders(f, sheet, []string{"采购申请编号", "物料/服务代码", "物料/服务描述", "数量", "供应商", "机房", "请购背景", "SRM需求提交时间", "PO单号", "到货时间"})
	writeRow(f, sheet, 2, []any{"PR-001", "ACC-001", "示例配件", 100, "示例供应商", "上海测试机房", "扩容备件", "2026-06-01 10:00:00", "PO-001", "2026-06-15 10:00:00"})
	writeXLSX(c, f, "accessory-arrivals-template")
}

func (h *DeliveryHandler) ImportAccessoryArrivals(c *gin.Context) {
	c.Set("audit_action", "delivery.accessory_arrivals.import")
	rows, fileOK := readXLSXRows(c)
	if !fileOK {
		return
	}
	result := importResult{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isEmptyRow(row) {
			continue
		}
		quantity, err := parsePositiveInt(cell(row, 3))
		if err != nil {
			result.fail(i+1, err.Error())
			continue
		}
		if _, err := h.service.CreateAccessoryArrival(c.Request.Context(), domain.AccessoryArrivalRecord{
			PurchaseRequestNo: cell(row, 0), MaterialServiceCode: cell(row, 1), MaterialServiceDescription: cell(row, 2), Quantity: quantity, Supplier: cell(row, 4), IDCRoom: cell(row, 5), PurchaseBackground: cell(row, 6), SRMRequirementSubmittedAt: cell(row, 7), PONo: cell(row, 8), ArrivalTime: cell(row, 9),
		}); err != nil {
			result.fail(i+1, err.Error())
			continue
		}
		result.Created++
	}
	ok(c, result)
}

func (h *DeliveryHandler) failDeliveryErr(c *gin.Context, err error) {
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		fail(c, 40401, err.Error())
		return
	}
	fail(c, 40001, err.Error())
}

type importFailure struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type importResult struct {
	Created  int             `json:"created"`
	Failed   int             `json:"failed"`
	Failures []importFailure `json:"failures,omitempty"`
}

func (r *importResult) fail(row int, reason string) {
	r.Failed++
	r.Failures = append(r.Failures, importFailure{Row: row, Reason: reason})
}

func readXLSXRows(c *gin.Context) ([][]string, bool) {
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传 Excel 文件")
		return nil, false
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		fail(c, 40002, "仅支持 .xlsx 文件")
		return nil, false
	}
	fp, err := file.Open()
	if err != nil {
		fail(c, 40003, "文件读取失败")
		return nil, false
	}
	defer fp.Close()
	xf, err := excelize.OpenReader(fp)
	if err != nil {
		fail(c, 40003, "文件格式无效，请确认是标准 .xlsx")
		return nil, false
	}
	rows, err := xf.GetRows(xf.GetSheetName(0))
	if err != nil {
		fail(c, 40003, "读取 Excel 行失败")
		return nil, false
	}
	return rows, true
}

func writeHeaders(f *excelize.File, sheet string, headers []string) {
	writeRow(f, sheet, 1, toAny(headers))
	_ = f.SetColWidth(sheet, "A", "Z", 18)
}

func writeRow(f *excelize.File, sheet string, row int, values []any) {
	for i, value := range values {
		cellName, _ := excelize.CoordinatesToCellName(i+1, row)
		_ = f.SetCellValue(sheet, cellName, value)
	}
}

func writeXLSX(c *gin.Context, f *excelize.File, prefix string) {
	filename := fmt.Sprintf("%s-%s.xlsx", prefix, time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := f.Write(c.Writer); err != nil {
		fail(c, 50001, err.Error())
	}
}

func toAny(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func parsePositiveInt(raw string) (int, error) {
	var out int
	if _, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &out); err != nil || out <= 0 {
		return 0, fmt.Errorf("quantity must be > 0")
	}
	return out, nil
}

func parseNonNegativeFloat(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	var out float64
	if _, err := fmt.Sscanf(raw, "%f", &out); err != nil || out < 0 {
		return 0, fmt.Errorf("rack_units must be >= 0")
	}
	return out, nil
}
