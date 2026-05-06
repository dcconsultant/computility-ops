-- v20: align value-score cost params with 020301 terminology

ALTER TABLE ops_value_score_cost_params
  ADD COLUMN IF NOT EXISTS server_avg_original_value_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER depreciation_months,
  ADD COLUMN IF NOT EXISTS network_device_share_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER server_avg_original_value_cny,
  ADD COLUMN IF NOT EXISTS server_renewal_fee_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER network_device_share_cny;

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
