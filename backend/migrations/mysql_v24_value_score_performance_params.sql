-- v24: value-score 3.2 performance params

CREATE TABLE IF NOT EXISTS ops_value_score_performance_params (
  config_type VARCHAR(128) NOT NULL,
  unavailable_cores INT NOT NULL DEFAULT 0,
  unavailable_memory_gb DECIMAL(18,4) NOT NULL DEFAULT 0,
  performance_score DECIMAL(18,4) NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (config_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
