package api

import (
	"bytes"
	"encoding/json"
	"errors"
	db "example/sensorHub/db"
	"example/sensorHub/service"
	"example/sensorHub/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthRouter() (*gin.Engine, *MockAuthService) {
	mockService := new(MockAuthService)
	InitAuthAPI(mockService)
	router := gin.New()
	return router, mockService
}

func TestLoginHandler_Success(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/login", loginHandler)

	reqBody := loginRequest{Username: "user", Password: "password"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("Login", "user", "password", mock.Anything, mock.Anything).Return("token", "csrf", false, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "csrf")
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/login", loginHandler)

	reqBody := loginRequest{Username: "user", Password: "wrong"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("Login", "user", "wrong", mock.Anything, mock.Anything).Return("", "", false, errors.New("invalid credentials"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_TooManyAttempts(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/login", loginHandler)

	reqBody := loginRequest{Username: "user", Password: "password"}
	jsonBody, _ := json.Marshal(reqBody)

	err := &service.TooManyAttemptsError{RetryAfterSeconds: 60}
	mockService.On("Login", "user", "password", mock.Anything, mock.Anything).Return("", "", false, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestLogoutHandler(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/logout", logoutHandler)

	mockService.On("Logout", "valid-token").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMeHandler_Success(t *testing.T) {
	router, mockService := setupAuthRouter()
	// Middleware normally sets currentUser, but here we mock it or set it manually if middleware isn't used
	// In meHandler, it expects "currentUser" in context.
	router.GET("/auth/me", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "me"})
		meHandler(c)
	})

	mockService.On("GetCSRFForToken", "valid-token").Return("csrf-token", nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "me")
}

func TestListSessionsHandler(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.GET("/auth/sessions", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		listSessionsHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100}}, nil)
	mockService.On("GetSessionIdForToken", "valid-token").Return(int64(100), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/sessions", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "100")
}

func TestRevokeSessionHandler_OwnSession(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100, UserId: 1}}, nil)
	revoker := 1
	mockService.On("RevokeSessionByIdWithActor", int64(100), &revoker, (*string)(nil)).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/100", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	router, _ := setupAuthRouter()
	router.POST("/auth/login", loginHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString("invalid-json"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestLoginHandler_MustChangePassword(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/login", loginHandler)

	reqBody := loginRequest{Username: "user", Password: "password"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("Login", "user", "password", mock.Anything, mock.Anything).Return("token", "csrf", true, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "must_change_password")
	assert.Contains(t, w.Body.String(), "true")
}

func TestLogoutHandler_MissingCookie(t *testing.T) {
	router, _ := setupAuthRouter()
	router.POST("/auth/logout", logoutHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	router.ServeHTTP(w, req)

	// Logout is idempotent - always returns OK even without cookie
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogoutHandler_ServiceError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/logout", logoutHandler)

	mockService.On("Logout", "valid-token").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	// Handler ignores service errors - always returns OK
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMeHandler_MissingCookie(t *testing.T) {
	router, _ := setupAuthRouter()
	router.GET("/auth/me", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "me"})
		meHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/me", nil)
	router.ServeHTTP(w, req)

	// Handler returns OK with empty csrf_token when no cookie
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "me")
}

func TestMeHandler_CSRFError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.GET("/auth/me", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "me"})
		meHandler(c)
	})

	mockService.On("GetCSRFForToken", "valid-token").Return("", errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	// Handler silently ignores CSRF errors and returns empty csrf_token
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "me")
}

func TestListSessionsHandler_ServiceError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.GET("/auth/sessions", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		listSessionsHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/sessions", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRevokeSessionHandler_InvalidID(t *testing.T) {
	router, _ := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeSessionHandler_MissingID(t *testing.T) {
	router, _ := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/", nil)
	router.ServeHTTP(w, req)

	// Gin won't match this route without :id, so it returns 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRevokeSessionHandler_NotOwnedSession(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100, UserId: 1}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/200", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRevokeSessionHandler_AdminRevokingOthers(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100, UserId: 1}}, nil)
	revoker := 1
	mockService.On("RevokeSessionByIdWithActor", int64(200), &revoker, (*string)(nil)).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/200", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRevokeSessionHandler_ListError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/100", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRevokeSessionHandler_RevokeError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100, UserId: 1}}, nil)
	revoker := 1
	mockService.On("RevokeSessionByIdWithActor", int64(100), &revoker, (*string)(nil)).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/100", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

