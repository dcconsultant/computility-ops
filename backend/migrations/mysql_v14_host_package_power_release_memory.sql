-- v14: host package add power/release_year/memory_capacity_gb

ALTER TABLE ops_host_packages
  ADD COLUMN power_watts DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER storage_capacity_tb,
  ADD COLUMN release_year INT NOT NULL DEFAULT 0 AFTER power_watts,
  ADD COLUMN memory_capacity_gb DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER release_year;
