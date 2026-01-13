-- V15: Alert management permissions

INSERT IGNORE INTO permissions (name, description) VALUES ('view_alerts', 'View sensor alert rules and history');
INSERT IGNORE INTO permissions (name, description) VALUES ('manage_alerts', 'Create, update, and delete alert rules');

-- Ensure admin role has all permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';
