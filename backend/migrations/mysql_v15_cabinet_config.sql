-- v15: cabinet config management for value score stage-1 dependency

CREATE TABLE IF NOT EXISTS ops_cabinet_settings (
  id           TINYINT       NOT NULL,
  utilization  DECIMAL(18,8) NOT NULL DEFAULT 1,
  updated_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO ops_cabinet_settings (id, utilization)
VALUES (1, 1)
ON DUPLICATE KEY UPDATE utilization = utilization;

CREATE TABLE IF NOT EXISTS ops_cabinet_configs (
  id              BIGINT        NOT NULL AUTO_INCREMENT,
  idc             VARCHAR(128)  NOT NULL,
  rated_power_kw  DECIMAL(18,8) NOT NULL,
  monthly_rent    DECIMAL(18,8) NOT NULL,
  updated_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_idc_rated_power (idc, rated_power_kw),
  KEY idx_idc (idc)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
