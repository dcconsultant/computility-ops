CREATE TABLE IF NOT EXISTS servers (
  id            VARCHAR(64)   NOT NULL,
  hostname      VARCHAR(128)  NOT NULL,
  location      VARCHAR(128)  NULL,
  cpu_cores     INT           NOT NULL,
  value_score   DECIMAL(10,2) NOT NULL,
  expire_date   DATE          NULL,
  remark        VARCHAR(255)  NULL,
  updated_at    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_value_score (value_score),
  KEY idx_cpu_cores (cpu_cores)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS renewal_plans (
  id             BIGINT       NOT NULL AUTO_INCREMENT,
  target_cores   INT          NOT NULL,
  selected_cores INT          NOT NULL,
  created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS renewal_plan_items (
  id            BIGINT        NOT NULL AUTO_INCREMENT,
  plan_id       BIGINT        NOT NULL,
  server_id     VARCHAR(64)   NOT NULL,
  hostname      VARCHAR(128)  NOT NULL,
  cpu_cores     INT           NOT NULL,
  value_score   DECIMAL(10,2) NOT NULL,
  rank_no       INT           NOT NULL,
  created_at    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_plan_id (plan_id),
  CONSTRAINT fk_plan_items_plan FOREIGN KEY (plan_id) REFERENCES renewal_plans(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
