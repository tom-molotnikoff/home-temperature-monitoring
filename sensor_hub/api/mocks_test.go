package api

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

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(user types.User, password string) (int, error) {
	args := m.Called(user, password)
	return args.Int(0), args.Error(1)
}

func (m *MockUserService) ListUsers() ([]types.User, error) {
	args := m.Called()
	return args.Get(0).([]types.User), args.Error(1)
}

func (m *MockUserService) GetUserById(id int) (*types.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(userId int, newPassword string, keepToken string) error {
	args := m.Called(userId, newPassword, keepToken)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

func (m *MockUserService) SetMustChangeFlag(userId int, mustChange bool) error {
	args := m.Called(userId, mustChange)
	return args.Error(0)
}

func (m *MockUserService) SetUserRoles(userId int, roles []string) error {
	args := m.Called(userId, roles)
	return args.Error(0)
}

type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) ListRoles() ([]db.RoleInfo, error) {
	args := m.Called()
	return args.Get(0).([]db.RoleInfo), args.Error(1)
}

func (m *MockRoleService) ListPermissions() ([]db.PermissionInfo, error) {
	args := m.Called()
	return args.Get(0).([]db.PermissionInfo), args.Error(1)
}

func (m *MockRoleService) ListPermissionsForRole(roleId int) ([]db.PermissionInfo, error) {
	args := m.Called(roleId)
	return args.Get(0).([]db.PermissionInfo), args.Error(1)
}

func (m *MockRoleService) AssignPermission(roleId int, permissionId int) error {
	args := m.Called(roleId, permissionId)
	return args.Error(0)
}

func (m *MockRoleService) RemovePermission(roleId int, permissionId int) error {
	args := m.Called(roleId, permissionId)
	return args.Error(0)
}

type MockSensorService struct {
	mock.Mock
}

func (m *MockSensorService) ServiceAddSensor(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorService) ServiceUpdateSensorById(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorService) ServiceDeleteSensorByName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockSensorService) ServiceGetSensorByName(name string) (*types.Sensor, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Sensor), args.Error(1)
}

func (m *MockSensorService) ServiceGetAllSensors() ([]types.Sensor, error) {
	args := m.Called()
	return args.Get(0).([]types.Sensor), args.Error(1)
}

func (m *MockSensorService) ServiceGetSensorsByType(sensorType string) ([]types.Sensor, error) {
	args := m.Called(sensorType)
	return args.Get(0).([]types.Sensor), args.Error(1)
}

func (m *MockSensorService) ServiceGetSensorIdByName(name string) (int, error) {
	args := m.Called(name)
	return args.Int(0), args.Error(1)
}

func (m *MockSensorService) ServiceSensorExists(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

func (m *MockSensorService) ServiceCollectAndStoreAllSensorReadings() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSensorService) ServiceCollectFromSensorByName(sensorName string) error {
	args := m.Called(sensorName)
	return args.Error(0)
}

func (m *MockSensorService) ServiceCollectReadingToValidateSensor(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorService) ServiceCollectAndStoreTemperatureReadings() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSensorService) ServiceStartPeriodicSensorCollection() {
	m.Called()
}

func (m *MockSensorService) ServiceDiscoverSensors() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSensorService) ServiceFetchTemperatureReadingFromSensor(sensor types.Sensor) (types.TemperatureReading, error) {
	args := m.Called(sensor)
	return args.Get(0).(types.TemperatureReading), args.Error(1)
}

func (m *MockSensorService) ServiceFetchAllTemperatureReadings() ([]types.TemperatureReading, error) {
	args := m.Called()
	return args.Get(0).([]types.TemperatureReading), args.Error(1)
}

func (m *MockSensorService) ServiceValidateSensorConfig(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorService) ServiceUpdateSensorHealthById(sensorId int, healthStatus types.SensorHealthStatus, healthReason string) {
	m.Called(sensorId, healthStatus, healthReason)
}

func (m *MockSensorService) ServiceSetEnabledSensorByName(name string, enabled bool) error {
	args := m.Called(name, enabled)
	return args.Error(0)
}

func (m *MockSensorService) ServiceGetSensorHealthHistoryByName(name string, limit int) ([]types.SensorHealthHistory, error) {
	args := m.Called(name, limit)
	return args.Get(0).([]types.SensorHealthHistory), args.Error(1)
}

func (m *MockSensorService) ServiceGetTotalReadingsForEachSensor() (map[string]int, error) {
	args := m.Called()
	return args.Get(0).(map[string]int), args.Error(1)
}

type MockPropertiesService struct {
	mock.Mock
}

func (m *MockPropertiesService) ServiceUpdateProperties(properties map[string]string) error {
	args := m.Called(properties)
	return args.Error(0)
}

func (m *MockPropertiesService) ServiceGetProperties() (map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// ============================================================================
// MockOAuthService
// ============================================================================

type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) GetStatus() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockOAuthService) GetAuthURL(state string) (string, error) {
	args := m.Called(state)
	return args.String(0), args.Error(1)
}

func (m *MockOAuthService) ExchangeCode(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockOAuthService) IsReady() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOAuthService) Reload() error {
	args := m.Called()
	return args.Error(0)
}
