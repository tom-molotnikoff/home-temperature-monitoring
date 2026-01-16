package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"example/sensorHub/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserRouter() (*gin.Engine, *MockUserService) {
	mockService := new(MockUserService)
	InitUsersAPI(mockService)
	router := gin.New()
	return router, mockService
}

func TestCreateUserHandler_Success(t *testing.T) {
	router, mockService := setupUserRouter()
	router.POST("/users", createUserHandler)

	reqBody := createUserRequest{Username: "newuser", Password: "password"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("CreateUser", mock.AnythingOfType("types.User"), "password").Return(1, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestListUsersHandler(t *testing.T) {
	router, mockService := setupUserRouter()
	router.GET("/users", listUsersHandler)

	mockService.On("ListUsers").Return([]types.User{{Username: "u1"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "u1")
}

func TestChangePasswordHandler_Self(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 1, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", 1, "newpass", "valid-token").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteUserHandler_Admin(t *testing.T) {
	router, mockService := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		deleteUserHandler(c)
	})

	mockService.On("DeleteUser", 2).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetMustChangeHandler_Admin(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setMustChangeHandler(c)
	})

	reqBody := mustChangeRequest{MustChange: true}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("SetMustChangeFlag", 2, true).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/must-change-password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetRolesHandler_Admin(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setRolesHandler(c)
	})

	reqBody := setRolesRequest{Roles: []string{"admin"}}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("SetUserRoles", 2, []string{"admin"}).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/roles", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateUserHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.POST("/users", createUserHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUserHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.POST("/users", createUserHandler)

	reqBody := createUserRequest{Username: "newuser", Password: "password"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("CreateUser", mock.AnythingOfType("types.User"), "password").Return(0, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListUsersHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.GET("/users", listUsersHandler)

	mockService.On("ListUsers").Return([]types.User{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChangePasswordHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		changePasswordHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangePasswordHandler_DefaultsToCurrentUser(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 0, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", 1, "newpass", "valid-token").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChangePasswordHandler_AdminChangingOthers(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 2, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", 2, "newpass", "").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChangePasswordHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 2, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestChangePasswordHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 1, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", 1, "newpass", "valid-token").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteUserHandler_InvalidID(t *testing.T) {
	router, _ := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		deleteUserHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		deleteUserHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteUserHandler_CannotDeleteSelf(t *testing.T) {
	router, _ := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		deleteUserHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		deleteUserHandler(c)
	})

	mockService.On("DeleteUser", 2).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSetMustChangeHandler_InvalidID(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setMustChangeHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/invalid/must-change-password", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetMustChangeHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setMustChangeHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/must-change-password", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetMustChangeHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		setMustChangeHandler(c)
	})

	reqBody := mustChangeRequest{MustChange: true}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/must-change-password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSetMustChangeHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setMustChangeHandler(c)
	})

	reqBody := mustChangeRequest{MustChange: true}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("SetMustChangeFlag", 2, true).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/must-change-password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSetRolesHandler_InvalidID(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setRolesHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/invalid/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetRolesHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setRolesHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/roles", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetRolesHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		setRolesHandler(c)
	})

	reqBody := setRolesRequest{Roles: []string{"admin"}}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/roles", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSetRolesHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setRolesHandler(c)
	})

	reqBody := setRolesRequest{Roles: []string{"admin"}}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("SetUserRoles", 2, []string{"admin"}).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/roles", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
