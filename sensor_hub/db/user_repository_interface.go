package database

import "example/sensorHub/types"

type UserRepository interface {
	CreateUser(user types.User, passwordHash string) (int, error)
	GetUserByUsername(username string) (*types.User, string, error) // returns user and passwordHash
	GetUserById(id int) (*types.User, error)
	ListUsers() ([]types.User, error)
	UpdatePassword(userId int, passwordHash string, mustChange bool) error
	SetDisabled(userId int, disabled bool) error
	AssignRoleToUser(userId int, roleName string) error
	GetRolesForUser(userId int) ([]string, error)
	DeleteSessionsForUser(userId int) error
	DeleteSessionsForUserExcept(userId int, keepToken string) error
	DeleteUserById(userId int) error
	SetMustChangeFlag(userId int, mustChange bool) error
	SetRolesForUser(userId int, roles []string) error
}
