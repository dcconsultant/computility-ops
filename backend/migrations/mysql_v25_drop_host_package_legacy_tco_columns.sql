-- v25: drop legacy host package TCO component columns (moved to other modules)

SET @has_monthly := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_host_packages' AND COLUMN_NAME = 'monthly_depreciation_cny'
);
SET @sql := IF(@has_monthly > 0,
  'ALTER TABLE ops_host_packages DROP COLUMN monthly_depreciation_cny',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_network := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_host_packages' AND COLUMN_NAME = 'network_cabinet_share_cny'
);
SET @sql := IF(@has_network > 0,
  'ALTER TABLE ops_host_packages DROP COLUMN network_cabinet_share_cny',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_other := (
  SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_host_packages' AND COLUMN_NAME = 'other_fixed_cost_cny'
);
SET @sql := IF(@has_other > 0,
  'ALTER TABLE ops_host_packages DROP COLUMN other_fixed_cost_cny',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
