# Service Package Comprehensive Tests Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Implement comprehensive unit tests for all 9 services in the service package, covering business logic, error handling, and edge cases.

**Architecture:** 
- One test file per service (matching source structure)
- Mock all repositories using testify/mock
- Mock HTTP calls using httptest
- Mock global config by setting appProps.AppConfig
- Mock WebSocket broadcasts (ws package calls are fire-and-forget, can be ignored or mocked)

**Tech Stack:** Go 1.25, testify/assert, testify/mock, httptest

---

## Services to Test

| Service | Test File | Methods | Estimated Tests |
|---------|-----------|---------|-----------------|
| AuthService | auth_service_test.go | 11 | ~35 |
| SensorService | sensor_service_test.go | 20 | ~45 |
| UserService | user_service_test.go | 7 | ~20 |
| CleanupService | cleanup_service_test.go | 2 | ~10 |
| RoleService | role_service_test.go | 5 | ~15 |
| TemperatureService | temperature_service_test.go | 2 | ~8 |
| PropertiesService | properties_service_test.go | 2 | ~10 |
| LoginLimiter | login_limiter_test.go | 5 | ~12 |
| AlertService | (existing) | 6 | 6 (done) |

**Total: ~155 new tests**

---

### Task 1: Create shared test mocks file

**Files:**
- Create: `sensor_hub/service/test_mocks_test.go`

**Step 1: Create shared mock definitions**

```go
package service

import (
	"example/sensorHub/db"
	"example/sensorHub/types"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository mocks database.UserRepository
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

// MockSessionRepository mocks database.SessionRepository
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

// MockFailedLoginRepository mocks database.FailedLoginRepository
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

// MockRoleRepository mocks database.RoleRepository
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

// MockSensorRepository mocks database.SensorRepositoryInterface
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

// MockTemperatureRepository mocks database.ReadingsRepository
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
```

**Step 2: Verify it compiles**

Run: `cd sensor_hub && go build ./service/...`
Expected: No errors

---

### Task 2: AuthService tests

**Files:**
- Create: `sensor_hub/service/auth_service_test.go`

**Methods to test:**
- `Login` - success, invalid credentials, user not found, disabled account, rate limiting, bcrypt error
- `ValidateSession` - success, invalid token, user not found, with permissions
- `Logout` - success, error
- `ChangePassword` - success, bcrypt error, repo error
- `CreateInitialAdminIfNone` - creates admin, skips if users exist, error cases
- `ListSessionsForUser` - success, error
- `RevokeSessionById` / `RevokeSessionByIdWithActor` - success, error
- `GetCSRFForToken` - success, error
- `GetSessionIdForToken` - success, error

**Estimated: ~35 tests**

---

### Task 3: UserService tests

**Files:**
- Create: `sensor_hub/service/user_service_test.go`

**Methods to test:**
- `CreateUser` - success, empty password, bcrypt error, with roles
- `ListUsers` - success, error
- `GetUserById` - success, not found
- `ChangePassword` - success, empty password, with keepToken, without keepToken
- `DeleteUser` - success, error
- `SetMustChangeFlag` - true, false, error
- `SetUserRoles` - success, error

**Estimated: ~20 tests**

---

### Task 4: SensorService tests (part 1 - CRUD operations)

**Files:**
- Create: `sensor_hub/service/sensor_service_test.go`

**Methods to test first:**
- `ServiceAddSensor` - success, already exists, validation error
- `ServiceUpdateSensorById` - success, validation error
- `ServiceDeleteSensorByName` - success, not exists, error
- `ServiceGetSensorByName` - success, empty name, not found
- `ServiceGetAllSensors` - success, error
- `ServiceGetSensorsByType` - success, error
- `ServiceGetSensorIdByName` - success, error
- `ServiceSensorExists` - true, false, error
- `ServiceSetEnabledSensorByName` - enable, disable, not exists
- `ServiceGetTotalReadingsForEachSensor` - success, error
- `ServiceGetSensorHealthHistoryByName` - success, error

**Estimated: ~25 tests**

---

### Task 5: SensorService tests (part 2 - reading collection with httptest)

**Methods to test:**
- `ServiceFetchTemperatureReadingFromSensor` - success, HTTP error, non-200, JSON decode error
- `ServiceCollectFromSensorByName` - success, sensor not found, disabled, unsupported type
- `ServiceFetchAllTemperatureReadings` - success with multiple sensors, skip disabled
- `ServiceCollectAndStoreTemperatureReadings` - success, storage error
- `ServiceCollectReadingToValidateSensor` - success, unsupported type, fetch error
- `ServiceValidateSensorConfig` - valid, empty fields, fetch fails

**Estimated: ~20 tests**

---

### Task 6: CleanupService tests

**Files:**
- Create: `sensor_hub/service/cleanup_service_test.go`

**Methods to test:**
- `performCleanup` - all cleanups success, skip zero retention, temperature error, health error, failed login error

**Estimated: ~10 tests**

---

### Task 7: RoleService tests

**Files:**
- Create: `sensor_hub/service/role_service_test.go`

**Methods to test:**
- `ListRoles` - success, error
- `ListPermissions` - success, error
- `ListPermissionsForRole` - success, error
- `AssignPermission` - success, error
- `RemovePermission` - success, error

**Estimated: ~15 tests**

---

### Task 8: TemperatureService tests

**Files:**
- Create: `sensor_hub/service/temperature_service_test.go`

**Methods to test:**
- `ServiceGetBetweenDates` - success, error
- `ServiceGetLatest` - success, empty, error

**Estimated: ~8 tests**

---

### Task 9: PropertiesService tests

**Files:**
- Create: `sensor_hub/service/properties_service_test.go`

**Methods to test:**
- `ServiceUpdateProperties` - success, validation error, skip sensitive unchanged
- `ServiceGetProperties` - success, masks sensitive values

**Estimated: ~10 tests**

---

### Task 10: LoginLimiter tests

**Files:**
- Create: `sensor_hub/service/login_limiter_test.go`

**Methods to test:**
- `getRemainingSeconds` - blocked, not blocked, expired
- `blockFor` - success, zero seconds
- `consumeAllowOnceIfReady` - success, not ready, exhausted
- `forceClearAllowOnce` - clears state

**Estimated: ~12 tests**

---

### Task 11: Final verification

**Step 1: Run all service tests**

Run: `cd sensor_hub && go test ./service/... -v`
Expected: All tests PASS

**Step 2: Count tests**

Run: `cd sensor_hub && go test ./service/... -v 2>&1 | grep -c "^--- PASS"`
Expected: ~155+ tests

**Step 3: Run full project tests**

Run: `cd sensor_hub && go test $(go list ./... | grep -v integration)`
Expected: All packages PASS

---

## Summary

| Task | File | Tests |
|------|------|-------|
| 1 | test_mocks_test.go | 0 (setup) |
| 2 | auth_service_test.go | ~35 |
| 3 | user_service_test.go | ~20 |
| 4 | sensor_service_test.go (CRUD) | ~25 |
| 5 | sensor_service_test.go (HTTP) | ~20 |
| 6 | cleanup_service_test.go | ~10 |
| 7 | role_service_test.go | ~15 |
| 8 | temperature_service_test.go | ~8 |
| 9 | properties_service_test.go | ~10 |
| 10 | login_limiter_test.go | ~12 |
| 11 | Final verification | 0 |

**Total: ~155 new tests**
