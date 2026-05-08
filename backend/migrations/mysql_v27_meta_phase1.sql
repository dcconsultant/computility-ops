CREATE TABLE IF NOT EXISTS md_model (
  id VARCHAR(64) PRIMARY KEY,
  model_code VARCHAR(64) NOT NULL,
  model_name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'draft',
  current_version INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_model_code (model_code),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS md_model_field (
  id VARCHAR(64) PRIMARY KEY,
  model_id VARCHAR(64) NOT NULL,
  field_code VARCHAR(64) NOT NULL,
  field_name VARCHAR(128) NOT NULL,
  category VARCHAR(32) NULL,
  value_type VARCHAR(32) NOT NULL,
  required_flag TINYINT NOT NULL DEFAULT 0,
  unique_flag TINYINT NOT NULL DEFAULT 0,
  filterable_flag TINYINT NOT NULL DEFAULT 0,
  sortable_flag TINYINT NOT NULL DEFAULT 0,
  visible_flag TINYINT NOT NULL DEFAULT 1,
  default_value TEXT NULL,
  validation_rule TEXT NULL,
  sort_no INT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_model_field_code (model_id, field_code),
  KEY idx_model_sort (model_id, sort_no),
  CONSTRAINT fk_md_model_field_model FOREIGN KEY (model_id) REFERENCES md_model(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS md_model_ref (
  id VARCHAR(64) PRIMARY KEY,
  model_id VARCHAR(64) NOT NULL,
  source_field_id VARCHAR(64) NOT NULL,
  target_model_id VARCHAR(64) NOT NULL,
  target_field_id VARCHAR(64) NOT NULL,
  display_fields_json JSON NULL,
  on_delete_action VARCHAR(16) NOT NULL DEFAULT 'restrict',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_model (model_id),
  CONSTRAINT fk_md_model_ref_model FOREIGN KEY (model_id) REFERENCES md_model(id) ON DELETE CASCADE,
  CONSTRAINT fk_md_model_ref_source_field FOREIGN KEY (source_field_id) REFERENCES md_model_field(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
