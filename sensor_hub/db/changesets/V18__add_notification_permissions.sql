-- V18: Notification permissions and channel preferences

INSERT IGNORE INTO permissions (name, description) VALUES 
('view_notifications', 'Access and view in-app notifications'),
('view_notifications_user_mgmt', 'Receive user management notifications'),
('view_notifications_config', 'Receive configuration change notifications'),
('manage_notifications', 'Configure notification channel preferences');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.name = 'admin' AND p.name IN (
    'view_notifications', 
    'view_notifications_user_mgmt', 
    'view_notifications_config', 
    'manage_notifications'
);

CREATE TABLE notification_channel_defaults (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category VARCHAR(50) NOT NULL UNIQUE,
    email_enabled BOOLEAN DEFAULT TRUE,
    inapp_enabled BOOLEAN DEFAULT TRUE
);

INSERT INTO notification_channel_defaults (category, email_enabled, inapp_enabled) VALUES
('threshold_alert', TRUE, TRUE),
('user_management', FALSE, TRUE),
('config_change', FALSE, TRUE);

CREATE TABLE notification_channel_preferences (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    category VARCHAR(50) NOT NULL,
    email_enabled BOOLEAN NOT NULL,
    inapp_enabled BOOLEAN NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_category_pref (user_id, category)
);
