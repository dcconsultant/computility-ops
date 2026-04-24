package domain

import "time"

type Contract struct {
	ContractID      string               `json:"contract_id"`
	ContractName    string               `json:"contract_name"`
	PeriodStart     string               `json:"period_start"`
	PeriodEnd       string               `json:"period_end"`
	PreTaxAmount    float64              `json:"pre_tax_amount"`
	Supplier        string               `json:"supplier"`
	BusinessContact string               `json:"business_contact"`
	TechContact     string               `json:"tech_contact"`
	Attachments     []ContractAttachment `json:"attachments,omitempty"`
	CreatedAt       string               `json:"created_at,omitempty"`
	UpdatedAt       string               `json:"updated_at,omitempty"`
}

type ContractAttachment struct {
	AttachmentID string `json:"attachment_id"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type,omitempty"`
	UploadedAt   string `json:"uploaded_at"`
}

type ContractAttachmentBlob struct {
	ContractID   string
	AttachmentID string
	FileName     string
	StoragePath  string
	FileSize     int64
	MimeType     string
	CreatedAt    time.Time
}
