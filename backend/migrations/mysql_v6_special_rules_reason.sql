-- 为续保例外清单增加原因字段（由导入模板提供）

ALTER TABLE ops_special_rules
  ADD COLUMN reason VARCHAR(255) NULL COMMENT '例外原因（导入提供）' AFTER policy;
