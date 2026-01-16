package middleware

import (
	"example/sensorHub/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
