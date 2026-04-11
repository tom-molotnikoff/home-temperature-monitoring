-- Migration 000008: MQTT Ingest
-- Adds broker configuration, topic subscriptions, and sensor status for auto-discovery.

-- ============================================================
-- 1. MQTT Brokers
-- ============================================================
CREATE TABLE IF NOT EXISTS mqtt_brokers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK (type IN ('embedded', 'external')) DEFAULT 'external',
    host TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 1883,
    username TEXT DEFAULT NULL,
    password TEXT DEFAULT NULL,
    client_id TEXT DEFAULT NULL,
    ca_cert_path TEXT DEFAULT NULL,
    client_cert_path TEXT DEFAULT NULL,
    client_key_path TEXT DEFAULT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

-- Seed a default embedded broker so auto-discovery works out of the box.
INSERT OR IGNORE INTO mqtt_brokers (name, type, host, port, enabled)
VALUES ('Embedded Broker', 'embedded', 'localhost', 1883, 1);

-- ============================================================
-- 2. MQTT Subscriptions
-- ============================================================
CREATE TABLE IF NOT EXISTS mqtt_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    broker_id INTEGER NOT NULL,
    topic_pattern TEXT NOT NULL,
    driver_type TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (broker_id) REFERENCES mqtt_brokers(id) ON DELETE CASCADE,
    UNIQUE (broker_id, topic_pattern)
);

CREATE INDEX IF NOT EXISTS idx_mqtt_sub_broker ON mqtt_subscriptions (broker_id);
CREATE INDEX IF NOT EXISTS idx_mqtt_sub_driver ON mqtt_subscriptions (driver_type);

-- ============================================================
-- 3. Sensor status for auto-discovery
-- ============================================================
ALTER TABLE sensors ADD COLUMN status TEXT NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'pending', 'dismissed'));

CREATE INDEX IF NOT EXISTS idx_sensor_status ON sensors (status);

-- ============================================================
-- 4. MQTT permissions
-- ============================================================
INSERT OR IGNORE INTO permissions (name, description) VALUES
    ('view_mqtt', 'View MQTT broker and subscription configuration'),
    ('manage_mqtt', 'Create, update, and delete MQTT brokers and subscriptions');

-- Grant view_mqtt to all roles
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('view_mqtt', 'manage_mqtt');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name = 'view_mqtt';

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name = 'view_mqtt';
