-- Dashboard storage
CREATE TABLE IF NOT EXISTS dashboards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{"widgets":[],"breakpoints":{"lg":12,"md":8,"sm":4}}',
    shared INTEGER NOT NULL DEFAULT 0,
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_dashboards_user_id ON dashboards(user_id);

-- New permissions
INSERT OR IGNORE INTO permissions (name, description) VALUES
    ('manage_dashboards', 'Create, edit, delete, and share dashboards'),
    ('view_dashboards', 'View dashboards');

-- Grant both to admin
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('manage_dashboards', 'view_dashboards');

-- Grant both to user role
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('manage_dashboards', 'view_dashboards');

-- Grant view_dashboards to viewer role
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name = 'view_dashboards';
