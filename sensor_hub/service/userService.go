package service

import (
	"example/sensorHub/application_properties"
	"example/sensorHub/db"
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
}

func NewUserService(u database.UserRepository) *UserService {
	return &UserService{userRepo: u}
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
	return s.userRepo.DeleteUserById(userId)
}

func (s *UserService) SetMustChangeFlag(userId int, mustChange bool) error {
	return s.userRepo.SetMustChangeFlag(userId, mustChange)
}

func (s *UserService) SetUserRoles(userId int, roles []string) error {
	return s.userRepo.SetRolesForUser(userId, roles)
}
