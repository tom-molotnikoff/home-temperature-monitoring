package service

import (
	"context"
	database "example/sensorHub/db"
	"log/slog"
)

type RoleServiceInterface interface {
	ListRoles(ctx context.Context) ([]database.RoleInfo, error)
	ListPermissions(ctx context.Context) ([]database.PermissionInfo, error)
	ListPermissionsForRole(ctx context.Context, roleId int) ([]database.PermissionInfo, error)
	AssignPermission(ctx context.Context, roleId int, permissionId int) error
	RemovePermission(ctx context.Context, roleId int, permissionId int) error
}

type RoleService struct {
	repo   database.RoleRepository
	logger *slog.Logger
}

func NewRoleService(r database.RoleRepository, logger *slog.Logger) *RoleService {
	return &RoleService{repo: r, logger: logger.With("component", "role_service")}
}

func (s *RoleService) ListRoles(ctx context.Context) ([]database.RoleInfo, error) {
	return s.repo.GetAllRoles(ctx)
}

func (s *RoleService) ListPermissions(ctx context.Context) ([]database.PermissionInfo, error) {
	return s.repo.GetAllPermissions(ctx)
}

func (s *RoleService) ListPermissionsForRole(ctx context.Context, roleId int) ([]database.PermissionInfo, error) {
	return s.repo.GetPermissionsForRole(ctx, roleId)
}

func (s *RoleService) AssignPermission(ctx context.Context, roleId int, permissionId int) error {
	return s.repo.AssignPermissionToRole(ctx, roleId, permissionId)
}

func (s *RoleService) RemovePermission(ctx context.Context, roleId int, permissionId int) error {
	return s.repo.RemovePermissionFromRole(ctx, roleId, permissionId)
}
