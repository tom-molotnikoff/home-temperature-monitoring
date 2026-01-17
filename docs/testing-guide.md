# Testing Guide

This document describes the test implementation for each package in the Sensor Hub application. It covers mocking infrastructure, test patterns, and how to write new tests for each layer of the architecture.

## Table of Contents

1. [Overview](#overview)
2. [Running Tests](#running-tests)
3. [Test Infrastructure](#test-infrastructure)
4. [Package-Specific Testing](#package-specific-testing)
   - [Database Package (db/)](#database-package-db)
   - [Service Package (service/)](#service-package-service)
   - [API Package (api/)](#api-package-api)
   - [OAuth Package (oauth/)](#oauth-package-oauth)
   - [Application Properties Package](#application-properties-package)
   - [Other Packages](#other-packages)
5. [Writing New Tests](#writing-new-tests)

---

## Overview

The Sensor Hub uses Go's standard testing framework with the following libraries:

- **github.com/stretchr/testify/assert** - Assertion helpers
- **github.com/stretchr/testify/mock** - Mock object framework
- **github.com/stretchr/testify/require** - Fatal assertion helpers
- **github.com/DATA-DOG/go-sqlmock** - SQL database mocking
- **github.com/gin-gonic/gin** - HTTP testing with test mode

All test files follow the Go convention of `*_test.go` naming and are excluded from production builds.

---

## Running Tests

```bash
# Run all tests
cd sensor_hub
go test ./...

# Run tests for a specific package
go test ./service

# Run tests with verbose output
go test -v ./service

# Run a specific test
go test -v ./service -run TestAuthService_Login_Success

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Test Infrastructure

### Mock Files Location

| Package | Mock Location | Purpose |
|---------|--------------|---------|
| `service/` | `service/test_mocks_test.go` | Repository interface mocks for service tests |
| `db/` | `db/test_helpers_test.go` | SQL mock helpers and test data factories |
| `api/` | `api/mocks_test.go` | Service interface mocks for API handler tests |
| `api/middleware/` | `api/middleware/mocks_test.go` | Service mocks for middleware tests |
| `oauth/` | `oauth/test_helpers_test.go` | File system mocks and OAuth test data |

### Testing Library: testify/mock

All mocks use the `testify/mock` library pattern:

```go
type MockService struct {
    mock.Mock
}

func (m *MockService) DoSomething(arg string) (string, error) {
    args := m.Called(arg)
    return args.String(0), args.Error(1)
}
```

Setting up expectations:

```go
mockService := new(MockService)
mockService.On("DoSomething", "input").Return("output", nil)

// Call the function under test
result, err := functionUnderTest(mockService, "input")

// Verify expectations
mockService.AssertExpectations(t)
```

---

## Package-Specific Testing

### Database Package (db/)

**Location:** `sensor_hub/db/`

**Test Files:**
- `test_helpers_test.go` - Mock DB helpers and test data factories
- `*_repository_test.go` - Individual repository tests

#### Mock Database Setup

The database package uses `go-sqlmock` to mock SQL queries:

```go
import (
    "github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })
    return db, mock
}
```

#### Test Data Factories

Pre-defined test data factories in `test_helpers_test.go`:

```go
func testSensor() types.Sensor { ... }
func testSensorWithID(id int, name string) types.Sensor { ... }
func testUser() types.User { ... }
func testUserWithID(id int, username string) types.User { ... }
func testAlertRule() alerting.AlertRule { ... }
func testTemperatureReading() types.TemperatureReading { ... }
func testSessionInfo() SessionInfo { ... }
func testSensorHealthHistory() types.SensorHealthHistory { ... }
```

#### Column Definitions

Pre-defined column slices for sqlmock rows:

```go
var sensorColumns = []string{"id", "name", "type", "url", "health_status", "health_reason", "enabled"}
var userColumns = []string{"id", "username", "email", "must_change_password", "disabled", "created_at", "updated_at"}
var sessionColumns = []string{"id", "user_id", "created_at", "expires_at", "last_accessed_at", "ip_address", "user_agent"}
// ... etc
```

#### Example Repository Test

```go
func TestSensorRepository_GetAllSensors(t *testing.T) {
    db, mock := newMockDB(t)
    repo := NewSensorRepository(db)

    sensor := testSensor()
    rows := sqlmock.NewRows(sensorColumns).
        AddRow(sensor.Id, sensor.Name, sensor.Type, sensor.URL, 
               sensor.HealthStatus, sensor.HealthReason, sensor.Enabled)

    mock.ExpectQuery("SELECT .* FROM sensors").WillReturnRows(rows)

    sensors, err := repo.GetAllSensors()

    assert.NoError(t, err)
    assert.Len(t, sensors, 1)
    assert.Equal(t, "test-sensor", sensors[0].Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

---

### Service Package (service/)

**Location:** `sensor_hub/service/`

**Test Files:**
- `test_mocks_test.go` - Repository interface mocks
- `*_service_test.go` - Individual service tests

#### Available Mock Repositories

All mock repositories are defined in `test_mocks_test.go`:

- `MockUserRepository` - User CRUD operations
- `MockSessionRepository` - Session management
- `MockFailedLoginRepository` - Login attempt tracking
- `MockRoleRepository` - RBAC operations
- `MockSensorRepository` - Sensor CRUD operations
- `MockTemperatureRepository` - Temperature readings

#### Service Test Setup Pattern

Each service test typically follows this pattern:

```go
func setupAuthService() (*AuthService, *MockUserRepository, *MockSessionRepository, *MockFailedLoginRepository, *MockRoleRepository) {
    userRepo := new(MockUserRepository)
    sessionRepo := new(MockSessionRepository)
    failedRepo := new(MockFailedLoginRepository)
    roleRepo := new(MockRoleRepository)

    service := NewAuthService(userRepo, sessionRepo, failedRepo, roleRepo)
    return service, userRepo, sessionRepo, failedRepo, roleRepo
}
```

#### Configuration Setup for Tests

Many services use global configuration. Use a setup/teardown pattern:

```go
func setupTestConfig() func() {
    origConfig := appProps.AppConfig
    appProps.AppConfig = &appProps.ApplicationConfiguration{
        AuthBcryptCost:                4, // Low cost for fast tests
        AuthSessionTTLMinutes:         60,
        AuthLoginBackoffWindowMinutes: 15,
        AuthLoginBackoffThreshold:     5,
        AuthLoginBackoffBaseSeconds:   2,
        AuthLoginBackoffMaxSeconds:    300,
    }
    return func() { appProps.AppConfig = origConfig }
}

func TestSomething(t *testing.T) {
    defer setupTestConfig()()
    // ... test code
}
```

#### Example Service Test

```go
func TestAuthService_Login_Success(t *testing.T) {
    defer setupTestConfig()()
    resetBlockers()

    service, userRepo, sessionRepo, failedRepo, _ := setupAuthService()

    // bcrypt hash of "password123" with cost 4
    passwordHash := "$2a$04$8/TZfgezGK2PM2Eoni4P6O/nUDjGtd4rLPMHqQ7g4n3DATqIDPRxq"
    user := &types.User{Id: 1, Username: "testuser", Disabled: false}

    failedRepo.On("CountRecentFailedAttemptsByUsername", "testuser", mock.Anything).Return(0, nil)
    failedRepo.On("CountRecentFailedAttemptsByIP", "192.168.1.1", mock.Anything).Return(0, nil)
    userRepo.On("GetUserByUsername", "testuser").Return(user, passwordHash, nil)
    sessionRepo.On("CreateSession", 1, mock.Anything, mock.Anything, "192.168.1.1", "TestAgent").Return("csrf-token", nil)
    failedRepo.On("DeleteRecentFailedAttemptsByIP", "192.168.1.1", mock.Anything).Return(nil)

    token, csrf, mustChange, err := service.Login("testuser", "password123", "192.168.1.1", "TestAgent")

    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    assert.Equal(t, "csrf-token", csrf)
    assert.False(t, mustChange)
    userRepo.AssertExpectations(t)
    sessionRepo.AssertExpectations(t)
}
```

---

### API Package (api/)

**Location:** `sensor_hub/api/`

**Test Files:**
- `mocks_test.go` - Service interface mocks
- `*_api_test.go` - Individual API handler tests

#### Available Service Mocks

All mocks are defined in `mocks_test.go`:

- `MockAuthService` - Authentication operations
- `MockUserService` - User management
- `MockRoleService` - Role management
- `MockSensorService` - Sensor operations
- `MockPropertiesService` - Application properties
- `MockOAuthService` - OAuth operations

#### HTTP Testing Pattern

API tests use Gin's test mode:

```go
import (
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
)

func TestHandler_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    mockService := new(MockService)
    mockService.On("GetData").Return([]Data{...}, nil)

    router := gin.New()
    router.GET("/api/data", createHandler(mockService))

    req := httptest.NewRequest("GET", "/api/data", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    // Parse and verify response body
    mockService.AssertExpectations(t)
}
```

#### Testing with Authentication Context

Many handlers require authenticated user context:

```go
func TestHandler_WithAuth(t *testing.T) {
    gin.SetMode(gin.TestMode)

    mockService := new(MockService)
    mockService.On("DoAction", 1).Return(nil) // userId = 1

    router := gin.New()
    router.POST("/api/action", func(c *gin.Context) {
        // Simulate authenticated user
        c.Set("user", &types.User{Id: 1, Username: "testuser"})
        c.Next()
    }, actionHandler)

    req := httptest.NewRequest("POST", "/api/action", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

#### Example API Test

```go
func TestOAuthStatusHandler_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)

    mockOAuth := new(MockOAuthService)
    InitOAuthAPI(mockOAuth)

    status := map[string]interface{}{
        "configured":   true,
        "token_valid": true,
    }
    mockOAuth.On("GetStatus").Return(status)

    router := gin.New()
    router.GET("/oauth/status", oauthStatusHandler)

    req := httptest.NewRequest("GET", "/oauth/status", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.True(t, response["configured"].(bool))
    
    mockOAuth.AssertExpectations(t)
}
```

---

### OAuth Package (oauth/)

**Location:** `sensor_hub/oauth/`

**Test Files:**
- `test_helpers_test.go` - File system mocks and test data factories
- `oauth_test.go` - OAuth service and XOauth2Auth tests

#### File System Mocking

OAuth tests mock file system operations:

```go
// MockFileReader implements FileReader for testing
type MockFileReader struct {
    Files map[string][]byte
    Err   error
}

func (m *MockFileReader) ReadFile(path string) ([]byte, error) {
    if m.Err != nil {
        return nil, m.Err
    }
    data, ok := m.Files[path]
    if !ok {
        return nil, errors.New("file not found: " + path)
    }
    return data, nil
}

// MockFileWriter implements FileWriter for testing
type MockFileWriter struct {
    Written map[string][]byte
    Err     error
}

func NewMockFileWriter() *MockFileWriter {
    return &MockFileWriter{Written: make(map[string][]byte)}
}

func (m *MockFileWriter) WriteFile(path string, data []byte, perm uint32) error {
    if m.Err != nil {
        return m.Err
    }
    m.Written[path] = data
    return nil
}
```

#### OAuth Test Data Factories

```go
func testCredentialsJSON() []byte {
    creds := map[string]interface{}{
        "installed": map[string]interface{}{
            "client_id":     "test-client-id.apps.googleusercontent.com",
            "client_secret": "test-client-secret",
            "auth_uri":      "https://accounts.google.com/o/oauth2/auth",
            "token_uri":     "https://oauth2.googleapis.com/token",
            "redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob", "http://localhost"},
        },
    }
    data, _ := json.Marshal(creds)
    return data
}

func testTokenJSON() []byte {
    token := oauth2.Token{
        AccessToken:  "test-access-token",
        TokenType:    "Bearer",
        RefreshToken: "test-refresh-token",
    }
    data, _ := json.Marshal(token)
    return data
}

func testExpiredTokenJSON() []byte {
    token := map[string]interface{}{
        "access_token":  "expired-access-token",
        "token_type":    "Bearer",
        "refresh_token": "test-refresh-token",
        "expiry":        "2020-01-01T00:00:00Z",
    }
    data, _ := json.Marshal(token)
    return data
}
```

#### Example OAuth Test

```go
func TestOAuthService_Initialise_Success(t *testing.T) {
    reader := &MockFileReader{
        Files: map[string][]byte{
            "credentials.json": testCredentialsJSON(),
            "token.json":       testTokenJSON(),
        },
    }
    writer := NewMockFileWriter()

    service := NewOAuthService("credentials.json", "token.json", 60, reader, writer)
    err := service.Initialise()

    assert.NoError(t, err)
    assert.True(t, service.IsReady())
    
    status := service.GetStatus()
    assert.True(t, status["configured"].(bool))
    assert.True(t, status["token_valid"].(bool))
}

func TestOAuthService_Initialise_NoCredentials(t *testing.T) {
    reader := &MockFileReader{
        Err: errors.New("file not found"),
    }
    writer := NewMockFileWriter()

    service := NewOAuthService("credentials.json", "token.json", 60, reader, writer)
    err := service.Initialise()

    assert.Error(t, err)
    assert.False(t, service.IsReady())
}
```

---

### Application Properties Package

**Location:** `sensor_hub/application_properties/`

**Test File:** `application_properties_test.go`

#### Testing Configuration Loading

```go
func TestLoadConfigurationFromMaps_OAuth(t *testing.T) {
    configMap := map[string]string{
        "oauth.credentials.file.path":          "/custom/credentials.json",
        "oauth.token.file.path":                "/custom/token.json",
        "oauth.token.refresh.interval.minutes": "30",
    }

    config := &ApplicationConfiguration{}
    
    LoadConfigurationFromMaps(configMap, nil, config)

    assert.Equal(t, "/custom/credentials.json", config.OAuthCredentialsFilePath)
    assert.Equal(t, "/custom/token.json", config.OAuthTokenFilePath)
    assert.Equal(t, 30, config.OAuthTokenRefreshIntervalMinutes)
}
```

---

### Other Packages

#### Alerting (alerting/)

Uses standard table-driven tests for alert rule evaluation:

```go
func TestAlertRule_Evaluate(t *testing.T) {
    tests := []struct {
        name     string
        rule     AlertRule
        reading  float64
        expected bool
    }{
        {"within range", AlertRule{HighThreshold: 30, LowThreshold: 10}, 20.0, false},
        {"above threshold", AlertRule{HighThreshold: 30, LowThreshold: 10}, 35.0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.rule.ShouldAlert(tt.reading)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### SMTP (smtp/)

Tests the SMTP client without actual network calls using interface mocking.

#### Utils (utils/)

Simple unit tests for utility functions.

#### WebSocket (ws/)

Tests for hub and client connection management.

---

## Writing New Tests

### Step 1: Identify the Layer

| Layer | What to Mock | Test File Location |
|-------|-------------|-------------------|
| Repository | SQL queries | `db/*_repository_test.go` |
| Service | Repository interfaces | `service/*_service_test.go` |
| API | Service interfaces | `api/*_api_test.go` |

### Step 2: Check for Existing Mocks

Look in the appropriate `*_mocks_test.go` or `test_helpers_test.go` file.

### Step 3: Create Mock if Needed

Follow the testify/mock pattern:

```go
type MockNewInterface struct {
    mock.Mock
}

func (m *MockNewInterface) MethodName(arg Type) (ResultType, error) {
    args := m.Called(arg)
    return args.Get(0).(ResultType), args.Error(1)
}
```

### Step 4: Write Test

```go
func TestNewFeature_Scenario(t *testing.T) {
    // 1. Setup - create mocks
    mock := new(MockService)
    
    // 2. Set expectations
    mock.On("Method", expectedArg).Return(expectedResult, nil)
    
    // 3. Execute - call the function under test
    result, err := functionUnderTest(mock, input)
    
    // 4. Assert - verify results
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    
    // 5. Verify mock expectations
    mock.AssertExpectations(t)
}
```

### Step 5: Run and Verify

```bash
go test -v ./package -run TestNewFeature
```

---

## Best Practices

1. **Use table-driven tests** for functions with multiple input scenarios
2. **Keep tests independent** - each test should set up its own state
3. **Use t.Cleanup()** for automatic cleanup after tests
4. **Use low bcrypt costs** (4) in tests for speed
5. **Always call AssertExpectations(t)** on mocks
6. **Use descriptive test names** following `TestType_Method_Scenario` convention
7. **Mock at boundaries** - repositories for services, services for APIs
8. **Use test data factories** for consistent test data
