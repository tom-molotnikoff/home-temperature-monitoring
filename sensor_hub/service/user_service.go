package service

import (
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/notifications"
	"example/sensorHub/types"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserServiceInterface interface {
	CreateUser(user types.User, plainPassword string) (int, error)
	ListUsers() ([]types.User, error)
	GetUserById(id int) (*types.User, error)
	ChangePassword(userId int, newPassword string, keepToken string) error
	DeleteUser(userId int) error
	SetMustChangeFlag(userId int, mustChange bool) error
	SetUserRoles(userId int, roles []string) error
}

type UserService struct {
	userRepo database.UserRepository
	notifSvc NotificationServiceInterface
}

func NewUserService(u database.UserRepository, n NotificationServiceInterface) *UserService {
	return &UserService{userRepo: u, notifSvc: n}
}

func (s *UserService) notifyUserEvent(action, username string, metadata map[string]interface{}) {
	if s.notifSvc == nil {
		return
	}
	notif := notifications.Notification{
		Category: notifications.CategoryUserManagement,
		Severity: notifications.SeverityInfo,
		Title:    fmt.Sprintf("User %s", action),
		Message:  fmt.Sprintf("User '%s' was %s", username, action),
		Metadata: metadata,
	}
	go s.notifSvc.CreateNotification(notif, "view_notifications_user_mgmt")
}

func (s *UserService) CreateUser(user types.User, plainPassword string) (int, error) {
	if plainPassword == "" {
		return 0, fmt.Errorf("password cannot be empty")
	}
	cost := 12
	if appProps.AppConfig != nil && appProps.AppConfig.AuthBcryptCost > 0 {
		cost = appProps.AppConfig.AuthBcryptCost
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), cost)
	if err != nil {
		return 0, err
	}
	user.MustChangePassword = true
	user.CreatedAt = time.Now()
	id, err := s.userRepo.CreateUser(user, string(hashBytes))
	if err != nil {
		return 0, err
	}
	for _, r := range user.Roles {
		err = s.userRepo.AssignRoleToUser(id, r)
		if err != nil {
			return 0, fmt.Errorf("failed to assign role %s to user: %w", r, err)
		}
	}
	s.notifyUserEvent("added", user.Username, map[string]interface{}{"user_id": id})
	return id, nil
}

func (s *UserService) ListUsers() ([]types.User, error) {
	return s.userRepo.ListUsers()
}

func (s *UserService) GetUserById(id int) (*types.User, error) {
	return s.userRepo.GetUserById(id)
}

func (s *UserService) ChangePassword(userId int, newPassword string, keepToken string) error {
	if newPassword == "" {
		return fmt.Errorf("password cannot be empty")
	}
	cost := 12
	if appProps.AppConfig != nil && appProps.AppConfig.AuthBcryptCost > 0 {
		cost = appProps.AppConfig.AuthBcryptCost
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), cost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(userId, string(hashBytes), false); err != nil {
		return err
	}
	if keepToken != "" {
		return s.userRepo.DeleteSessionsForUserExcept(userId, keepToken)
	}
	return s.userRepo.DeleteSessionsForUser(userId)
}

func (s *UserService) DeleteUser(userId int) error {
	user, _ := s.userRepo.GetUserById(userId)
	err := s.userRepo.DeleteUserById(userId)
	if err != nil {
		return err
	}
	if user != nil {
		s.notifyUserEvent("removed", user.Username, map[string]interface{}{"user_id": userId})
	}
	return nil
}

func (s *UserService) SetMustChangeFlag(userId int, mustChange bool) error {
	return s.userRepo.SetMustChangeFlag(userId, mustChange)
}

func (s *UserService) SetUserRoles(userId int, roles []string) error {
	err := s.userRepo.SetRolesForUser(userId, roles)
	if err != nil {
		return err
	}
	user, _ := s.userRepo.GetUserById(userId)
	if user != nil {
		s.notifyUserEvent("role changed", user.Username, map[string]interface{}{
			"user_id": userId,
			"roles":   roles,
		})
	}
	return nil
}
