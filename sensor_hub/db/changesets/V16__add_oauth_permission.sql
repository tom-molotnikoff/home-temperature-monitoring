-- V16: Add manage_oauth permission for OAuth configuration management

INSERT IGNORE INTO permissions (name, description) VALUES ('manage_oauth', 'Manage OAuth credentials and re-authorize');

-- Grant manage_oauth permission to admin role
INSERT IGNORE INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin' AND p.name = 'manage_oauth';
