-- V11: additional role permissions

INSERT IGNORE INTO permissions (name, description) VALUES ('view_sensors', 'View sensor list and details');
INSERT IGNORE INTO permissions (name, description) VALUES ('delete_sensors', 'Delete sensors from the system');
INSERT IGNORE INTO permissions (name, description) VALUES ('view_roles', 'View role list and details');
INSERT IGNORE INTO permissions (name, description) VALUES ('manage_roles', 'Create and manage roles and permissions');
INSERT IGNORE INTO permissions (name, description) VALUES ('view_properties', 'View the list of application properties');
INSERT IGNORE INTO permissions (name, description) VALUES ('manage_properties', 'Update application properties');

-- Ensure admin role has all permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';
