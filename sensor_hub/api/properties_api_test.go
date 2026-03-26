package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupPropertiesRouter() (*gin.Engine, *gin.RouterGroup, *MockPropertiesService) {
	mockService := new(MockPropertiesService)
	InitPropertiesAPI(mockService)
	router := gin.New()
	apiGroup := router.Group("/api")
	return router, apiGroup, mockService
}

func TestUpdatePropertiesHandler(t *testing.T) {
	router, api, mockService := setupPropertiesRouter()
	api.PATCH("/properties", updatePropertiesHandler)

	props := map[string]string{"key": "value"}
	jsonBody, _ := json.Marshal(props)

	mockService.On("ServiceUpdateProperties", mock.Anything, props).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/properties", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestGetPropertiesHandler(t *testing.T) {
	router, api, mockService := setupPropertiesRouter()
	api.GET("/properties", getPropertiesHandler)

	mockService.On("ServiceGetProperties", mock.Anything).Return(map[string]interface{}{"key": "value"}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/properties", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "value")
}

func TestUpdatePropertiesHandler_InvalidJSON(t *testing.T) {
	router, api, _ := setupPropertiesRouter()
	api.PATCH("/properties", updatePropertiesHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/properties", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePropertiesHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupPropertiesRouter()
	api.PATCH("/properties", updatePropertiesHandler)

	props := map[string]string{"key": "value"}
	jsonBody, _ := json.Marshal(props)

	mockService.On("ServiceUpdateProperties", mock.Anything, props).Return(errors.New("validation error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/properties", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetPropertiesHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupPropertiesRouter()
	api.GET("/properties", getPropertiesHandler)

	mockService.On("ServiceGetProperties", mock.Anything).Return(map[string]interface{}{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/properties", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
