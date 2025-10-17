ALTER TABLE sensors ADD COLUMN health_status VARCHAR(20) DEFAULT 'unknown';

ALTER TABLE sensors ADD COLUMN health_reason TEXT DEFAULT NULL;

