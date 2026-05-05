-- v18: host package add TCO component fields for value-score monthly TCO

ALTER TABLE ops_host_packages
  ADD COLUMN monthly_depreciation_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER server_value_score,
  ADD COLUMN network_cabinet_share_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER monthly_depreciation_cny,
  ADD COLUMN other_fixed_cost_cny DECIMAL(18,4) NOT NULL DEFAULT 0 AFTER network_cabinet_share_cny;
