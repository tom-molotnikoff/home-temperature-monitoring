package service

import (
	"time"

	database "example/sensorHub/db"
	"example/sensorHub/types"

	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MockUserRepository
// ============================================================================

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByUsername(username string) (*types.User, string, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*types.User), args.String(1), args.Error(2)
}

func (m *MockUserRepository) GetUserById(id int) (*types.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(user types.User, passwordHash string) (int, error) {
	args := m.Called(user, passwordHash)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) ListUsers() ([]types.User, error) {
	args := m.Called()
	return args.Get(0).([]types.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(userId int, passwordHash string, mustChange bool) error {
	args := m.Called(userId, passwordHash, mustChange)
	return args.Error(0)
}

func (m *MockUserRepository) SetDisabled(userId int, disabled bool) error {
	args := m.Called(userId, disabled)
	return args.Error(0)
}

func (m *MockUserRepository) AssignRoleToUser(userId int, roleName string) error {
	args := m.Called(userId, roleName)
	return args.Error(0)
}

func (m *MockUserRepository) GetRolesForUser(userId int) ([]string, error) {
	args := m.Called(userId)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserRepository) DeleteSessionsForUser(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteSessionsForUserExcept(userId int, keepToken string) error {
	args := m.Called(userId, keepToken)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUserById(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

func (m *MockUserRepository) SetMustChangeFlag(userId int, mustChange bool) error {
	args := m.Called(userId, mustChange)
	return args.Error(0)
}

func (m *MockUserRepository) SetRolesForUser(userId int, roles []string) error {
	args := m.Called(userId, roles)
	return args.Error(0)
}

// ============================================================================
// MockSessionRepository
// ============================================================================

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(userId int, rawToken string, expiresAt time.Time, ip string, userAgent string) (string, error) {
	args := m.Called(userId, rawToken, expiresAt, ip, userAgent)
	return args.String(0), args.Error(1)
}

func (m *MockSessionRepository) GetUserIdByToken(rawToken string) (int, error) {
	args := m.Called(rawToken)
	return args.Int(0), args.Error(1)
}

func (m *MockSessionRepository) DeleteSessionByToken(rawToken string) error {
	args := m.Called(rawToken)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpiredSessions() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSessionRepository) ListSessionsForUser(userId int) ([]database.SessionInfo, error) {
	args := m.Called(userId)
	return args.Get(0).([]database.SessionInfo), args.Error(1)
}

func (m *MockSessionRepository) RevokeSessionById(sessionId int64) error {
	args := m.Called(sessionId)
	return args.Error(0)
}

func (m *MockSessionRepository) InsertSessionAudit(sessionId int64, revokedByUserId *int, action string, reason *string) error {
	args := m.Called(sessionId, revokedByUserId, action, reason)
	return args.Error(0)
}

func (m *MockSessionRepository) GetCSRFForToken(rawToken string) (string, error) {
	args := m.Called(rawToken)
	return args.String(0), args.Error(1)
}

func (m *MockSessionRepository) GetSessionIdByToken(rawToken string) (int64, error) {
	args := m.Called(rawToken)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepository) DeleteSessionsForUser(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

// ============================================================================
// MockFailedLoginRepository
// ============================================================================

type MockFailedLoginRepository struct {
	mock.Mock
}

func (m *MockFailedLoginRepository) RecordFailedAttempt(username string, userId *int, ip string, reason string) error {
	args := m.Called(username, userId, ip, reason)
	return args.Error(0)
}

func (m *MockFailedLoginRepository) CountRecentFailedAttemptsByUsername(username string, window time.Duration) (int, error) {
	args := m.Called(username, window)
	return args.Int(0), args.Error(1)
}

func (m *MockFailedLoginRepository) CountRecentFailedAttemptsByIP(ip string, window time.Duration) (int, error) {
	args := m.Called(ip, window)
	return args.Int(0), args.Error(1)
}

func (m *MockFailedLoginRepository) DeleteRecentFailedAttemptsByIP(ip string, window time.Duration) error {
	args := m.Called(ip, window)
	return args.Error(0)
}

func (m *MockFailedLoginRepository) DeleteAttemptsOlderThan(threshold time.Time) error {
	args := m.Called(threshold)
	return args.Error(0)
}

// ============================================================================
// MockRoleRepository
// ============================================================================

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetPermissionsForUser(userId int) ([]string, error) {
	args := m.Called(userId)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRoleRepository) GetAllRoles() ([]database.RoleInfo, error) {
	args := m.Called()
	return args.Get(0).([]database.RoleInfo), args.Error(1)
}

func (m *MockRoleRepository) GetAllPermissions() ([]database.PermissionInfo, error) {
	args := m.Called()
	return args.Get(0).([]database.PermissionInfo), args.Error(1)
}

func (m *MockRoleRepository) GetPermissionsForRole(roleId int) ([]database.PermissionInfo, error) {
	args := m.Called(roleId)
	return args.Get(0).([]database.PermissionInfo), args.Error(1)
}

func (m *MockRoleRepository) AssignPermissionToRole(roleId int, permissionId int) error {
	args := m.Called(roleId, permissionId)
	return args.Error(0)
}

func (m *MockRoleRepository) RemovePermissionFromRole(roleId int, permissionId int) error {
	args := m.Called(roleId, permissionId)
	return args.Error(0)
}

// ============================================================================
// MockSensorRepository
// ============================================================================

type MockSensorRepository struct {
	mock.Mock
}

func (m *MockSensorRepository) AddSensor(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorRepository) GetAllSensors() ([]types.Sensor, error) {
	args := m.Called()
	return args.Get(0).([]types.Sensor), args.Error(1)
}

func (m *MockSensorRepository) GetSensorByName(name string) (*types.Sensor, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Sensor), args.Error(1)
}

func (m *MockSensorRepository) GetSensorsByType(sensorType string) ([]types.Sensor, error) {
	args := m.Called(sensorType)
	return args.Get(0).([]types.Sensor), args.Error(1)
}

func (m *MockSensorRepository) GetSensorIdByName(name string) (int, error) {
	args := m.Called(name)
	return args.Int(0), args.Error(1)
}

func (m *MockSensorRepository) SensorExists(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

func (m *MockSensorRepository) UpdateSensorById(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorRepository) DeleteSensorByName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockSensorRepository) UpdateSensorHealthById(sensorId int, health types.SensorHealthStatus, reason string) error {
	args := m.Called(sensorId, health, reason)
	return args.Error(0)
}

func (m *MockSensorRepository) GetSensorHealthHistoryById(sensorId int, limit int) ([]types.SensorHealthHistory, error) {
	args := m.Called(sensorId, limit)
	return args.Get(0).([]types.SensorHealthHistory), args.Error(1)
}

func (m *MockSensorRepository) SetEnabledSensorByName(name string, enabled bool) error {
	args := m.Called(name, enabled)
	return args.Error(0)
}

func (m *MockSensorRepository) DeleteHealthHistoryOlderThan(threshold time.Time) error {
	args := m.Called(threshold)
	return args.Error(0)
}

// ============================================================================
// MockTemperatureRepository
// ============================================================================

type MockTemperatureRepository struct {
	mock.Mock
}

func (m *MockTemperatureRepository) Add(readings []types.TemperatureReading) error {
	args := m.Called(readings)
	return args.Error(0)
}

func (m *MockTemperatureRepository) GetBetweenDates(tableName string, startDate string, endDate string) ([]types.TemperatureReading, error) {
	args := m.Called(tableName, startDate, endDate)
	return args.Get(0).([]types.TemperatureReading), args.Error(1)
}

func (m *MockTemperatureRepository) GetLatest() ([]types.TemperatureReading, error) {
	args := m.Called()
	return args.Get(0).([]types.TemperatureReading), args.Error(1)
}

func (m *MockTemperatureRepository) GetTotalReadingsBySensorId(sensorId int) (int, error) {
	args := m.Called(sensorId)
	return args.Int(0), args.Error(1)
}

func (m *MockTemperatureRepository) DeleteReadingsOlderThan(threshold time.Time) error {
	args := m.Called(threshold)
	return args.Error(0)
}
