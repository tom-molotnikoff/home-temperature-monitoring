package middleware

import (
	"context"
	db "example/sensorHub/db"
	"example/sensorHub/types"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password, ip, userAgent string) (string, string, bool, error) {
	args := m.Called(ctx, username, password, ip, userAgent)
	return args.String(0), args.String(1), args.Bool(2), args.Error(3)
}

func (m *MockAuthService) ValidateSession(ctx context.Context, rawToken string) (*types.User, error) {
	args := m.Called(ctx, rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, rawToken string) error {
	args := m.Called(ctx, rawToken)
	return args.Error(0)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userId int, newPassword string) error {
	args := m.Called(ctx, userId, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) CreateInitialAdminIfNone(ctx context.Context, username, password string) error {
	args := m.Called(ctx, username, password)
	return args.Error(0)
}

func (m *MockAuthService) ListSessionsForUser(ctx context.Context, userId int) ([]db.SessionInfo, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]db.SessionInfo), args.Error(1)
}

func (m *MockAuthService) RevokeSessionById(ctx context.Context, sessionId int64) error {
	args := m.Called(ctx, sessionId)
	return args.Error(0)
}

func (m *MockAuthService) RevokeSessionByIdWithActor(ctx context.Context, sessionId int64, revokedByUserId *int, reason *string) error {
	args := m.Called(ctx, sessionId, revokedByUserId, reason)
	return args.Error(0)
}

func (m *MockAuthService) GetCSRFForToken(ctx context.Context, rawToken string) (string, error) {
	args := m.Called(ctx, rawToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetSessionIdForToken(ctx context.Context, rawToken string) (int64, error) {
	args := m.Called(ctx, rawToken)
	return args.Get(0).(int64), args.Error(1)
}

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetPermissionsForUser(ctx context.Context, userId int) ([]string, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRoleRepository) GetAllRoles(ctx context.Context) ([]db.RoleInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.RoleInfo), args.Error(1)
}

func (m *MockRoleRepository) GetAllPermissions(ctx context.Context) ([]db.PermissionInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.PermissionInfo), args.Error(1)
}

func (m *MockRoleRepository) GetPermissionsForRole(ctx context.Context, roleId int) ([]db.PermissionInfo, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).([]db.PermissionInfo), args.Error(1)
}

func (m *MockRoleRepository) AssignPermissionToRole(ctx context.Context, roleId int, permissionId int) error {
	args := m.Called(ctx, roleId, permissionId)
	return args.Error(0)
}

func (m *MockRoleRepository) RemovePermissionFromRole(ctx context.Context, roleId int, permissionId int) error {
	args := m.Called(ctx, roleId, permissionId)
	return args.Error(0)
}
