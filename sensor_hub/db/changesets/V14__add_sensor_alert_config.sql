-- Alert rules configuration per sensor
CREATE TABLE IF NOT EXISTS sensor_alert_rules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sensor_id INT NOT NULL,
    alert_type VARCHAR(32) NOT NULL,
    high_threshold FLOAT(4) NULL,
    low_threshold FLOAT(4) NULL,
    trigger_status VARCHAR(64) NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit_hours INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    INDEX idx_sensor_id (sensor_id),
    INDEX idx_enabled (enabled)
);

-- Track when alerts were last sent to implement rate limiting
CREATE TABLE IF NOT EXISTS alert_sent_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    alert_rule_id INT NOT NULL,
    sensor_id INT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    alert_reason TEXT,
    reading_value FLOAT(4) NULL,
    reading_status VARCHAR(64) NULL,
    FOREIGN KEY (alert_rule_id) REFERENCES sensor_alert_rules(id) ON DELETE CASCADE,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    INDEX idx_alert_rule_sent_at (alert_rule_id, sent_at DESC),
    INDEX idx_sensor_sent_at (sensor_id, sent_at DESC)
);

-- Migrate existing temperature thresholds from application.properties to database
-- Using actual values from configuration/application.properties:
-- email.alert.high.temperature.threshold=25
-- email.alert.low.temperature.threshold=12
INSERT INTO sensor_alert_rules (sensor_id, alert_type, high_threshold, low_threshold, enabled, rate_limit_hours)
SELECT 
    id,
    'numeric_range',
    25.0,
    12.0,
    TRUE,
    1
FROM sensors
WHERE type = 'temperature';
