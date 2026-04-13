-- Add warm-storage TOP100 extra capacity fields

ALTER TABLE ops_storage_top_server_rates
  ADD COLUMN IF NOT EXISTS single_disk_capacity_tb DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER data_disk_count,
  ADD COLUMN IF NOT EXISTS total_capacity_tb DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER single_disk_capacity_tb;
