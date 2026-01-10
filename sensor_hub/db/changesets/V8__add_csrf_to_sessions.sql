-- V8: Add csrf_token to sessions table

ALTER TABLE sessions
  ADD COLUMN csrf_token VARCHAR(128) NULL AFTER token_hash;

CREATE INDEX idx_sessions_csrf_token ON sessions (csrf_token);
