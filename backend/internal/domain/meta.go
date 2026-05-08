package domain

import "time"

type MetaModelStatus string

const (
	MetaModelStatusDraft     MetaModelStatus = "draft"
	MetaModelStatusPublished MetaModelStatus = "published"
	MetaModelStatusArchived  MetaModelStatus = "archived"
)

type MetaModel struct {
	ID             string          `json:"id"`
	ModelCode      string          `json:"model_code"`
	ModelName      string          `json:"model_name"`
	Description    string          `json:"description,omitempty"`
	Status         MetaModelStatus `json:"status"`
	CurrentVersion int             `json:"current_version"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type MetaField struct {
	ID             string    `json:"id"`
	ModelID        string    `json:"model_id"`
	FieldCode      string    `json:"field_code"`
	FieldName      string    `json:"field_name"`
	Category       string    `json:"category,omitempty"`
	ValueType      string    `json:"value_type"`
	Required       bool      `json:"required"`
	Unique         bool      `json:"unique"`
	Filterable     bool      `json:"filterable"`
	Sortable       bool      `json:"sortable"`
	Visible        bool      `json:"visible"`
	DefaultValue   string    `json:"default_value,omitempty"`
	ValidationRule string    `json:"validation_rule,omitempty"`
	SortNo         int       `json:"sort_no"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MetaReference struct {
	ID             string    `json:"id"`
	ModelID        string    `json:"model_id"`
	SourceFieldID  string    `json:"source_field_id"`
	TargetModelID  string    `json:"target_model_id"`
	TargetFieldID  string    `json:"target_field_id"`
	DisplayFields  []string  `json:"display_fields"`
	OnDeleteAction string    `json:"on_delete_action"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
