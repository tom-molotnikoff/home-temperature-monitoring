-- V7: Failed login attempts tracking

CREATE TABLE IF NOT EXISTS failed_login_attempts (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(150) NULL,
  user_id INT NULL,
  ip_address VARCHAR(45) NULL,
  attempt_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  reason VARCHAR(255) NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
  INDEX (username),
  INDEX (user_id),
  INDEX (ip_address)
);

CREATE TABLE IF NOT EXISTS failed_login_summary (
  identifier VARCHAR(255) PRIMARY KEY,
  attempts INT NOT NULL DEFAULT 0,
  last_attempt_at TIMESTAMP NULL
);

