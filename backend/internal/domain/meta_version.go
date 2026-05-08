package domain

import "time"

type MetaModelVersion struct {
	ID            string    `json:"id"`
	ModelID       string    `json:"model_id"`
	VersionNo     int       `json:"version_no"`
	SnapshotJSON  string    `json:"snapshot_json"`
	PublishedAt   time.Time `json:"published_at"`
	PublishedBy   string    `json:"published_by,omitempty"`
	ChangeSummary string    `json:"change_summary,omitempty"`
}

type MetaModelSnapshot struct {
	Model      MetaModel       `json:"model"`
	Fields     []MetaField     `json:"fields"`
	References []MetaReference `json:"references"`
}
