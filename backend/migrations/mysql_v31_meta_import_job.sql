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
  KEY idx_md_import_job_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 兼容老库：确保 model_id 与 md_model.id 的 collation 完全一致，再加外键
SET @model_id_collation := (
  SELECT COLLATION_NAME
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'md_model'
    AND COLUMN_NAME = 'id'
  LIMIT 1
);
SET @sql := IF(
  @model_id_collation IS NULL,
  'SELECT 1',
  CONCAT('ALTER TABLE md_import_job MODIFY model_id VARCHAR(64) CHARACTER SET utf8mb4 COLLATE ', @model_id_collation, ' NOT NULL')
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE md_import_job
  ADD CONSTRAINT fk_md_import_job_model
  FOREIGN KEY (model_id) REFERENCES md_model(id) ON DELETE CASCADE;
