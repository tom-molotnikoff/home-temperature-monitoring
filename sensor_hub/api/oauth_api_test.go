package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"example/sensorHub/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupOAuthRouter() (*gin.Engine, *MockOAuthService) {
	mockService := new(MockOAuthService)
	InitOAuthAPI(mockService)
	router := gin.New()
	return router, mockService
}

func TestGetOAuthStatus_Success(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.GET("/oauth/status", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthStatusHandler(c)
	})

	mockService.On("GetStatus").Return(map[string]interface{}{
		"configured":  true,
		"needs_auth":  false,
		"token_valid": true,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "configured")
}

func TestGetOAuthStatus_ServiceUnavailable(t *testing.T) {
	router := gin.New()
	InitOAuthAPI(nil)
	router.GET("/oauth/status", oauthStatusHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestGetOAuthAuthURL_Success(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.GET("/oauth/authorize", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthAuthorizeHandler(c)
	})

	mockService.On("GetAuthURL", mock.Anything).Return("https://accounts.google.com/oauth?state=abc", nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/authorize", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "auth_url")
	assert.Contains(t, w.Body.String(), "state")
}

func TestGetOAuthAuthURL_Error(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.GET("/oauth/authorize", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthAuthorizeHandler(c)
	})

	mockService.On("GetAuthURL", mock.Anything).Return("", errors.New("not configured"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/authorize", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOAuthSubmitCode_Success(t *testing.T) {
	router, mockService := setupOAuthRouter()

	// First get a state token
	router.GET("/oauth/authorize", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthAuthorizeHandler(c)
	})
	router.POST("/oauth/submit-code", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthSubmitCodeHandler(c)
	})

	mockService.On("GetAuthURL", mock.Anything).Return("https://accounts.google.com/oauth", nil)
	mockService.On("ExchangeCode", "test-code").Return(nil)

	// Get auth URL to create a pending state
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/authorize", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var authResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &authResp)
	state := authResp["state"]

	// Submit code with state
	body, _ := json.Marshal(map[string]string{"code": "test-code", "state": state})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/oauth/submit-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "successful")
}

func TestOAuthSubmitCode_InvalidState(t *testing.T) {
	router, _ := setupOAuthRouter()
	router.POST("/oauth/submit-code", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthSubmitCodeHandler(c)
	})

	body, _ := json.Marshal(map[string]string{"code": "test-code", "state": "invalid-state"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/oauth/submit-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired state")
}

func TestOAuthSubmitCode_MissingFields(t *testing.T) {
	router, _ := setupOAuthRouter()
	router.POST("/oauth/submit-code", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthSubmitCodeHandler(c)
	})

	body, _ := json.Marshal(map[string]string{"code": "test-code"}) // missing state
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/oauth/submit-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOAuthSubmitCode_ExchangeError(t *testing.T) {
	router, mockService := setupOAuthRouter()

	router.GET("/oauth/authorize", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthAuthorizeHandler(c)
	})
	router.POST("/oauth/submit-code", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthSubmitCodeHandler(c)
	})

	mockService.On("GetAuthURL", mock.Anything).Return("https://accounts.google.com/oauth", nil)
	mockService.On("ExchangeCode", "bad-code").Return(errors.New("invalid code"))

	// Get state
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/oauth/authorize", nil)
	router.ServeHTTP(w, req)
	var authResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &authResp)
	state := authResp["state"]

	// Submit bad code
	body, _ := json.Marshal(map[string]string{"code": "bad-code", "state": state})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/oauth/submit-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOAuthReload_Success(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.POST("/oauth/reload", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthReloadHandler(c)
	})

	mockService.On("Reload").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/oauth/reload", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "reloaded")
	mockService.AssertExpectations(t)
}

func TestOAuthReload_Error(t *testing.T) {
	router, mockService := setupOAuthRouter()
	router.POST("/oauth/reload", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "admin", Roles: []string{"admin"}})
		oauthReloadHandler(c)
	})

	mockService.On("Reload").Return(errors.New("credentials not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/oauth/reload", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to reload")
}

func TestOAuthReload_ServiceUnavailable(t *testing.T) {
	router := gin.New()
	InitOAuthAPI(nil)
	router.POST("/oauth/reload", oauthReloadHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/oauth/reload", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
