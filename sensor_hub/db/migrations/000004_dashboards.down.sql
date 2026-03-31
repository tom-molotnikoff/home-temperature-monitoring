-- Remove role_permissions for dashboard permissions
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('manage_dashboards', 'view_dashboards')
);

-- Remove permissions
DELETE FROM permissions WHERE name IN ('manage_dashboards', 'view_dashboards');

-- Drop table
DROP TABLE IF EXISTS dashboards;
