package service

import (
	database "example/sensorHub/db"
)

type RoleServiceInterface interface {
	ListRoles() ([]database.RoleInfo, error)
	ListPermissions() ([]database.PermissionInfo, error)
	ListPermissionsForRole(roleId int) ([]database.PermissionInfo, error)
	AssignPermission(roleId int, permissionId int) error
	RemovePermission(roleId int, permissionId int) error
}

type RoleService struct {
	repo database.RoleRepository
}

func NewRoleService(r database.RoleRepository) *RoleService {
	return &RoleService{repo: r}
}

func (s *RoleService) ListRoles() ([]database.RoleInfo, error) {
	return s.repo.GetAllRoles()
}

func (s *RoleService) ListPermissions() ([]database.PermissionInfo, error) {
	return s.repo.GetAllPermissions()
}

func (s *RoleService) ListPermissionsForRole(roleId int) ([]database.PermissionInfo, error) {
	return s.repo.GetPermissionsForRole(roleId)
}

func (s *RoleService) AssignPermission(roleId int, permissionId int) error {
	return s.repo.AssignPermissionToRole(roleId, permissionId)
}

func (s *RoleService) RemovePermission(roleId int, permissionId int) error {
	return s.repo.RemovePermissionFromRole(roleId, permissionId)
}
