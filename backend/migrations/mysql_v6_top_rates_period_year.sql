-- Add period/year dimensions for package-level TOP fault rates

ALTER TABLE ops_package_failure_rates
  ADD COLUMN period VARCHAR(32) NOT NULL DEFAULT 'history' AFTER id,
  ADD COLUMN stat_year INT NOT NULL DEFAULT 0 AFTER period;

ALTER TABLE ops_package_failure_rates
  ADD KEY idx_period_year_cfg (period, stat_year, config_type);

ALTER TABLE ops_package_model_failure_rates
  ADD COLUMN period VARCHAR(32) NOT NULL DEFAULT 'history' AFTER id,
  ADD COLUMN stat_year INT NOT NULL DEFAULT 0 AFTER period;

ALTER TABLE ops_package_model_failure_rates
  ADD KEY idx_period_year_cfg_model (period, stat_year, config_type, manufacturer, model);
