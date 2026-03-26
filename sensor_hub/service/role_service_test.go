package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	database "example/sensorHub/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Test helpers
// ============================================================================

func setupRoleService() (*RoleService, *MockRoleRepository) {
	repo := new(MockRoleRepository)
	service := NewRoleService(repo, slog.Default())
	return service, repo
}

// ============================================================================
// ListRoles tests
// ============================================================================

func TestRoleService_ListRoles_Success(t *testing.T) {
	service, repo := setupRoleService()

	roles := []database.RoleInfo{
		{Id: 1, Name: "admin"},
		{Id: 2, Name: "user"},
	}
	repo.On("GetAllRoles", mock.Anything).Return(roles, nil)

	result, err := service.ListRoles(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "admin", result[0].Name)
	assert.Equal(t, "user", result[1].Name)
}

func TestRoleService_ListRoles_Empty(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("GetAllRoles", mock.Anything).Return([]database.RoleInfo{}, nil)

	result, err := service.ListRoles(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestRoleService_ListRoles_Error(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("GetAllRoles", mock.Anything).Return([]database.RoleInfo{}, errors.New("database error"))

	result, err := service.ListRoles(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Empty(t, result)
}

// ============================================================================
// ListPermissions tests
// ============================================================================

func TestRoleService_ListPermissions_Success(t *testing.T) {
	service, repo := setupRoleService()

	permissions := []database.PermissionInfo{
		{Id: 1, Name: "sensors:read"},
		{Id: 2, Name: "sensors:write"},
		{Id: 3, Name: "users:manage"},
	}
	repo.On("GetAllPermissions", mock.Anything).Return(permissions, nil)

	result, err := service.ListPermissions(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "sensors:read", result[0].Name)
}

func TestRoleService_ListPermissions_Empty(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("GetAllPermissions", mock.Anything).Return([]database.PermissionInfo{}, nil)

	result, err := service.ListPermissions(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestRoleService_ListPermissions_Error(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("GetAllPermissions", mock.Anything).Return([]database.PermissionInfo{}, errors.New("database error"))

	result, err := service.ListPermissions(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Empty(t, result)
}

// ============================================================================
// ListPermissionsForRole tests
// ============================================================================

func TestRoleService_ListPermissionsForRole_Success(t *testing.T) {
	service, repo := setupRoleService()

	permissions := []database.PermissionInfo{
		{Id: 1, Name: "sensors:read"},
		{Id: 2, Name: "sensors:write"},
	}
	repo.On("GetPermissionsForRole", mock.Anything, 1).Return(permissions, nil)

	result, err := service.ListPermissionsForRole(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRoleService_ListPermissionsForRole_Empty(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("GetPermissionsForRole", mock.Anything, 2).Return([]database.PermissionInfo{}, nil)

	result, err := service.ListPermissionsForRole(context.Background(), 2)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestRoleService_ListPermissionsForRole_Error(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("GetPermissionsForRole", mock.Anything, 99).Return([]database.PermissionInfo{}, errors.New("role not found"))

	result, err := service.ListPermissionsForRole(context.Background(), 99)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
	assert.Empty(t, result)
}

// ============================================================================
// AssignPermission tests
// ============================================================================

func TestRoleService_AssignPermission_Success(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("AssignPermissionToRole", mock.Anything, 1, 2).Return(nil)

	err := service.AssignPermission(context.Background(), 1, 2)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRoleService_AssignPermission_Error(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("AssignPermissionToRole", mock.Anything, 1, 999).Return(errors.New("permission not found"))

	err := service.AssignPermission(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission not found")
}

func TestRoleService_AssignPermission_DuplicateError(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("AssignPermissionToRole", mock.Anything, 1, 2).Return(errors.New("permission already assigned"))

	err := service.AssignPermission(context.Background(), 1, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission already assigned")
}

// ============================================================================
// RemovePermission tests
// ============================================================================

func TestRoleService_RemovePermission_Success(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("RemovePermissionFromRole", mock.Anything, 1, 2).Return(nil)

	err := service.RemovePermission(context.Background(), 1, 2)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRoleService_RemovePermission_Error(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("RemovePermissionFromRole", mock.Anything, 1, 999).Return(errors.New("permission not found"))

	err := service.RemovePermission(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission not found")
}

func TestRoleService_RemovePermission_NotAssigned(t *testing.T) {
	service, repo := setupRoleService()

	repo.On("RemovePermissionFromRole", mock.Anything, 1, 5).Return(errors.New("permission not assigned to role"))

	err := service.RemovePermission(context.Background(), 1, 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission not assigned to role")
}

// ============================================================================
// NewRoleService tests
// ============================================================================

func TestNewRoleService_ReturnsService(t *testing.T) {
	repo := new(MockRoleRepository)
	service := NewRoleService(repo, slog.Default())

	assert.NotNil(t, service)
}
