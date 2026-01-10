package database

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"
)

type SessionRepository interface {
	CreateSession(userId int, rawToken string, expiresAt time.Time, ip string, userAgent string) (string, error) // returns csrfToken
	GetUserIdByToken(rawToken string) (int, error)
	GetSessionIdByToken(rawToken string) (int64, error)
	DeleteSessionByToken(rawToken string) error
	DeleteSessionsForUser(userId int) error
	ListSessionsForUser(userId int) ([]SessionInfo, error)
	RevokeSessionById(sessionId int64) error
	GetCSRFForToken(rawToken string) (string, error)
	InsertSessionAudit(sessionId int64, revokedByUserId *int, eventType string, reason *string) error
}

type SessionInfo struct {
	Id             int64     `json:"id"`
	UserId         int       `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	IpAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
}

type SqlSessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SqlSessionRepository {
	return &SqlSessionRepository{db: db}
}

func tokenHash(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generateCSRFToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (r *SqlSessionRepository) CreateSession(userId int, rawToken string, expiresAt time.Time, ip string, userAgent string) (string, error) {
	csrf, err := generateCSRFToken(24)
	if err != nil {
		return "", fmt.Errorf("failed to generate csrf token: %w", err)
	}
	query := "INSERT INTO sessions (user_id, token_hash, csrf_token, created_at, expires_at, last_accessed_at, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = r.db.Exec(query, userId, tokenHash(rawToken), csrf, time.Now(), expiresAt, time.Now(), ip, userAgent)
	if err != nil {
		return "", fmt.Errorf("error creating session: %w", err)
	}
	return csrf, nil
}

func (r *SqlSessionRepository) GetUserIdByToken(rawToken string) (int, error) {
	query := "SELECT user_id, expires_at FROM sessions WHERE token_hash = ?"
	var userId int
	var expiresAt time.Time
	err := r.db.QueryRow(query, tokenHash(rawToken)).Scan(&userId, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error querying session: %w", err)
	}
	if time.Now().After(expiresAt) {
		_ = r.DeleteSessionByToken(rawToken)
		return 0, nil
	}
	_, err = r.db.Exec("UPDATE sessions SET last_accessed_at = ? WHERE token_hash = ?", time.Now(), tokenHash(rawToken))
	if err != nil {
		log.Printf("error updating last accessed time: %v", err)
	}
	return userId, nil
}

func (r *SqlSessionRepository) GetSessionIdByToken(rawToken string) (int64, error) {
	query := "SELECT id, expires_at FROM sessions WHERE token_hash = ?"
	var id int64
	var expiresAt time.Time
	err := r.db.QueryRow(query, tokenHash(rawToken)).Scan(&id, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error querying session id: %w", err)
	}
	if time.Now().After(expiresAt) {
		_ = r.DeleteSessionByToken(rawToken)
		return 0, nil
	}
	return id, nil
}

func (r *SqlSessionRepository) DeleteSessionByToken(rawToken string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE token_hash = ?", tokenHash(rawToken))
	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}
	return nil
}

func (r *SqlSessionRepository) DeleteSessionsForUser(userId int) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE user_id = ?", userId)
	if err != nil {
		return fmt.Errorf("error deleting sessions for user: %w", err)
	}
	return nil
}

func (r *SqlSessionRepository) ListSessionsForUser(userId int) ([]SessionInfo, error) {
	rows, err := r.db.Query("SELECT id, user_id, created_at, expires_at, last_accessed_at, ip_address, user_agent FROM sessions WHERE user_id = ? ORDER BY created_at DESC", userId)
	if err != nil {
		return nil, fmt.Errorf("error querying sessions for user: %w", err)
	}
	defer rows.Close()
	var sessions []SessionInfo
	for rows.Next() {
		var s SessionInfo
		if err := rows.Scan(&s.Id, &s.UserId, &s.CreatedAt, &s.ExpiresAt, &s.LastAccessedAt, &s.IpAddress, &s.UserAgent); err != nil {
			return nil, fmt.Errorf("error scanning session row: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions rows: %w", err)
	}
	return sessions, nil
}

func (r *SqlSessionRepository) RevokeSessionById(sessionId int64) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE id = ?", sessionId)
	if err != nil {
		return fmt.Errorf("error revoking session: %w", err)
	}
	return nil
}

func (r *SqlSessionRepository) GetCSRFForToken(rawToken string) (string, error) {
	query := "SELECT csrf_token, expires_at FROM sessions WHERE token_hash = ?"
	var csrf sql.NullString
	var expiresAt time.Time
	err := r.db.QueryRow(query, tokenHash(rawToken)).Scan(&csrf, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("error querying csrf token: %w", err)
	}
	if time.Now().After(expiresAt) {
		_ = r.DeleteSessionByToken(rawToken)
		return "", nil
	}
	if csrf.Valid {
		return csrf.String, nil
	}
	return "", nil
}

func (r *SqlSessionRepository) InsertSessionAudit(sessionId int64, revokedByUserId *int, eventType string, reason *string) error {
	_, err := r.db.Exec("INSERT INTO session_audit (session_id, revoked_by_user_id, event_type, reason, created_at) VALUES (?, ?, ?, ?, ?)", sessionId, revokedByUserId, eventType, reason, time.Now())
	if err != nil {
		return fmt.Errorf("error inserting session audit: %w", err)
	}
	return nil
}
