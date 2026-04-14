-- computility-ops MySQL 持久化表（与当前 backend domain 对齐）

CREATE TABLE IF NOT EXISTS ops_servers (
  sn                  VARCHAR(64)   NOT NULL,
  manufacturer        VARCHAR(128)  NULL,
  model               VARCHAR(128)  NULL,
  psa                 LONGTEXT      NOT NULL,
  psa_hash            CHAR(64)      NOT NULL,
  idc                 VARCHAR(128)  NULL,
  environment         VARCHAR(64)   NULL,
  config_type         VARCHAR(128)  NOT NULL,
  warranty_end_date   VARCHAR(32)   NULL,
  launch_date         VARCHAR(32)   NULL,
  created_at          TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (sn),
  KEY idx_psa_hash (psa_hash),
  KEY idx_config_type (config_type),
  KEY idx_environment (environment)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ops_host_packages (
  id                        BIGINT        NOT NULL AUTO_INCREMENT,
  config_type               VARCHAR(128)  NOT NULL,
  scene_category            VARCHAR(64)   NULL,
  cpu_logical_cores         INT           NOT NULL,
  gpu_card_count            INT           NOT NULL DEFAULT 0,
  data_disk_type            VARCHAR(64)   NULL,
  data_disk_count           INT           NOT NULL DEFAULT 0,
  storage_capacity_tb       DECIMAL(18,4) NOT NULL DEFAULT 0,
  server_value_score        DECIMAL(18,4) NOT NULL DEFAULT 0,
  arch_standardized_factor  DECIMAL(18,4) NOT NULL DEFAULT 1,
  created_at                TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_config_type (config_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ops_special_rules (
  id                  BIGINT        NOT NULL AUTO_INCREMENT,
  sn                  VARCHAR(64)   NOT NULL,
  manufacturer        VARCHAR(128)  NULL,
  model               VARCHAR(128)  NULL,
  psa                 LONGTEXT      NULL,
  psa_hash            CHAR(64)      NULL,
  idc                 VARCHAR(128)  NULL,
  package_type        VARCHAR(128)  NULL,
  warranty_end_date   VARCHAR(32)   NULL,
  launch_date         VARCHAR(32)   NULL,
  policy              VARCHAR(32)   NOT NULL,
  reason              VARCHAR(255)  NULL,
  created_at          TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_sn (sn),
  KEY idx_psa_hash (psa_hash),
  KEY idx_policy (policy)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ops_model_failure_rates (
  id                          BIGINT        NOT NULL AUTO_INCREMENT,
  manufacturer                VARCHAR(128)  NOT NULL,
  model                       VARCHAR(128)  NOT NULL,
  failure_rate                DECIMAL(18,8) NOT NULL,
  over_warranty_failure_rate  DECIMAL(18,8) NOT NULL DEFAULT 0,
  created_at                  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_model (manufacturer, model)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ops_package_failure_rates (
  id                          BIGINT        NOT NULL AUTO_INCREMENT,
  period                      VARCHAR(32)   NOT NULL DEFAULT 'history',
  stat_year                   INT           NOT NULL DEFAULT 0,
  config_type                 VARCHAR(128)  NOT NULL,
  failure_rate                DECIMAL(18,8) NOT NULL,
  over_warranty_failure_rate  DECIMAL(18,8) NOT NULL DEFAULT 0,
  created_at                  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_period_year_cfg (period, stat_year, config_type),
  KEY idx_config_type (config_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ops_package_model_failure_rates (
  id                          BIGINT        NOT NULL AUTO_INCREMENT,
  period                      VARCHAR(32)   NOT NULL DEFAULT 'history',
  stat_year                   INT           NOT NULL DEFAULT 0,
  config_type                 VARCHAR(128)  NOT NULL,
  manufacturer                VARCHAR(128)  NOT NULL,
  model                       VARCHAR(128)  NOT NULL,
  failure_rate                DECIMAL(18,8) NOT NULL,
  over_warranty_failure_rate  DECIMAL(18,8) NOT NULL DEFAULT 0,
  created_at                  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_period_year_cfg_model (period, stat_year, config_type, manufacturer, model),
  KEY idx_cfg_model (config_type, manufacturer, model)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ops_overall_failure_rates (
  id                          BIGINT        NOT NULL AUTO_INCREMENT,
  period                      VARCHAR(32)   NOT NULL,
  stat_year                   INT           NOT NULL DEFAULT 0,
  scope                       VARCHAR(32)   NOT NULL,
  segment                     VARCHAR(32)   NOT NULL,
  full_cycle_failure_rate     DECIMAL(18,8) NOT NULL,
  over_warranty_failure_rate  DECIMAL(18,8) NOT NULL DEFAULT 0,
  fault_count                 INT           NOT NULL DEFAULT 0,
  over_warranty_fault_count   INT           NOT NULL DEFAULT 0,
  server_years                DECIMAL(18,4) NOT NULL DEFAULT 0,
  over_warranty_years         DECIMAL(18,4) NOT NULL DEFAULT 0,
  created_at                  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_period_scope_segment (period, scope, segment)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ops_renewal_plans (
  plan_id      VARCHAR(64)   NOT NULL,
  payload_json LONGTEXT      NOT NULL,
  created_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (plan_id),
  KEY idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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

CREATE TABLE IF NOT EXISTS ops_storage_top_server_rates (
  id              BIGINT        NOT NULL AUTO_INCREMENT,
  sn              VARCHAR(128)  NOT NULL,
  manufacturer    VARCHAR(128)  NULL,
  model           VARCHAR(128)  NULL,
  config_type     VARCHAR(128)  NULL,
  environment     VARCHAR(64)   NULL,
  idc             VARCHAR(128)  NULL,
  data_disk_count INT           NOT NULL DEFAULT 0,
  single_disk_capacity_tb DECIMAL(18,4) NOT NULL DEFAULT 0,
  total_capacity_tb DECIMAL(18,4) NOT NULL DEFAULT 0,
  fault_count     INT           NOT NULL DEFAULT 0,
  denominator     DECIMAL(18,4) NOT NULL DEFAULT 0,
  fault_rate      DECIMAL(18,8) NOT NULL DEFAULT 0,
  created_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_fault_rate (fault_rate),
  KEY idx_sn (sn)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
