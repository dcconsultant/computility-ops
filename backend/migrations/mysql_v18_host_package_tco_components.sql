-- v18: host package add TCO component fields for value-score monthly TCO

ALTER TABLE ops_host_packages
  ADD COLUMN monthly_depreciation_cny DECIMAL(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN network_cabinet_share_cny DECIMAL(18,4) NOT NULL DEFAULT 0,
  ADD COLUMN other_fixed_cost_cny DECIMAL(18,4) NOT NULL DEFAULT 0;
