-- v16: value score step1 cost settings

CREATE TABLE IF NOT EXISTS ops_value_score_cost_settings (
  id                              TINYINT       NOT NULL,
  electricity_price_cny_per_kwh   DECIMAL(18,8) NOT NULL DEFAULT 0,
  depreciation_months             INT           NOT NULL DEFAULT 60,
  cabinet_utilization             DECIMAL(18,8) NOT NULL DEFAULT 1,
  updated_at                      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at                      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO ops_value_score_cost_settings (id, electricity_price_cny_per_kwh, depreciation_months, cabinet_utilization)
VALUES (1, 0, 60, 1)
ON DUPLICATE KEY UPDATE id=id;
