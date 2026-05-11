-- v32: move cabinet baseline params to global cost params
-- add cabinet_utilization/rated_power_kw/monthly_rent_cny into ops_value_score_cost_params
-- and backfill from legacy cabinet tables for compatibility

SET @has_cabinet_utilization := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'cabinet_utilization'
);
SET @sql_add_cabinet_utilization := IF(@has_cabinet_utilization = 0,
  'ALTER TABLE ops_value_score_cost_params ADD COLUMN cabinet_utilization DECIMAL(18,8) NOT NULL DEFAULT 1 AFTER server_renewal_fee_cny',
  'SELECT 1');
PREPARE stmt_add_cabinet_utilization FROM @sql_add_cabinet_utilization; EXECUTE stmt_add_cabinet_utilization; DEALLOCATE PREPARE stmt_add_cabinet_utilization;

SET @has_rated_power_kw := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'rated_power_kw'
);
SET @sql_add_rated_power_kw := IF(@has_rated_power_kw = 0,
  'ALTER TABLE ops_value_score_cost_params ADD COLUMN rated_power_kw DECIMAL(18,8) NOT NULL DEFAULT 0 AFTER cabinet_utilization',
  'SELECT 1');
PREPARE stmt_add_rated_power_kw FROM @sql_add_rated_power_kw; EXECUTE stmt_add_rated_power_kw; DEALLOCATE PREPARE stmt_add_rated_power_kw;

SET @has_monthly_rent_cny := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_value_score_cost_params' AND COLUMN_NAME = 'monthly_rent_cny'
);
SET @sql_add_monthly_rent_cny := IF(@has_monthly_rent_cny = 0,
  'ALTER TABLE ops_value_score_cost_params ADD COLUMN monthly_rent_cny DECIMAL(18,8) NOT NULL DEFAULT 0 AFTER rated_power_kw',
  'SELECT 1');
PREPARE stmt_add_monthly_rent_cny FROM @sql_add_monthly_rent_cny; EXECUTE stmt_add_monthly_rent_cny; DEALLOCATE PREPARE stmt_add_monthly_rent_cny;

-- ensure base row exists
INSERT INTO ops_value_score_cost_params (id, depreciation_months, network_device_share_cny, server_renewal_fee_cny, cabinet_utilization, rated_power_kw, monthly_rent_cny)
VALUES (1, 60, 0, 0, 1, 0, 0)
ON DUPLICATE KEY UPDATE id = id;

-- backfill cabinet_utilization from legacy singleton table when available
SET @has_legacy_util := (
  SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_cabinet_settings'
);
SET @sql_backfill_util := IF(@has_legacy_util = 1,
  'UPDATE ops_value_score_cost_params p
   JOIN ops_cabinet_settings s ON s.id = 1
   SET p.cabinet_utilization = IFNULL(NULLIF(s.utilization,0),1)
   WHERE p.id = 1 AND (p.cabinet_utilization IS NULL OR p.cabinet_utilization <= 0)',
  'SELECT 1');
PREPARE stmt_backfill_util FROM @sql_backfill_util; EXECUTE stmt_backfill_util; DEALLOCATE PREPARE stmt_backfill_util;

-- backfill rated_power_kw/monthly_rent_cny from legacy cabinet config logic
-- use target idc CN-N01-TJ01-ZJ01 + minimum rated_power_kw rule
SET @has_legacy_cabinet := (
  SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ops_cabinet_configs'
);
SET @sql_backfill_cabinet := IF(@has_legacy_cabinet = 1,
  'UPDATE ops_value_score_cost_params p
   JOIN (
     SELECT c.rated_power_kw, c.monthly_rent
     FROM ops_cabinet_configs c
     WHERE c.idc = ''CN-N01-TJ01-ZJ01''
     ORDER BY c.rated_power_kw ASC, c.monthly_rent ASC, c.id ASC
     LIMIT 1
   ) t ON 1=1
   SET p.rated_power_kw = IF(p.rated_power_kw <= 0, t.rated_power_kw, p.rated_power_kw),
       p.monthly_rent_cny = IF(p.monthly_rent_cny <= 0, t.monthly_rent, p.monthly_rent_cny)
   WHERE p.id = 1',
  'SELECT 1');
PREPARE stmt_backfill_cabinet FROM @sql_backfill_cabinet; EXECUTE stmt_backfill_cabinet; DEALLOCATE PREPARE stmt_backfill_cabinet;

-- final guard
UPDATE ops_value_score_cost_params
SET cabinet_utilization = IFNULL(NULLIF(cabinet_utilization, 0), 1)
WHERE id = 1;
