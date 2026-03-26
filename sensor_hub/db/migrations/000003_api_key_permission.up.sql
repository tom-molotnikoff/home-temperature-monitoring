INSERT INTO permissions (name, description) VALUES ('manage_api_keys', 'Create and manage API keys');
INSERT INTO role_permissions (role_id, permission_id) SELECT r.id, p.id FROM roles r, permissions p WHERE p.name = 'manage_api_keys';
