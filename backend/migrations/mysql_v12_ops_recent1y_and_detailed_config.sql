-- v12: 服务器详细配置 + 套餐最近1年故障率

ALTER TABLE ops_servers
  ADD COLUMN IF NOT EXISTS detailed_config LONGTEXT NULL COMMENT '详细配置';

ALTER TABLE ops_package_failure_rates
  ADD COLUMN IF NOT EXISTS recent_1y_failure_rate DECIMAL(18,8) NOT NULL DEFAULT 0 COMMENT '最近1年故障率';
