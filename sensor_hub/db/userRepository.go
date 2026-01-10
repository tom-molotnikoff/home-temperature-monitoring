package database

import (
	"database/sql"
	"errors"
	"example/sensorHub/types"
	"fmt"
	"log"
	"time"
)

type SqlUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *SqlUserRepository {
	return &SqlUserRepository{db: db}
}

func (r *SqlUserRepository) CreateUser(user types.User, passwordHash string) (int, error) {
	query := "INSERT INTO users (username, email, password_hash, must_change_password, disabled, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	res, err := r.db.Exec(query, user.Username, user.Email, passwordHash, user.MustChangePassword, user.Disabled, time.Now())
	if err != nil {
		return 0, fmt.Errorf("error creating user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error fetching last insert id: %w", err)
	}
	for _, role := range user.Roles {
		if err := r.AssignRoleToUser(int(id), role); err != nil {
			log.Printf("warning: failed to assign role %s to user %d: %v", role, id, err)
		}
	}
	return int(id), nil
}

func (r *SqlUserRepository) GetUserByUsername(username string) (*types.User, string, error) {
	query := "SELECT id, username, email, must_change_password, disabled, created_at, updated_at, password_hash FROM users WHERE username = ?"
	var user types.User
	var passwordHash string
	var updatedAt sql.NullTime
	err := r.db.QueryRow(query, username).Scan(&user.Id, &user.Username, &user.Email, &user.MustChangePassword, &user.Disabled, &user.CreatedAt, &updatedAt, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("error querying user by username: %w", err)
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}
	roles, err := r.GetRolesForUser(user.Id)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching roles for user: %w", err)
	}
	user.Roles = roles
	return &user, passwordHash, nil
}

func (r *SqlUserRepository) GetUserById(id int) (*types.User, error) {
	query := "SELECT id, username, email, must_change_password, disabled, created_at, updated_at FROM users WHERE id = ?"
	var user types.User
	var updatedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(&user.Id, &user.Username, &user.Email, &user.MustChangePassword, &user.Disabled, &user.CreatedAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying user by id: %w", err)
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}
	roles, err := r.GetRolesForUser(user.Id)
	if err != nil {
		return nil, fmt.Errorf("error fetching roles for user: %w", err)
	}
	user.Roles = roles
	return &user, nil
}

func (r *SqlUserRepository) ListUsers() ([]types.User, error) {
	query := "SELECT id, username, email, must_change_password, disabled, created_at, updated_at FROM users"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []types.User
	for rows.Next() {
		var user types.User
		var updatedAt sql.NullTime
		if err := rows.Scan(&user.Id, &user.Username, &user.Email, &user.MustChangePassword, &user.Disabled, &user.CreatedAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.Time
		}
		roles, err := r.GetRolesForUser(user.Id)
		if err != nil {
			return nil, fmt.Errorf("error fetching roles for user: %w", err)
		}
		user.Roles = roles
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user rows: %w", err)
	}
	return users, nil
}

func (r *SqlUserRepository) UpdatePassword(userId int, passwordHash string, mustChange bool) error {
	query := "UPDATE users SET password_hash = ?, must_change_password = ?, updated_at = ? WHERE id = ?"
	_, err := r.db.Exec(query, passwordHash, mustChange, time.Now(), userId)
	if err != nil {
		return fmt.Errorf("error updating password for user: %w", err)
	}
	return nil
}

func (r *SqlUserRepository) SetDisabled(userId int, disabled bool) error {
	query := "UPDATE users SET disabled = ?, updated_at = ? WHERE id = ?"
	_, err := r.db.Exec(query, disabled, time.Now(), userId)
	if err != nil {
		return fmt.Errorf("error updating disabled flag for user: %w", err)
	}
	return nil
}

func (r *SqlUserRepository) AssignRoleToUser(userId int, roleName string) error {
	var roleId int
	err := r.db.QueryRow("SELECT id FROM roles WHERE name = ?", roleName).Scan(&roleId)
	if err != nil {
		return fmt.Errorf("error finding role %s: %w", roleName, err)
	}
	_, err = r.db.Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userId, roleId)
	if err != nil {
		return fmt.Errorf("error assigning role to user: %w", err)
	}
	return nil
}

func (r *SqlUserRepository) GetRolesForUser(userId int) ([]string, error) {
	query := "SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = ?"
	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying roles for user: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("error scanning role row: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over role rows: %w", err)
	}
	return roles, nil
}

func (r *SqlUserRepository) DeleteSessionsForUser(userId int) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE user_id = ?", userId)
	if err != nil {
		return fmt.Errorf("error deleting sessions for user: %w", err)
	}
	return nil
}

func (r *SqlUserRepository) DeleteSessionsForUserExcept(userId int, keepToken string) error {
	if keepToken == "" {
		return r.DeleteSessionsForUser(userId)
	}
	keepHash := tokenHash(keepToken)
	var cnt int
	err := r.db.QueryRow("SELECT COUNT(1) FROM sessions WHERE user_id = ? AND token_hash = ?", userId, keepHash).Scan(&cnt)
	if err != nil {
		return fmt.Errorf("error checking current session existence: %w", err)
	}
	if cnt == 0 {
		log.Printf("warning: requested to preserve session token for user %d but token not found; skipping deletion to avoid lockout", userId)
		return nil
	}
	res, err := r.db.Exec("DELETE FROM sessions WHERE user_id = ? AND token_hash != ?", userId, keepHash)
	if err != nil {
		return fmt.Errorf("error deleting sessions for user except token: %w", err)
	}
	if ra, err := res.RowsAffected(); err == nil {
		log.Printf("deleted %d sessions for user %d (preserved token hash %s)", ra, userId, keepHash)
	}
	return nil
}

func (r *SqlUserRepository) DeleteUserById(userId int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userId); err != nil {
		return fmt.Errorf("error deleting user roles: %w", err)
	}
	if _, err = tx.Exec("DELETE FROM sessions WHERE user_id = ?", userId); err != nil {
		return fmt.Errorf("error deleting sessions for user: %w", err)
	}
	if _, err = tx.Exec("DELETE FROM users WHERE id = ?", userId); err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing delete user transaction: %w", err)
	}
	return nil
}

func (r *SqlUserRepository) SetMustChangeFlag(userId int, mustChange bool) error {
	_, err := r.db.Exec("UPDATE users SET must_change_password = ?, updated_at = ? WHERE id = ?", mustChange, time.Now(), userId)
	if err != nil {
		return fmt.Errorf("error updating must_change_password: %w", err)
	}
	return nil
}

func (r *SqlUserRepository) SetRolesForUser(userId int, roles []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userId); err != nil {
		return fmt.Errorf("error clearing user roles: %w", err)
	}
	for _, role := range roles {
		var roleId int
		if err := tx.QueryRow("SELECT id FROM roles WHERE name = ?", role).Scan(&roleId); err != nil {
			return fmt.Errorf("error finding role %s: %w", role, err)
		}
		if _, err = tx.Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userId, roleId); err != nil {
			return fmt.Errorf("error assigning role %s to user: %w", role, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing set roles transaction: %w", err)
	}
	return nil
}
