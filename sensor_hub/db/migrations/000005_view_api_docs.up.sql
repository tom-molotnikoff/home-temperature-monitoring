INSERT OR IGNORE INTO permissions (name, description) VALUES
    ('view_api_docs', 'View API documentation');

-- Grant to admin
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name = 'view_api_docs';

-- Grant to user role
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name = 'view_api_docs';

-- Grant to viewer role
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name = 'view_api_docs';
