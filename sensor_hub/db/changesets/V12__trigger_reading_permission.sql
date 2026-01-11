-- V12: Trigger reading permission

INSERT IGNORE INTO permissions (name, description) VALUES ('trigger_readings', 'Trigger sensor readings on demand');

-- Ensure admin role has all permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';
