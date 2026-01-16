package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type FailedLoginRepository interface {
	RecordFailedAttempt(username string, userId *int, ip string, reason string) error
	CountRecentFailedAttemptsByUsername(username string, window time.Duration) (int, error)
	CountRecentFailedAttemptsByIP(ip string, window time.Duration) (int, error)
	DeleteRecentFailedAttemptsByIP(ip string, window time.Duration) error
	DeleteAttemptsOlderThan(threshold time.Time) error
}

type SqlFailedLoginRepository struct {
	db *sql.DB
}

func NewFailedLoginRepository(db *sql.DB) *SqlFailedLoginRepository {
	return &SqlFailedLoginRepository{db: db}
}

func (r *SqlFailedLoginRepository) RecordFailedAttempt(username string, userId *int, ip string, reason string) error {
	log.Printf("Recording failed login attempt: username=%v, userId=%v, ip=%s, reason=%s", username, userId, ip, reason)
	query := "INSERT INTO failed_login_attempts (username, user_id, ip_address, attempt_time, reason) VALUES (?, ?, ?, ?, ?)"
	_, err := r.db.Exec(query, username, userId, ip, time.Now(), reason)
	if err != nil {
		return fmt.Errorf("error recording failed login attempt: %w", err)
	}
	return nil
}

func (r *SqlFailedLoginRepository) CountRecentFailedAttemptsByUsername(username string, window time.Duration) (int, error) {
	query := "SELECT COUNT(1) FROM failed_login_attempts WHERE username = ? AND attempt_time > ?"
	threshold := time.Now().Add(-window)
	var count int
	err := r.db.QueryRow(query, username, threshold).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting failed attempts by username: %w", err)
	}
	return count, nil
}

func (r *SqlFailedLoginRepository) CountRecentFailedAttemptsByIP(ip string, window time.Duration) (int, error) {
	query := "SELECT COUNT(1) FROM failed_login_attempts WHERE ip_address = ? AND attempt_time > ?"
	threshold := time.Now().Add(-window)
	var count int
	err := r.db.QueryRow(query, ip, threshold).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting failed attempts by ip: %w", err)
	}
	return count, nil
}

func (r *SqlFailedLoginRepository) DeleteRecentFailedAttemptsByIP(ip string, window time.Duration) error {
	query := "DELETE FROM failed_login_attempts WHERE ip_address = ? AND attempt_time > ?"
	threshold := time.Now().Add(-window)
	_, err := r.db.Exec(query, ip, threshold)
	if err != nil {
		return fmt.Errorf("error deleting failed attempts by ip: %w", err)
	}
	return nil
}

func (r *SqlFailedLoginRepository) DeleteAttemptsOlderThan(threshold time.Time) error {
	query := "DELETE FROM failed_login_attempts WHERE attempt_time < ?"
	_, err := r.db.Exec(query, threshold)
	if err != nil {
		return fmt.Errorf("error deleting old failed login attempts: %w", err)
	}
	return nil
}
