package middleware

import (
	db "example/sensorHub/db"
	"example/sensorHub/types"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(username, password, ip, userAgent string) (string, string, bool, error) {
	args := m.Called(username, password, ip, userAgent)
	return args.String(0), args.String(1), args.Bool(2), args.Error(3)
}

func (m *MockAuthService) ValidateSession(rawToken string) (*types.User, error) {
	args := m.Called(rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockAuthService) Logout(rawToken string) error {
	args := m.Called(rawToken)
	return args.Error(0)
}

func (m *MockAuthService) ChangePassword(userId int, newPassword string) error {
	args := m.Called(userId, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) CreateInitialAdminIfNone(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MockAuthService) ListSessionsForUser(userId int) ([]db.SessionInfo, error) {
	args := m.Called(userId)
	return args.Get(0).([]db.SessionInfo), args.Error(1)
}

func (m *MockAuthService) RevokeSessionById(sessionId int64) error {
	args := m.Called(sessionId)
	return args.Error(0)
}

func (m *MockAuthService) RevokeSessionByIdWithActor(sessionId int64, revokedByUserId *int, reason *string) error {
	args := m.Called(sessionId, revokedByUserId, reason)
	return args.Error(0)
}

func (m *MockAuthService) GetCSRFForToken(rawToken string) (string, error) {
	args := m.Called(rawToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetSessionIdForToken(rawToken string) (int64, error) {
	args := m.Called(rawToken)
	return args.Get(0).(int64), args.Error(1)
}

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetPermissionsForUser(userId int) ([]string, error) {
	args := m.Called(userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRoleRepository) GetAllRoles() ([]db.RoleInfo, error) {
	args := m.Called()
	return args.Get(0).([]db.RoleInfo), args.Error(1)
}

func (m *MockRoleRepository) GetAllPermissions() ([]db.PermissionInfo, error) {
	args := m.Called()
	return args.Get(0).([]db.PermissionInfo), args.Error(1)
}

func (m *MockRoleRepository) GetPermissionsForRole(roleId int) ([]db.PermissionInfo, error) {
	args := m.Called(roleId)
	return args.Get(0).([]db.PermissionInfo), args.Error(1)
}

func (m *MockRoleRepository) AssignPermissionToRole(roleId int, permissionId int) error {
	args := m.Called(roleId, permissionId)
	return args.Error(0)
}

func (m *MockRoleRepository) RemovePermissionFromRole(roleId int, permissionId int) error {
	args := m.Called(roleId, permissionId)
	return args.Error(0)
}
