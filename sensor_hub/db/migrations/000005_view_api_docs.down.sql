DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE name = 'view_api_docs');
DELETE FROM permissions WHERE name = 'view_api_docs';
