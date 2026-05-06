-- v19: value-score cost parameter settings (3.1.1)

CREATE TABLE IF NOT EXISTS ops_value_score_cost_params (
  id BIGINT NOT NULL PRIMARY KEY,
  depreciation_months INT NOT NULL DEFAULT 60,
  server_avg_original_value_cny DECIMAL(18,4) NOT NULL DEFAULT 0,
  network_device_share_cny DECIMAL(18,4) NOT NULL DEFAULT 0,
  server_renewal_fee_cny DECIMAL(18,4) NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO ops_value_score_cost_params (id, depreciation_months, server_avg_original_value_cny, network_device_share_cny, server_renewal_fee_cny)
VALUES (1, 60, 0, 0, 0)
ON DUPLICATE KEY UPDATE id = id;
