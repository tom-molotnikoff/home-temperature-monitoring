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
)

func setupRoleRouter() (*gin.Engine, *MockRoleService) {
	mockService := new(MockRoleService)
	InitRolesAPI(mockService)
	router := gin.New()
	return router, mockService
}

func TestListRolesHandler(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/roles", listRolesHandler)

	mockService.On("ListRoles").Return([]db.RoleInfo{{Id: 1, Name: "admin"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin")
}

func TestListPermissionsHandler(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/permissions", listPermissionsHandler)

	mockService.On("ListPermissions").Return([]db.PermissionInfo{{Id: 1, Name: "read"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "read")
}

func TestAssignPermissionHandler(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.POST("/roles/:id/permissions", assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("AssignPermission", 1, 10).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/roles/1/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRemovePermissionHandler(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.DELETE("/roles/:id/permissions/:pid", removePermissionHandler)

	mockService.On("RemovePermission", 1, 10).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/roles/1/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetRolePermissionsHandler(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/roles/:id/permissions", getRolePermissionsHandler)

	mockService.On("ListPermissionsForRole", 1).Return([]db.PermissionInfo{{Id: 10, Name: "read"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles/1/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "read")
}

func TestListRolesHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/roles", listRolesHandler)

	mockService.On("ListRoles").Return([]db.RoleInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListPermissionsHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/permissions", listPermissionsHandler)

	mockService.On("ListPermissions").Return([]db.PermissionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetRolePermissionsHandler_InvalidID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.GET("/roles/:id/permissions", getRolePermissionsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles/invalid/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetRolePermissionsHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/roles/:id/permissions", getRolePermissionsHandler)

	mockService.On("ListPermissionsForRole", 1).Return([]db.PermissionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles/1/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssignPermissionHandler_InvalidID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.POST("/roles/:id/permissions", assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/roles/invalid/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignPermissionHandler_InvalidJSON(t *testing.T) {
	router, _ := setupRoleRouter()
	router.POST("/roles/:id/permissions", assignPermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/roles/1/permissions", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignPermissionHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.POST("/roles/:id/permissions", assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("AssignPermission", 1, 10).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/roles/1/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRemovePermissionHandler_InvalidRoleID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.DELETE("/roles/:id/permissions/:pid", removePermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/roles/invalid/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemovePermissionHandler_InvalidPermissionID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.DELETE("/roles/:id/permissions/:pid", removePermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/roles/1/permissions/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemovePermissionHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.DELETE("/roles/:id/permissions/:pid", removePermissionHandler)

	mockService.On("RemovePermission", 1, 10).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/roles/1/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
