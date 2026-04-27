package api

import (
	"bytes"
	"encoding/json"
	"errors"
	db "example/sensorHub/db"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRoleRouter() (*gin.Engine, *gin.RouterGroup, *Server, *MockRoleService) {
	mockService := new(MockRoleService)
	s := &Server{roleService: mockService}
	router := gin.New()
	apiGroup := router.Group("/api")
	return router, apiGroup, s, mockService
}

func TestListRolesHandler(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.GET("/roles", s.listRolesHandler)

	mockService.On("ListRoles", mock.Anything).Return([]db.RoleInfo{{Id: 1, Name: "admin"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin")
}

func TestListPermissionsHandler(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.GET("/permissions", s.listPermissionsHandler)

	mockService.On("ListPermissions", mock.Anything).Return([]db.PermissionInfo{{Id: 1, Name: "read"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "read")
}

func TestAssignPermissionHandler(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.POST("/roles/:id/permissions", s.assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("AssignPermission", mock.Anything, 1, 10).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/roles/1/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRemovePermissionHandler(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.DELETE("/roles/:id/permissions/:pid", s.removePermissionHandler)

	mockService.On("RemovePermission", mock.Anything, 1, 10).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/roles/1/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetRolePermissionsHandler(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.GET("/roles/:id/permissions", s.getRolePermissionsHandler)

	mockService.On("ListPermissionsForRole", mock.Anything, 1).Return([]db.PermissionInfo{{Id: 10, Name: "read"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/roles/1/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "read")
}

func TestListRolesHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.GET("/roles", s.listRolesHandler)

	mockService.On("ListRoles", mock.Anything).Return([]db.RoleInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListPermissionsHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.GET("/permissions", s.listPermissionsHandler)

	mockService.On("ListPermissions", mock.Anything).Return([]db.PermissionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetRolePermissionsHandler_InvalidID(t *testing.T) {
	router, api, s, _ := setupRoleRouter()
	api.GET("/roles/:id/permissions", s.getRolePermissionsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/roles/invalid/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetRolePermissionsHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.GET("/roles/:id/permissions", s.getRolePermissionsHandler)

	mockService.On("ListPermissionsForRole", mock.Anything, 1).Return([]db.PermissionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/roles/1/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssignPermissionHandler_InvalidID(t *testing.T) {
	router, api, s, _ := setupRoleRouter()
	api.POST("/roles/:id/permissions", s.assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/roles/invalid/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignPermissionHandler_InvalidJSON(t *testing.T) {
	router, api, s, _ := setupRoleRouter()
	api.POST("/roles/:id/permissions", s.assignPermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/roles/1/permissions", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignPermissionHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.POST("/roles/:id/permissions", s.assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("AssignPermission", mock.Anything, 1, 10).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/roles/1/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRemovePermissionHandler_InvalidRoleID(t *testing.T) {
	router, api, s, _ := setupRoleRouter()
	api.DELETE("/roles/:id/permissions/:pid", s.removePermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/roles/invalid/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemovePermissionHandler_InvalidPermissionID(t *testing.T) {
	router, api, s, _ := setupRoleRouter()
	api.DELETE("/roles/:id/permissions/:pid", s.removePermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/roles/1/permissions/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemovePermissionHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupRoleRouter()
	api.DELETE("/roles/:id/permissions/:pid", s.removePermissionHandler)

	mockService.On("RemovePermission", mock.Anything, 1, 10).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/roles/1/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
