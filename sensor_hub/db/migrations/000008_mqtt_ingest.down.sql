-- Rollback migration 000008: MQTT Ingest

-- 1. Remove MQTT permissions
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('view_mqtt', 'manage_mqtt')
);
DELETE FROM permissions WHERE name IN ('view_mqtt', 'manage_mqtt');

-- 2. Drop sensor status index and column
DROP INDEX IF EXISTS idx_sensor_status;
ALTER TABLE sensors DROP COLUMN status;

-- 3. Drop MQTT tables
DROP TABLE IF EXISTS mqtt_subscriptions;
DROP TABLE IF EXISTS mqtt_brokers;
