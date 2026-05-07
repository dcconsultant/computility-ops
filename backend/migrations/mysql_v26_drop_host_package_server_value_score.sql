-- v26: remove legacy server_value_score from host package config
ALTER TABLE ops_host_packages
  DROP COLUMN IF EXISTS server_value_score;
