-- v26: remove legacy server_value_score from host package config
-- compatible with MySQL versions that do not support `DROP COLUMN IF EXISTS`
SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ops_host_packages'
    AND COLUMN_NAME = 'server_value_score'
);
SET @ddl := IF(
  @col_exists > 0,
  'ALTER TABLE ops_host_packages DROP COLUMN server_value_score',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
