CREATE TABLE IF NOT EXISTS md_import_job (
  job_id VARCHAR(64) PRIMARY KEY,
  model_id VARCHAR(64) NOT NULL,
  status VARCHAR(16) NOT NULL,
  total INT NOT NULL DEFAULT 0,
  processed INT NOT NULL DEFAULT 0,
  success INT NOT NULL DEFAULT 0,
  failed INT NOT NULL DEFAULT 0,
  errors_json LONGTEXT NULL,
  started_at DATETIME NOT NULL,
  finished_at DATETIME NULL,
  message VARCHAR(500) NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_md_import_job_model (model_id),
  KEY idx_md_import_job_status (status),
  CONSTRAINT fk_md_import_job_model FOREIGN KEY (model_id) REFERENCES md_model(id) ON DELETE CASCADE
);
