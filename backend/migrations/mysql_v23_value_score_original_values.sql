-- v23: value-score 3.1.2 原值（按配置类型）

CREATE TABLE IF NOT EXISTS ops_value_score_original_values (
  config_type VARCHAR(128) NOT NULL,
  server_original_cny DECIMAL(18,4) NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (config_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
