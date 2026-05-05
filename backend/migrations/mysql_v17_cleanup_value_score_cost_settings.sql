-- v17: cleanup obsolete table introduced by deprecated value-score cost settings design
-- safe to run repeatedly

DROP TABLE IF EXISTS ops_value_score_cost_settings;
