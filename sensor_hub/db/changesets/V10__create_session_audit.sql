-- V10: Create session_audit table to record session revocations and other events

CREATE TABLE IF NOT EXISTS session_audit (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT NOT NULL,
  revoked_by_user_id INT NULL,
  event_type VARCHAR(50) NOT NULL,
  reason VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX idx_session_audit_session_id ON session_audit(session_id);

