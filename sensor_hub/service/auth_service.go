package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"log"
	"math"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	Login(username, password, ip, userAgent string) (rawToken string, csrfToken string, mustChange bool, err error)
	ValidateSession(rawToken string) (*types.User, error)
	Logout(rawToken string) error
	ChangePassword(userId int, newPassword string) error
	CreateInitialAdminIfNone(username, password string) error
	ListSessionsForUser(userId int) ([]database.SessionInfo, error)
	RevokeSessionById(sessionId int64) error
	RevokeSessionByIdWithActor(sessionId int64, revokedByUserId *int, reason *string) error
	GetCSRFForToken(rawToken string) (string, error)
	GetSessionIdForToken(rawToken string) (int64, error)
}

type AuthService struct {
	userRepo    database.UserRepository
	sessionRepo database.SessionRepository
	failedRepo  database.FailedLoginRepository
	roleRepo    database.RoleRepository
}

func NewAuthService(u database.UserRepository, s database.SessionRepository, f database.FailedLoginRepository, r database.RoleRepository) *AuthService {
	return &AuthService{userRepo: u, sessionRepo: s, failedRepo: f, roleRepo: r}
}

func (a *AuthService) generateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (a *AuthService) bcryptHash(password string) (string, error) {
	cost := 12
	if appProps.AppConfig != nil && appProps.AppConfig.AuthBcryptCost > 0 {
		cost = appProps.AppConfig.AuthBcryptCost
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type TooManyAttemptsError struct {
	RetryAfterSeconds int
	FailedByUser      int
	FailedByIP        int
	Threshold         int
	Exponent          int
}

func (e *TooManyAttemptsError) Error() string {
	return "too many failed login attempts"
}

func (a *AuthService) Login(username, password, ip, userAgent string) (string, string, bool, error) {
	windowMinutes := 15
	threshold := 5
	baseSeconds := 2
	maxSeconds := 300
	if appProps.AppConfig != nil {
		if appProps.AppConfig.AuthLoginBackoffWindowMinutes > 0 {
			windowMinutes = appProps.AppConfig.AuthLoginBackoffWindowMinutes
		}
		if appProps.AppConfig.AuthLoginBackoffThreshold > 0 {
			threshold = appProps.AppConfig.AuthLoginBackoffThreshold
		}
		if appProps.AppConfig.AuthLoginBackoffBaseSeconds > 0 {
			baseSeconds = appProps.AppConfig.AuthLoginBackoffBaseSeconds
		}
		if appProps.AppConfig.AuthLoginBackoffMaxSeconds > 0 {
			maxSeconds = appProps.AppConfig.AuthLoginBackoffMaxSeconds
		}
	}

	ipAllowedOnce := ipBlocker.consumeAllowOnceIfReady("ip:" + ip)
	userAllowedOnce := userBlocker.consumeAllowOnceIfReady("user:" + username)
	if ipAllowedOnce {
		log.Printf("allowing one post-block login attempt for ip=%s (allow-once consumed)", ip)
	}
	if userAllowedOnce {
		log.Printf("allowing one post-block login attempt for user=%s (allow-once consumed)", username)
	}

	if remaining := ipBlocker.getRemainingSeconds("ip:" + ip); remaining > 0 {
		if !ipAllowedOnce {
			failedByUser := 0
			if c, err := a.failedRepo.CountRecentFailedAttemptsByUsername(username, time.Duration(windowMinutes)*time.Minute); err == nil {
				failedByUser = c
			}
			failedByIP := 0
			if c, err := a.failedRepo.CountRecentFailedAttemptsByIP(ip, time.Duration(windowMinutes)*time.Minute); err == nil {
				failedByIP = c
			}
			expiresAt := time.Now().Add(time.Duration(remaining) * time.Second).UTC().Format(time.RFC3339)
			log.Printf("login blocked by in-memory ipBlocker until %s (remaining=%ds) (failedByUser=%d, failedByIP=%d)", expiresAt, remaining, failedByUser, failedByIP)
			return "", "", false, &TooManyAttemptsError{RetryAfterSeconds: remaining, FailedByUser: failedByUser, FailedByIP: failedByIP}
		}
	}
	if remaining := userBlocker.getRemainingSeconds("user:" + username); remaining > 0 {
		if !userAllowedOnce {
			failedByUser := 0
			if c, err := a.failedRepo.CountRecentFailedAttemptsByUsername(username, time.Duration(windowMinutes)*time.Minute); err == nil {
				failedByUser = c
			}
			failedByIP := 0
			if c, err := a.failedRepo.CountRecentFailedAttemptsByIP(ip, time.Duration(windowMinutes)*time.Minute); err == nil {
				failedByIP = c
			}
			expiresAt := time.Now().Add(time.Duration(remaining) * time.Second).UTC().Format(time.RFC3339)
			log.Printf("login blocked by in-memory userBlocker until %s (remaining=%ds) (failedByUser=%d, failedByIP=%d)", expiresAt, remaining, failedByUser, failedByIP)
			return "", "", false, &TooManyAttemptsError{RetryAfterSeconds: remaining, FailedByUser: failedByUser, FailedByIP: failedByIP}
		}
	}

	failedByUser := 0
	if c, err := a.failedRepo.CountRecentFailedAttemptsByUsername(username, time.Duration(windowMinutes)*time.Minute); err == nil {
		failedByUser = c
	}
	failedByIP := 0
	if c, err := a.failedRepo.CountRecentFailedAttemptsByIP(ip, time.Duration(windowMinutes)*time.Minute); err == nil {
		failedByIP = c
	}
	failedCount := failedByUser
	if failedByIP > failedCount {
		failedCount = failedByIP
	}
	if failedCount >= threshold {
		exponent := failedCount - threshold
		d := float64(baseSeconds) * math.Pow(2, float64(exponent))
		delay := int(math.Min(float64(maxSeconds), d))
		minBlock := 30
		finalDelay := delay
		if finalDelay < minBlock {
			finalDelay = minBlock
		}
		if !(ipAllowedOnce || userAllowedOnce) {
			expiresAt := time.Now().Add(time.Duration(finalDelay) * time.Second).UTC().Format(time.RFC3339)
			log.Printf("blocking login for %d seconds until %s due to previous failed attempts (failedByUser=%d, failedByIP=%d, threshold=%d, exponent=%d)", finalDelay, expiresAt, failedByUser, failedByIP, threshold, exponent)
			ipBlocker.blockFor("ip:"+ip, finalDelay)
			userBlocker.blockFor("user:"+username, finalDelay)
			return "", "", false, &TooManyAttemptsError{RetryAfterSeconds: finalDelay, FailedByUser: failedByUser, FailedByIP: failedByIP, Threshold: threshold, Exponent: exponent}
		}
		log.Printf("allow-once consumed; proceeding to credential check despite failedCount=%d (threshold=%d)", failedCount, threshold)
	}

	user, passwordHash, err := a.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", "", false, err
	}
	if user == nil {
		err = a.failedRepo.RecordFailedAttempt(username, nil, ip, "no_such_user")
		if err != nil {
			log.Printf("error recording failed login attempt: %v", err)
		}
		return "", "", false, errors.New("invalid credentials")
	}
	if user.Disabled {
		return "", "", false, errors.New("account disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		err = a.failedRepo.RecordFailedAttempt(username, &user.Id, ip, "bad_password")
		if err != nil {
			log.Printf("error recording failed login attempt: %v", err)
		}
		return "", "", false, errors.New("invalid credentials")
	}

	token, err := a.generateToken(32)
	if err != nil {
		log.Printf("error generating session token: %v", err)
		return "", "", false, err
	}

	ttlMinutes := 60 * 24 * 30 // 30 days default
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionTTLMinutes > 0 {
		ttlMinutes = appProps.AppConfig.AuthSessionTTLMinutes
	}
	expires := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	csrf, err := a.sessionRepo.CreateSession(user.Id, token, expires, ip, userAgent)
	if err != nil {
		log.Printf("error creating session: %v", err)
		return "", "", false, err
	}

	ipBlocker.forceClearAllowOnce("ip:" + ip)
	userBlocker.forceClearAllowOnce("user:" + username)

	if appProps.AppConfig != nil {
		windowMinutes := appProps.AppConfig.AuthLoginBackoffWindowMinutes
		if windowMinutes <= 0 {
			windowMinutes = 15
		}
		if err := a.failedRepo.DeleteRecentFailedAttemptsByIP(ip, time.Duration(windowMinutes)*time.Minute); err != nil {
			log.Printf("error clearing failed login attempts for ip %s: %v", ip, err)
		}
	} else {
		if err := a.failedRepo.DeleteRecentFailedAttemptsByIP(ip, time.Duration(15)*time.Minute); err != nil {
			log.Printf("error clearing failed login attempts for ip %s: %v", ip, err)
		}
	}
	return token, csrf, user.MustChangePassword, nil
}

func (a *AuthService) ValidateSession(rawToken string) (*types.User, error) {
	userId, err := a.sessionRepo.GetUserIdByToken(rawToken)
	if err != nil {
		return nil, err
	}
	if userId == 0 {
		return nil, nil
	}
	user, err := a.userRepo.GetUserById(userId)
	if err != nil {
		return nil, err
	}
	if a.roleRepo != nil {
		perms, err := a.roleRepo.GetPermissionsForUser(user.Id)
		if err == nil {
			user.Permissions = perms
		}
	}
	return user, nil
}

func (a *AuthService) Logout(rawToken string) error {
	return a.sessionRepo.DeleteSessionByToken(rawToken)
}

func (a *AuthService) ChangePassword(userId int, newPassword string) error {
	hash, err := a.bcryptHash(newPassword)
	if err != nil {
		return err
	}
	return a.userRepo.UpdatePassword(userId, hash, false)
}

func (a *AuthService) CreateInitialAdminIfNone(username, password string) error {
	users, err := a.userRepo.ListUsers()
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return nil // already users present
	}
	hash, err := a.bcryptHash(password)
	if err != nil {
		return err
	}
	user := types.User{Username: username, Email: "", Disabled: false, MustChangePassword: true, Roles: []string{types.RoleAdmin}}
	id, err := a.userRepo.CreateUser(user, hash)
	if err != nil {
		return err
	}
	err = a.userRepo.AssignRoleToUser(id, types.RoleAdmin)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthService) ListSessionsForUser(userId int) ([]database.SessionInfo, error) {
	return a.sessionRepo.ListSessionsForUser(userId)
}

func (a *AuthService) RevokeSessionByIdWithActor(sessionId int64, revokedByUserId *int, reason *string) error {
	err := a.sessionRepo.InsertSessionAudit(sessionId, revokedByUserId, "revoked", reason)
	if err != nil {
		log.Printf("error inserting session audit record: %v", err)
	}
	return a.sessionRepo.RevokeSessionById(sessionId)
}

func (a *AuthService) RevokeSessionById(sessionId int64) error {
	return a.RevokeSessionByIdWithActor(sessionId, nil, nil)
}

func (a *AuthService) GetCSRFForToken(rawToken string) (string, error) {
	return a.sessionRepo.GetCSRFForToken(rawToken)
}

func (a *AuthService) GetSessionIdForToken(rawToken string) (int64, error) {
	return a.sessionRepo.GetSessionIdByToken(rawToken)
}
