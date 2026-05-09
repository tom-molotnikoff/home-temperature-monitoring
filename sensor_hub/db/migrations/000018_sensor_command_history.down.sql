DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name = 'control_sensors'
);

DELETE FROM permissions
WHERE name = 'control_sensors';

DROP INDEX IF EXISTS idx_sensor_command_history_pending;
DROP INDEX IF EXISTS idx_sensor_command_history_sensor_id;
DROP TABLE IF EXISTS sensor_command_history;
