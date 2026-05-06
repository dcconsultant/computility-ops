-- v22: cleanup unused host package avg original value column

SET @has_col := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ops_host_packages'
    AND COLUMN_NAME = 'server_avg_original_value_cny'
);

SET @sql := IF(@has_col = 1,
  'ALTER TABLE ops_host_packages DROP COLUMN server_avg_original_value_cny',
  'SELECT 1');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
