DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE name = 'manage_api_keys');
DELETE FROM permissions WHERE name = 'manage_api_keys';
