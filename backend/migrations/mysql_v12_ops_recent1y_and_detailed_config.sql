-- v12: 服务器详细配置 + 套餐最近1年故障率
-- 兼容 MySQL 5.7/8.0（不依赖 ADD COLUMN IF NOT EXISTS）

SET @db := DATABASE();

-- 1) ops_servers.detailed_config
SET @sql := (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE ops_servers ADD COLUMN detailed_config LONGTEXT NULL COMMENT ''详细配置''',
    'SELECT ''skip: ops_servers.detailed_config exists'''
  )
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'ops_servers'
    AND COLUMN_NAME = 'detailed_config'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2) ops_package_failure_rates.recent_1y_failure_rate
SET @sql := (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE ops_package_failure_rates ADD COLUMN recent_1y_failure_rate DECIMAL(18,8) NOT NULL DEFAULT 0 COMMENT ''最近1年故障率''',
    'SELECT ''skip: ops_package_failure_rates.recent_1y_failure_rate exists'''
  )
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'ops_package_failure_rates'
    AND COLUMN_NAME = 'recent_1y_failure_rate'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
