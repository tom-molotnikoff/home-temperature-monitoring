package database

import (
	"context"

	"example/sensorHub/types"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user types.User, passwordHash string) (int, error)
	GetUserByUsername(ctx context.Context, username string) (*types.User, string, error) // returns user and passwordHash
	GetUserById(ctx context.Context, id int) (*types.User, error)
	ListUsers(ctx context.Context) ([]types.User, error)
	UpdatePassword(ctx context.Context, userId int, passwordHash string, mustChange bool) error
	SetDisabled(ctx context.Context, userId int, disabled bool) error
	AssignRoleToUser(ctx context.Context, userId int, roleName string) error
	GetRolesForUser(ctx context.Context, userId int) ([]string, error)
	DeleteSessionsForUser(ctx context.Context, userId int) error
	DeleteSessionsForUserExcept(ctx context.Context, userId int, keepToken string) error
	DeleteUserById(ctx context.Context, userId int) error
	SetMustChangeFlag(ctx context.Context, userId int, mustChange bool) error
	SetRolesForUser(ctx context.Context, userId int, roles []string) error
}
