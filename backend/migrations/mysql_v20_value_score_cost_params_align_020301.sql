-- v20: align value-score cost params with 020301 terminology

-- Add columns in a MySQL-version-compatible way (no ADD COLUMN IF NOT EXISTS)
SET @has_col_server_avg := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'server_avg_original_value_cny'
);
SET @sql_add_server_avg := IF(@has_col_server_avg = 0,
  'ALTER TABLE ops_value_score_cost_params ADD COLUMN server_avg_original_value_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER depreciation_months',
  'SELECT 1');
PREPARE stmt_add_server_avg FROM @sql_add_server_avg; EXECUTE stmt_add_server_avg; DEALLOCATE PREPARE stmt_add_server_avg;

SET @has_col_network_device := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'network_device_share_cny'
);
SET @sql_add_network_device := IF(@has_col_network_device = 0,
  'ALTER TABLE ops_value_score_cost_params ADD COLUMN network_device_share_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER server_avg_original_value_cny',
  'SELECT 1');
PREPARE stmt_add_network_device FROM @sql_add_network_device; EXECUTE stmt_add_network_device; DEALLOCATE PREPARE stmt_add_network_device;

SET @has_col_server_renewal := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'server_renewal_fee_cny'
);
SET @sql_add_server_renewal := IF(@has_col_server_renewal = 0,
  'ALTER TABLE ops_value_score_cost_params ADD COLUMN server_renewal_fee_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER network_device_share_cny',
  'SELECT 1');
PREPARE stmt_add_server_renewal FROM @sql_add_server_renewal; EXECUTE stmt_add_server_renewal; DEALLOCATE PREPARE stmt_add_server_renewal;

-- Backfill from old columns if they exist
SET @has_old_network := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'network_cabinet_share_cny'
);
SET @has_old_other := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'other_fixed_cost_cny'
);

SET @sql1 := IF(@has_old_network = 1,
  'UPDATE ops_value_score_cost_params SET network_device_share_cny = COALESCE(network_cabinet_share_cny,0) WHERE id=1',
  'SELECT 1');
PREPARE stmt1 FROM @sql1; EXECUTE stmt1; DEALLOCATE PREPARE stmt1;

SET @sql2 := IF(@has_old_other = 1,
  'UPDATE ops_value_score_cost_params SET server_renewal_fee_cny = COALESCE(other_fixed_cost_cny,0) WHERE id=1',
  'SELECT 1');
PREPARE stmt2 FROM @sql2; EXECUTE stmt2; DEALLOCATE PREPARE stmt2;
