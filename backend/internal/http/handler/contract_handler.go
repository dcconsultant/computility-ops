package handler

import (
	"fmt"
	"path/filepath"
	"strings"

	"computility-ops/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ContractHandler struct {
	service *service.ContractService
}

func NewContractHandler(s *service.ContractService) *ContractHandler {
	return &ContractHandler{service: s}
}

func (h *ContractHandler) CreateContract(c *gin.Context) {
	c.Set("audit_action", "contracts.create")
	var req CreateContractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查合同字段")
		return
	}
	contract, err := h.service.CreateContract(c.Request.Context(), service.CreateContractInput{
		ContractName:    req.ContractName,
		PeriodStart:     req.PeriodStart,
		PeriodEnd:       req.PeriodEnd,
		PreTaxAmount:    req.PreTaxAmount,
		Supplier:        req.Supplier,
		BusinessContact: req.BusinessContact,
		TechContact:     req.TechContact,
	})
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, contract)
}

func (h *ContractHandler) UpdateContract(c *gin.Context) {
	c.Set("audit_action", "contracts.update")
	var req UpdateContractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效，请检查合同字段")
		return
	}
	contract, err := h.service.UpdateContract(c.Request.Context(), c.Param("contract_id"), service.UpdateContractInput{
		ContractName:    req.ContractName,
		PeriodStart:     req.PeriodStart,
		PeriodEnd:       req.PeriodEnd,
		PreTaxAmount:    req.PreTaxAmount,
		Supplier:        req.Supplier,
		BusinessContact: req.BusinessContact,
		TechContact:     req.TechContact,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, contract)
}

func (h *ContractHandler) GetContract(c *gin.Context) {
	c.Set("audit_action", "contracts.get")
	contract, err := h.service.GetContract(c.Request.Context(), c.Param("contract_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, contract)
}

func (h *ContractHandler) ListContracts(c *gin.Context) {
	c.Set("audit_action", "contracts.list")
	list, err := h.service.ListContracts(c.Request.Context())
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": len(list), "page": 1, "page_size": len(list)})
}

func (h *ContractHandler) DeleteContract(c *gin.Context) {
	c.Set("audit_action", "contracts.delete")
	if err := h.service.DeleteContract(c.Request.Context(), c.Param("contract_id")); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "contract_id": c.Param("contract_id")})
}

func (h *ContractHandler) UploadAttachment(c *gin.Context) {
	c.Set("audit_action", "contracts.upload_attachment")
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, 40001, "请上传附件文件")
		return
	}
	contract, attachment, err := h.service.SaveAttachment(c.Request.Context(), c.Param("contract_id"), file)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 40001, err.Error())
		return
	}
	ok(c, gin.H{"contract": contract, "attachment": attachment})
}

func (h *ContractHandler) DownloadAttachment(c *gin.Context) {
	c.Set("audit_action", "contracts.download_attachment")
	blob, err := h.service.GetAttachment(c.Request.Context(), c.Param("contract_id"), c.Param("attachment_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	filename := filepath.Base(strings.TrimSpace(blob.FileName))
	if filename == "" {
		filename = fmt.Sprintf("attachment_%s", blob.AttachmentID)
	}
	mimeType := strings.TrimSpace(blob.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.FileAttachment(blob.StoragePath, filename)
}

func (h *ContractHandler) DeleteAttachment(c *gin.Context) {
	c.Set("audit_action", "contracts.delete_attachment")
	contract, err := h.service.DeleteAttachment(c.Request.Context(), c.Param("contract_id"), c.Param("attachment_id"))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			fail(c, 40401, err.Error())
			return
		}
		fail(c, 50001, err.Error())
		return
	}
	ok(c, contract)
}
