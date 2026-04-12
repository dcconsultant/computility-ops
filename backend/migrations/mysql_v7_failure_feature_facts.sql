-- Phase1: failure feature facts by record-year and age buckets

CREATE TABLE IF NOT EXISTS ops_failure_feature_facts (
  id                    BIGINT        NOT NULL AUTO_INCREMENT,
  record_year_index     INT           NOT NULL,
  record_year_start     DATE          NOT NULL,
  record_year_end       DATE          NOT NULL,
  scope                 VARCHAR(32)   NOT NULL,
  scene_group           VARCHAR(32)   NOT NULL,
  age_bucket            INT           NOT NULL,
  denominator_weighted  DECIMAL(18,4) NOT NULL DEFAULT 0,
  fault_count           INT           NOT NULL DEFAULT 0,
  fault_rate            DECIMAL(18,8) NOT NULL DEFAULT 0,
  created_at            TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_year_scope_scene_age (record_year_index, scope, scene_group, age_bucket),
  KEY idx_scope_scene_year (scope, scene_group, record_year_index)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
