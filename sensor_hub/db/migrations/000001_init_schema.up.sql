-- Consolidated schema for sensor_hub (SQLite)
-- Equivalent to MySQL Flyway migrations V1–V18, merged into final-state tables.

-- Sensors
CREATE TABLE IF NOT EXISTS sensors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    type TEXT DEFAULT NULL,
    url TEXT DEFAULT NULL,
    health_status TEXT DEFAULT 'unknown',
    health_reason TEXT DEFAULT NULL,
    enabled INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_sensor_type ON sensors (type);

-- Temperature readings (raw)
CREATE TABLE IF NOT EXISTS temperature_readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    time TEXT NOT NULL,
    temperature REAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id)
);

CREATE INDEX IF NOT EXISTS idx_time ON temperature_readings (time DESC);
CREATE INDEX IF NOT EXISTS idx_sensor_id ON temperature_readings (sensor_id);

-- Hourly averaged temperature readings
CREATE TABLE IF NOT EXISTS hourly_avg_temperature (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    time TEXT NOT NULL,
    average_temperature REAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id),
    UNIQUE (sensor_id, time)
);

CREATE INDEX IF NOT EXISTS hourly_idx_time ON hourly_avg_temperature (time DESC);
CREATE INDEX IF NOT EXISTS hourly_idx_sensor_id ON hourly_avg_temperature (sensor_id);

-- Sensor health history
CREATE TABLE IF NOT EXISTS sensor_health_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    health_status TEXT NOT NULL,
    recorded_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (sensor_id) REFERENCES sensors(id)
);

-- Roles
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TEXT DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO roles (name, description) VALUES ('admin', 'Full administrative access');
INSERT OR IGNORE INTO roles (name, description) VALUES ('user', 'Standard logged-in user');
INSERT OR IGNORE INTO roles (name, description) VALUES ('viewer', 'Read-only access');

-- Users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT,
    password_hash TEXT NOT NULL,
    must_change_password INTEGER NOT NULL DEFAULT 1,
    disabled INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT NULL
);

-- User ↔ Role mapping
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Sessions (token hash stored, not raw token)
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL,
    csrf_token TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL,
    last_accessed_at TEXT DEFAULT (datetime('now')),
    ip_address TEXT,
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions (token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_csrf_token ON sessions (csrf_token);

-- Session audit trail
CREATE TABLE IF NOT EXISTS session_audit (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    revoked_by_user_id INTEGER,
    event_type TEXT NOT NULL,
    reason TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_session_audit_session_id ON session_audit (session_id);

-- Failed login tracking
CREATE TABLE IF NOT EXISTS failed_login_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT,
    user_id INTEGER,
    ip_address TEXT,
    attempt_time TEXT DEFAULT (datetime('now')),
    reason TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_failed_login_username ON failed_login_attempts (username);
CREATE INDEX IF NOT EXISTS idx_failed_login_user_id ON failed_login_attempts (user_id);
CREATE INDEX IF NOT EXISTS idx_failed_login_ip ON failed_login_attempts (ip_address);

CREATE TABLE IF NOT EXISTS failed_login_summary (
    identifier TEXT PRIMARY KEY,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TEXT
);

-- Permissions
CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

INSERT OR IGNORE INTO permissions (name, description) VALUES ('manage_users', 'Create and manage users');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_users', 'View user list and details');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('manage_sensors', 'Add, update and disable sensors (no deletion)');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_readings', 'View sensor readings');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_sensors', 'View sensor list and details');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('delete_sensors', 'Delete sensors from the system');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_roles', 'View role list and details');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('manage_roles', 'Create and manage roles and permissions');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_properties', 'View the list of application properties');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('manage_properties', 'Update application properties');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('trigger_readings', 'Trigger sensor readings on demand');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_alerts', 'View sensor alert rules and history');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('manage_alerts', 'Create, update, and delete alert rules');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('manage_oauth', 'Manage OAuth credentials and re-authorize');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_notifications', 'Access and view in-app notifications');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_notifications_user_mgmt', 'Receive user management notifications');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('view_notifications_config', 'Receive configuration change notifications');
INSERT OR IGNORE INTO permissions (name, description) VALUES ('manage_notifications', 'Configure notification channel preferences');

-- Role ↔ Permission mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Grant admin role all permissions
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';

-- Sensor alert rules
CREATE TABLE IF NOT EXISTS sensor_alert_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    high_threshold REAL,
    low_threshold REAL,
    trigger_status TEXT,
    enabled INTEGER NOT NULL DEFAULT 1,
    rate_limit_hours INTEGER NOT NULL DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alert_sensor_id ON sensor_alert_rules (sensor_id);
CREATE INDEX IF NOT EXISTS idx_alert_enabled ON sensor_alert_rules (enabled);

-- Alert sent history (for rate limiting)
CREATE TABLE IF NOT EXISTS alert_sent_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alert_rule_id INTEGER NOT NULL,
    sensor_id INTEGER NOT NULL,
    sent_at TEXT DEFAULT (datetime('now')),
    alert_reason TEXT,
    reading_value REAL,
    reading_status TEXT,
    FOREIGN KEY (alert_rule_id) REFERENCES sensor_alert_rules(id) ON DELETE CASCADE,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alert_rule_sent_at ON alert_sent_history (alert_rule_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_sensor_sent_at ON alert_sent_history (sensor_id, sent_at DESC);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT NOT NULL,
    severity TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    metadata TEXT,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_category ON notifications (category);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at);

-- User ↔ Notification mapping
CREATE TABLE IF NOT EXISTS user_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    notification_id INTEGER NOT NULL,
    is_read INTEGER DEFAULT 0,
    is_dismissed INTEGER DEFAULT 0,
    read_at TEXT,
    dismissed_at TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    UNIQUE (user_id, notification_id)
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_unread ON user_notifications (user_id, is_read, is_dismissed);

-- Notification channel defaults
CREATE TABLE IF NOT EXISTS notification_channel_defaults (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT NOT NULL UNIQUE,
    email_enabled INTEGER DEFAULT 1,
    inapp_enabled INTEGER DEFAULT 1
);

INSERT OR IGNORE INTO notification_channel_defaults (category, email_enabled, inapp_enabled) VALUES
('threshold_alert', 1, 1),
('user_management', 0, 1),
('config_change', 0, 1);

-- Per-user notification channel preferences
CREATE TABLE IF NOT EXISTS notification_channel_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    category TEXT NOT NULL,
    email_enabled INTEGER NOT NULL,
    inapp_enabled INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id, category)
);
