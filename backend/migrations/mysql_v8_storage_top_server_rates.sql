-- Recent 1-year storage server fault-rate TOP table

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
