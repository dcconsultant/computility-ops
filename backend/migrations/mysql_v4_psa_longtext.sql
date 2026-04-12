-- Extend PSA field capacity for real business payloads (up to very long text)
-- Safe to run multiple times in environments where columns may already be LONGTEXT.

ALTER TABLE ops_servers
  MODIFY COLUMN psa LONGTEXT NOT NULL;

ALTER TABLE ops_special_rules
  MODIFY COLUMN psa LONGTEXT NULL;
