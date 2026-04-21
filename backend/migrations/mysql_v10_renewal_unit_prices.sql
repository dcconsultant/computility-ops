-- 续保单价维护表（国家 + 场景大类）

CREATE TABLE IF NOT EXISTS ops_renewal_unit_prices (
  id             BIGINT        NOT NULL AUTO_INCREMENT,
  country        VARCHAR(32)   NOT NULL,
  scene_category VARCHAR(32)   NOT NULL,
  unit_price     DECIMAL(18,4) NOT NULL DEFAULT 0,
  created_at     TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_country_scene (country, scene_category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
