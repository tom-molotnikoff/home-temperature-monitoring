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
