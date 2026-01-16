# API Test Coverage Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Increase test coverage for the API package, focusing on critical middleware and authentication components first.

**Architecture:** We will use `testify/mock` for mocking dependencies (AuthService, RoleRepository) and `httptest` for integration-style testing of Gin handlers. Tests will follow the Arrange-Act-Assert pattern.

**Tech Stack:** Go, Gin, Testify

---

### Task 1: Create Shared Mocks for Middleware Tests

**Files:**
- Create: `sensor_hub/api/middleware/mocks_test.go`

**Step 1: Create mock implementations**

Create the shared mock file that will be used by all middleware tests. This includes mocks for `AuthService` and `RoleRepository`.

```go
package middleware

import (
	"example/sensorHub/db"
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
```

**Step 2: Verify file creation**

Run: `ls -l sensor_hub/api/middleware/mocks_test.go`

---

### Task 2: Implement Auth Middleware Tests

**Files:**
- Create: `sensor_hub/api/middleware/auth_middleware_test.go`

**Step 1: Write tests for AuthRequired and RequireAdmin**

Tests should cover:
- Valid session
- Missing session cookie
- Invalid session token
- User with MustChangePassword (access control)
- Admin access
- Non-admin access to admin routes

```go
package middleware

import (
	"example/sensorHub/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthRequired_ValidSession(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService)

	user := &types.User{Id: 1, Username: "testuser"}
	mockService.On("ValidateSession", "valid-token").Return(user, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/protected", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})

	AuthRequired()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	u, exists := c.Get("currentUser")
	assert.True(t, exists)
	assert.Equal(t, user, u)
}

func TestAuthRequired_NoCookie(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/protected", nil)

	AuthRequired()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_InvalidSession(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService)

	mockService.On("ValidateSession", "invalid-token").Return(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/protected", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "invalid-token"})

	AuthRequired()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_MustChangePassword_Allowed(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService)

	user := &types.User{Id: 1, Username: "testuser", MustChangePassword: true}
	mockService.On("ValidateSession", "valid-token").Return(user, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/users/password", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})

	AuthRequired()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRequired_MustChangePassword_Forbidden(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService)

	user := &types.User{Id: 1, Username: "testuser", MustChangePassword: true}
	mockService.On("ValidateSession", "valid-token").Return(user, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/other", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})

	AuthRequired()(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_AdminUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := &types.User{Id: 1, Roles: []string{"admin"}}
	c.Set("currentUser", user)

	RequireAdmin()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdmin_NonAdminUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := &types.User{Id: 1, Roles: []string{"user"}}
	c.Set("currentUser", user)

	RequireAdmin()(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_NoUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RequireAdmin()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

**Step 2: Run tests**

Run: `go test ./sensor_hub/api/middleware/... -v`
Expected: PASS

---

### Task 3: Implement CSRF Middleware Tests

**Files:**
- Create: `sensor_hub/api/middleware/csrf_middleware_test.go`

**Step 1: Write tests for CSRFMiddleware**

Tests should cover:
- GET requests (bypass)
- POST to /auth/login (bypass)
- Valid CSRF token
- Missing/invalid token
- Mismatched token

```go
package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddleware_GET_Bypass(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/data", nil)

	CSRFMiddleware()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_Login_Bypass(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/login", nil)

	CSRFMiddleware()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_ValidToken(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService) // Assuming this sets the global authService

	mockService.On("GetCSRFForToken", "valid-token").Return("csrf-secret", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/data", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	c.Request.Header.Set("X-CSRF-Token", "csrf-secret")

	CSRFMiddleware()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_MissingCookie(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/data", nil)

	CSRFMiddleware()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCSRFMiddleware_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/data", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})

	CSRFMiddleware()(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_MismatchedToken(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService)

	mockService.On("GetCSRFForToken", "valid-token").Return("server-secret", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/data", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	c.Request.Header.Set("X-CSRF-Token", "client-secret")

	CSRFMiddleware()(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_ServiceError(t *testing.T) {
	mockService := new(MockAuthService)
	InitAuthMiddleware(mockService)

	mockService.On("GetCSRFForToken", "valid-token").Return("", errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/data", nil)
	c.Request.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	c.Request.Header.Set("X-CSRF-Token", "csrf-secret")

	CSRFMiddleware()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 2: Run tests**

Run: `go test ./sensor_hub/api/middleware/... -v`
Expected: PASS

---

### Task 4: Implement Permission Middleware Tests

**Files:**
- Create: `sensor_hub/api/middleware/permission_middleware_test.go`

**Step 1: Write tests for RequirePermission**

Tests should cover:
- User with permission (cached in user struct)
- User with permission (fetched from DB)
- User without permission
- DB error

```go
package middleware

import (
	"errors"
	"example/sensorHub/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequirePermission_Cached(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := &types.User{Id: 1, Permissions: []string{"test_perm"}}
	c.Set("currentUser", user)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_FromDB(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	InitPermissionMiddleware(mockRepo)

	user := &types.User{Id: 1} // No permissions cached
	mockRepo.On("GetPermissionsForUser", 1).Return([]string{"test_perm"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", user)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := &types.User{Id: 1, Permissions: []string{"other_perm"}}
	c.Set("currentUser", user)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_NoUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_DBError(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	InitPermissionMiddleware(mockRepo)

	user := &types.User{Id: 1}
	mockRepo.On("GetPermissionsForUser", 1).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", user)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 2: Run tests**

Run: `go test ./sensor_hub/api/middleware/... -v`
Expected: PASS

