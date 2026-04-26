package middleware

import (
	"errors"
	gen "example/sensorHub/gen"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequirePermission_Cached(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := &gen.User{Id: 1, Permissions: []string{"test_perm"}}
	c.Set("currentUser", user)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_FromDB(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	InitPermissionMiddleware(mockRepo)

	user := &gen.User{Id: 1} // No permissions cached
	mockRepo.On("GetPermissionsForUser", mock.Anything, 1).Return([]string{"test_perm"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("currentUser", user)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := &gen.User{Id: 1, Permissions: []string{"other_perm"}}
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

	user := &gen.User{Id: 1}
	mockRepo.On("GetPermissionsForUser", mock.Anything, 1).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("currentUser", user)

	RequirePermission("test_perm")(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
