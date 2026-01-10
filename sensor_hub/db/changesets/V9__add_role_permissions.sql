-- V9: role_permissions mapping table

CREATE TABLE IF NOT EXISTS permissions (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  description VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id INT NOT NULL,
  permission_id INT NOT NULL,
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

INSERT IGNORE INTO permissions (name, description) VALUES ('manage_users', 'Create and manage users');
INSERT IGNORE INTO permissions (name, description) VALUES ('view_users', 'View user list and details');
INSERT IGNORE INTO permissions (name, description) VALUES ('manage_sensors', 'Add/update/delete sensors');
INSERT IGNORE INTO permissions (name, description) VALUES ('view_readings', 'View sensor readings');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';

