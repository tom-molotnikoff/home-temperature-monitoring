package service

import (
	"errors"
	"testing"

	appProps "example/sensorHub/application_properties"
	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Test helpers
// ============================================================================

func setupUserService() (*UserService, *MockUserRepository) {
	userRepo := new(MockUserRepository)
	service := NewUserService(userRepo, nil)
	return service, userRepo
}

// ============================================================================
// CreateUser tests
// ============================================================================

func TestUserService_CreateUser_Success(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	user := types.User{Username: "newuser", Email: "test@test.com", Roles: []string{"user"}}

	userRepo.On("CreateUser", mock.Anything, mock.Anything).Return(1, nil)
	userRepo.On("AssignRoleToUser", 1, "user").Return(nil)

	id, err := service.CreateUser(user, "password123")

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	userRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmptyPassword(t *testing.T) {
	service, _ := setupUserService()

	user := types.User{Username: "newuser"}

	id, err := service.CreateUser(user, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
	assert.Equal(t, 0, id)
}

func TestUserService_CreateUser_MultipleRoles(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	user := types.User{Username: "admin", Roles: []string{"admin", "user"}}

	userRepo.On("CreateUser", mock.Anything, mock.Anything).Return(1, nil)
	userRepo.On("AssignRoleToUser", 1, "admin").Return(nil)
	userRepo.On("AssignRoleToUser", 1, "user").Return(nil)

	id, err := service.CreateUser(user, "password123")

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	userRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_DBError(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	user := types.User{Username: "newuser"}

	userRepo.On("CreateUser", mock.Anything, mock.Anything).Return(0, errors.New("database error"))

	id, err := service.CreateUser(user, "password123")

	assert.Error(t, err)
	assert.Equal(t, 0, id)
}

func TestUserService_CreateUser_RoleAssignmentError(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	user := types.User{Username: "newuser", Roles: []string{"admin"}}

	userRepo.On("CreateUser", mock.Anything, mock.Anything).Return(1, nil)
	userRepo.On("AssignRoleToUser", 1, "admin").Return(errors.New("role not found"))

	id, err := service.CreateUser(user, "password123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign role")
	assert.Equal(t, 0, id)
}

func TestUserService_CreateUser_SetsMustChangePassword(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	user := types.User{Username: "newuser", MustChangePassword: false}

	userRepo.On("CreateUser", mock.MatchedBy(func(u types.User) bool {
		return u.MustChangePassword == true
	}), mock.Anything).Return(1, nil)

	id, err := service.CreateUser(user, "password123")

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	userRepo.AssertExpectations(t)
}

// ============================================================================
// ListUsers tests
// ============================================================================

func TestUserService_ListUsers_Success(t *testing.T) {
	service, userRepo := setupUserService()

	users := []types.User{
		{Id: 1, Username: "user1"},
		{Id: 2, Username: "user2"},
	}
	userRepo.On("ListUsers").Return(users, nil)

	result, err := service.ListUsers()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUserService_ListUsers_Empty(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("ListUsers").Return([]types.User{}, nil)

	result, err := service.ListUsers()

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestUserService_ListUsers_Error(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("ListUsers").Return([]types.User{}, errors.New("database error"))

	_, err := service.ListUsers()

	assert.Error(t, err)
}

// ============================================================================
// GetUserById tests
// ============================================================================

func TestUserService_GetUserById_Success(t *testing.T) {
	service, userRepo := setupUserService()

	user := &types.User{Id: 1, Username: "testuser"}
	userRepo.On("GetUserById", 1).Return(user, nil)

	result, err := service.GetUserById(1)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", result.Username)
}

func TestUserService_GetUserById_NotFound(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("GetUserById", 999).Return(nil, nil)

	result, err := service.GetUserById(999)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestUserService_GetUserById_Error(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("GetUserById", 1).Return(nil, errors.New("database error"))

	_, err := service.GetUserById(1)

	assert.Error(t, err)
}

// ============================================================================
// ChangePassword tests
// ============================================================================

func TestUserService_ChangePassword_Success(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	userRepo.On("UpdatePassword", 1, mock.Anything, false).Return(nil)
	userRepo.On("DeleteSessionsForUser", 1).Return(nil)

	err := service.ChangePassword(1, "newpassword", "")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_EmptyPassword(t *testing.T) {
	service, _ := setupUserService()

	err := service.ChangePassword(1, "", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestUserService_ChangePassword_WithKeepToken(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	userRepo.On("UpdatePassword", 1, mock.Anything, false).Return(nil)
	userRepo.On("DeleteSessionsForUserExcept", 1, "keep-this-token").Return(nil)

	err := service.ChangePassword(1, "newpassword", "keep-this-token")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	userRepo.AssertNotCalled(t, "DeleteSessionsForUser")
}

func TestUserService_ChangePassword_UpdateError(t *testing.T) {
	defer setupTestConfig()()

	service, userRepo := setupUserService()

	userRepo.On("UpdatePassword", 1, mock.Anything, false).Return(errors.New("database error"))

	err := service.ChangePassword(1, "newpassword", "")

	assert.Error(t, err)
}

// ============================================================================
// DeleteUser tests
// ============================================================================

func TestUserService_DeleteUser_Success(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("GetUserById", 1).Return(&types.User{Id: 1, Username: "testuser"}, nil)
	userRepo.On("DeleteUserById", 1).Return(nil)

	err := service.DeleteUser(1)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Error(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("GetUserById", 1).Return(&types.User{Id: 1, Username: "testuser"}, nil)
	userRepo.On("DeleteUserById", 1).Return(errors.New("database error"))

	err := service.DeleteUser(1)

	assert.Error(t, err)
}

// ============================================================================
// SetMustChangeFlag tests
// ============================================================================

func TestUserService_SetMustChangeFlag_True(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("SetMustChangeFlag", 1, true).Return(nil)

	err := service.SetMustChangeFlag(1, true)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_SetMustChangeFlag_False(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("SetMustChangeFlag", 1, false).Return(nil)

	err := service.SetMustChangeFlag(1, false)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_SetMustChangeFlag_Error(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("SetMustChangeFlag", 1, true).Return(errors.New("database error"))

	err := service.SetMustChangeFlag(1, true)

	assert.Error(t, err)
}

// ============================================================================
// SetUserRoles tests
// ============================================================================

func TestUserService_SetUserRoles_Success(t *testing.T) {
	service, userRepo := setupUserService()

	roles := []string{"admin", "user"}
	userRepo.On("SetRolesForUser", 1, roles).Return(nil)
	userRepo.On("GetUserById", 1).Return(&types.User{Id: 1, Username: "testuser"}, nil)

	err := service.SetUserRoles(1, roles)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_SetUserRoles_EmptyRoles(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("SetRolesForUser", 1, []string{}).Return(nil)
	userRepo.On("GetUserById", 1).Return(&types.User{Id: 1, Username: "testuser"}, nil)

	err := service.SetUserRoles(1, []string{})

	assert.NoError(t, err)
}

func TestUserService_SetUserRoles_Error(t *testing.T) {
	service, userRepo := setupUserService()

	userRepo.On("SetRolesForUser", 1, []string{"admin"}).Return(errors.New("database error"))

	err := service.SetUserRoles(1, []string{"admin"})

	assert.Error(t, err)
}

// ============================================================================
// Edge cases
// ============================================================================

func TestUserService_CreateUser_NilConfig(t *testing.T) {
	// Test with nil AppConfig (should use default bcrypt cost)
	origConfig := appProps.AppConfig
	appProps.AppConfig = nil
	defer func() { appProps.AppConfig = origConfig }()

	service, userRepo := setupUserService()

	user := types.User{Username: "newuser"}

	userRepo.On("CreateUser", mock.Anything, mock.Anything).Return(1, nil)

	id, err := service.CreateUser(user, "password123")

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
}
