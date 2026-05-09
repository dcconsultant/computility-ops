CREATE TABLE IF NOT EXISTS md_record (
  id VARCHAR(64) PRIMARY KEY,
  model_id VARCHAR(64) NOT NULL,
  data_json LONGTEXT NOT NULL,
  deleted_flag TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_model_deleted_updated (model_id, deleted_flag, updated_at),
  CONSTRAINT fk_md_record_model FOREIGN KEY (model_id) REFERENCES md_model(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
