package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupHealthRouter() *gin.Engine {
	router := gin.New()
	s := new(Server)
	router.GET("/api/health", s.GetHealth)
	return router
}

func TestGetHealth_ReturnsOK(t *testing.T) {
	router := setupHealthRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status"`)
	assert.Contains(t, w.Body.String(), `"ok"`)
}
