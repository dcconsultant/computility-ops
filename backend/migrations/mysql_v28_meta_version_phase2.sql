CREATE TABLE IF NOT EXISTS md_model_version (
  id VARCHAR(64) PRIMARY KEY,
  model_id VARCHAR(64) NOT NULL,
  version_no INT NOT NULL,
  snapshot_json LONGTEXT NOT NULL,
  published_at DATETIME NOT NULL,
  published_by VARCHAR(64) NULL,
  change_summary VARCHAR(512) NULL,
  UNIQUE KEY uk_model_version (model_id, version_no),
  KEY idx_published_at (published_at),
  CONSTRAINT fk_md_model_version_model FOREIGN KEY (model_id) REFERENCES md_model(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
