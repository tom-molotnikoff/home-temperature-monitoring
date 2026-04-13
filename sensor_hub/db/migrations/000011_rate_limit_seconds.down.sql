-- Revert rate_limit_seconds back to rate_limit_hours
UPDATE sensor_alert_rules SET rate_limit_seconds = rate_limit_seconds / 3600;
ALTER TABLE sensor_alert_rules RENAME COLUMN rate_limit_seconds TO rate_limit_hours;
