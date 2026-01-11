-- V13: Alter description of 'manage_sensors' permission

INSERT IGNORE INTO permissions (name, description)
VALUES ('manage_sensors', 'Add, update and disable sensors (no deletion)');

UPDATE permissions SET description = 'Add, update and disable sensors (no deletion)' WHERE name = 'manage_sensors';

-- Ensure admin role has all permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';