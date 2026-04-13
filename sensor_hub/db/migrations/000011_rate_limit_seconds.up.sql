-- Convert rate_limit_hours to rate_limit_seconds
-- Existing values are multiplied by 3600 to preserve the same duration
ALTER TABLE sensor_alert_rules RENAME COLUMN rate_limit_hours TO rate_limit_seconds;
UPDATE sensor_alert_rules SET rate_limit_seconds = rate_limit_seconds * 3600;
