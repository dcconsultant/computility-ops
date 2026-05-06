-- v21: host package add server average original value for value-score depreciation

ALTER TABLE ops_host_packages
  ADD COLUMN server_avg_original_value_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER server_value_score;
