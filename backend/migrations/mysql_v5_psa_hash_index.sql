-- Add hash acceleration for exact PSA matching while preserving LONGTEXT payload.

ALTER TABLE ops_servers
  ADD COLUMN psa_hash CHAR(64) NOT NULL DEFAULT '' AFTER psa;

UPDATE ops_servers
SET psa_hash = SHA2(COALESCE(psa, ''), 256)
WHERE psa_hash = '' OR psa_hash IS NULL;

ALTER TABLE ops_servers
  ADD KEY idx_psa_hash (psa_hash);

ALTER TABLE ops_special_rules
  ADD COLUMN psa_hash CHAR(64) NULL AFTER psa;

UPDATE ops_special_rules
SET psa_hash = CASE
  WHEN psa IS NULL OR psa = '' THEN NULL
  ELSE SHA2(psa, 256)
END
WHERE psa_hash IS NULL;

ALTER TABLE ops_special_rules
  ADD KEY idx_psa_hash (psa_hash);
