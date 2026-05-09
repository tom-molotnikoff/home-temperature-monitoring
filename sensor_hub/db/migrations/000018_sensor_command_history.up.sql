CREATE TABLE sensor_command_history (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id          INTEGER NOT NULL REFERENCES sensors(id),
    user_id            INTEGER REFERENCES users(id),
    property           TEXT NOT NULL,
    value              TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'sent',
    mqtt_topic         TEXT NOT NULL,
    mqtt_payload       TEXT NOT NULL,
    timeout_seconds    INTEGER NOT NULL DEFAULT 10,
    sent_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at    DATETIME,
    acknowledged_value TEXT NULL,
    created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sensor_command_history_sensor_id
    ON sensor_command_history(sensor_id, sent_at DESC);

CREATE INDEX idx_sensor_command_history_pending
    ON sensor_command_history(sensor_id, property, sent_at DESC)
    WHERE status = 'sent';

INSERT OR IGNORE INTO permissions (name, description)
VALUES ('control_sensors', 'Send commands to controllable sensors');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'control_sensors'
WHERE r.name IN ('admin', 'user');
