-- 故障率看板（机龄趋势 + 概览卡片）持久化表

CREATE TABLE IF NOT EXISTS failure_overview_cards (
  id                        BIGINT        NOT NULL AUTO_INCREMENT,
  segment                   VARCHAR(32)   NOT NULL,
  stat_year                 INT           NOT NULL,
  current_year_fault_rate   DECIMAL(18,8) NOT NULL DEFAULT 0,
  history_avg_fault_rate    DECIMAL(18,8) NOT NULL DEFAULT 0,
  current_year_fault_count  INT           NOT NULL DEFAULT 0,
  current_year_denominator  DECIMAL(18,4) NOT NULL DEFAULT 0,
  history_fault_count       INT           NOT NULL DEFAULT 0,
  history_denominator       DECIMAL(18,4) NOT NULL DEFAULT 0,
  created_at                TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_segment_year (segment, stat_year)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS failure_age_trend_points (
  id                      BIGINT        NOT NULL AUTO_INCREMENT,
  segment                 VARCHAR(32)   NOT NULL,
  age_bucket              INT           NOT NULL,
  numerator_fault_count   INT           NOT NULL DEFAULT 0,
  denominator_exposure    DECIMAL(18,4) NOT NULL DEFAULT 0,
  fault_rate              DECIMAL(18,8) NOT NULL DEFAULT 0,
  created_at              TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_segment_age (segment, age_bucket)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
