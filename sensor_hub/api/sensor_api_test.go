package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	appProps "example/sensorHub/application_properties"
	gen "example/sensorHub/gen"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
func init() {
	appProps.AppConfig = &appProps.ApplicationConfiguration{
		HealthHistoryDefaultResponseNumber: 10,
	}
}

func setupSensorRouter() (*gin.Engine, *gin.RouterGroup, *Server, *MockSensorService) {
	mockService := new(MockSensorService)
	s := &Server{sensorService: mockService}
	router := gin.New()
	apiGroup := router.Group("/api")
	return router, apiGroup, s, mockService
}

func TestAddSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors", s.AddSensor)

	sensor := gen.Sensor{Name: "test-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)

	mockService.On("ServiceAddSensor", mock.Anything, sensor).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetAllSensorsHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors", s.GetAllSensors)

	mockService.On("ServiceGetAllSensors", mock.Anything).Return([]gen.Sensor{{Name: "s1"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestGetSensorByNameHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/:name", func(c *gin.Context) {
		s.GetSensorByName(c, c.Param("name"))
	})

	mockService.On("ServiceGetSensorByName", mock.Anything, "s1").Return(&gen.Sensor{Name: "s1"}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
	assert.Contains(t, w.Body.String(), "effective_retention_hours")
}

func TestUpdateSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.PUT("/sensors/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.UpdateSensorById(c, id)
	})

	existing := gen.Sensor{Id: 1, Name: "s1", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}, Enabled: true}
	update := map[string]interface{}{"name": "s1-updated", "sensor_driver": "sensor-hub-http-temperature", "config": map[string]interface{}{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(update)

	expected := existing
	expected.Name = "s1-updated"

	mockService.On("ServiceGetSensorById", mock.Anything, 1).Return(&existing, nil)
	mockService.On("ServiceUpdateSensorById", mock.Anything, expected, mock.AnythingOfType("bool")).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/1", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.DELETE("/sensors/:name", func(c *gin.Context) {
		s.DeleteSensorByName(c, c.Param("name"))
	})

	mockService.On("ServiceDeleteSensorByName", mock.Anything, "s1").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCollectAndStoreAllSensorReadingsHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/collect", s.CollectAllSensorReadings)

	mockService.On("ServiceCollectAndStoreAllSensorReadings", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCollectFromSensorByNameHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/collect", func(c *gin.Context) {
		s.CollectFromSensor(c, c.Param("sensorName"))
	})

	mockService.On("ServiceCollectFromSensorByName", mock.Anything, "s1").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnableSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/enable", func(c *gin.Context) {
		s.EnableSensor(c, c.Param("sensorName"))
	})

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", true).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDisableSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/disable", func(c *gin.Context) {
		s.DisableSensor(c, c.Param("sensorName"))
	})

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", false).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTotalReadingsPerSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/readings/total", s.GetTotalReadingsPerSensor)

	mockService.On("ServiceGetTotalReadingsForEachSensor", mock.Anything).Return(map[string]int{"s1": 10}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/readings/total", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "10")
}

func TestGetSensorsByDriverHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/driver/:driver", func(c *gin.Context) {
		s.GetSensorsByDriver(c, c.Param("driver"))
	})

	mockService.On("ServiceGetSensorsByDriver", mock.Anything, "sensor-hub-http-temperature").Return([]gen.Sensor{{Name: "s1"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/driver/sensor-hub-http-temperature", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestSensorExistsHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.HEAD("/sensors/:name", func(c *gin.Context) {
		s.SensorExists(c, c.Param("name"))
	})

	mockService.On("ServiceSensorExists", mock.Anything, "s1").Return(true, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/:name/health", func(c *gin.Context) {
		var params gen.GetSensorHealthHistoryByNameParams
		if limitStr := c.Query("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter"})
				return
			}
			params.Limit = &limit
		}
		s.GetSensorHealthHistoryByName(c, c.Param("name"), params)
	})

	mockService.On("ServiceGetSensorHealthHistoryByName", mock.Anything, "s1", 10).Return([]gen.SensorHealthHistory{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSensorHandler_InvalidJSON(t *testing.T) {
	router, api, s, _ := setupSensorRouter()
	api.POST("/sensors", s.AddSensor)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddSensorHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors", s.AddSensor)

	sensor := gen.Sensor{Name: "test-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)

	mockService.On("ServiceAddSensor", mock.Anything, sensor).Return(errors.New("validation error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorByNameHandler_NotFound(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/:name", func(c *gin.Context) {
		s.GetSensorByName(c, c.Param("name"))
	})

	mockService.On("ServiceGetSensorByName", mock.Anything, "notfound").Return((*gen.Sensor)(nil), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSensorByNameHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/:name", func(c *gin.Context) {
		s.GetSensorByName(c, c.Param("name"))
	})

	mockService.On("ServiceGetSensorByName", mock.Anything, "s1").Return((*gen.Sensor)(nil), errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateSensorHandler_InvalidID(t *testing.T) {
	router, api, s, _ := setupSensorRouter()
	api.PUT("/sensors/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.UpdateSensorById(c, id)
	})

	sensor := gen.Sensor{Name: "s1-updated", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/invalid", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_InvalidJSON(t *testing.T) {
	router, api, s, _ := setupSensorRouter()
	api.PUT("/sensors/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.UpdateSensorById(c, id)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/1", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.PUT("/sensors/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.UpdateSensorById(c, id)
	})

	existing := gen.Sensor{Id: 1, Name: "s1", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	update := map[string]interface{}{"name": "s1-updated", "sensor_driver": "sensor-hub-http-temperature", "config": map[string]interface{}{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(update)

	expected := existing
	expected.Name = "s1-updated"

	mockService.On("ServiceGetSensorById", mock.Anything, 1).Return(&existing, nil)
	mockService.On("ServiceUpdateSensorById", mock.Anything, expected, mock.AnythingOfType("bool")).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/1", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteSensorHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.DELETE("/sensors/:name", func(c *gin.Context) {
		s.DeleteSensorByName(c, c.Param("name"))
	})

	mockService.On("ServiceDeleteSensorByName", mock.Anything, "s1").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAllSensorsHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors", s.GetAllSensors)

	mockService.On("ServiceGetAllSensors", mock.Anything).Return([]gen.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorsByDriverHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/driver/:driver", func(c *gin.Context) {
		s.GetSensorsByDriver(c, c.Param("driver"))
	})

	mockService.On("ServiceGetSensorsByDriver", mock.Anything, "sensor-hub-http-temperature").Return([]gen.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/driver/sensor-hub-http-temperature", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSensorExistsHandler_NotFound(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.HEAD("/sensors/:name", func(c *gin.Context) {
		s.SensorExists(c, c.Param("name"))
	})

	mockService.On("ServiceSensorExists", mock.Anything, "notfound").Return(false, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/api/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSensorExistsHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.HEAD("/sensors/:name", func(c *gin.Context) {
		s.SensorExists(c, c.Param("name"))
	})

	mockService.On("ServiceSensorExists", mock.Anything, "s1").Return(false, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectAndStoreAllSensorReadingsHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/collect", s.CollectAllSensorReadings)

	mockService.On("ServiceCollectAndStoreAllSensorReadings", mock.Anything).Return(errors.New("collection error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectFromSensorByNameHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/collect", func(c *gin.Context) {
		s.CollectFromSensor(c, c.Param("sensorName"))
	})

	mockService.On("ServiceCollectFromSensorByName", mock.Anything, "s1").Return(errors.New("sensor offline"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEnableSensorHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/enable", func(c *gin.Context) {
		s.EnableSensor(c, c.Param("sensorName"))
	})

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", true).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDisableSensorHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/disable", func(c *gin.Context) {
		s.DisableSensor(c, c.Param("sensorName"))
	})

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", false).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTotalReadingsPerSensorHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/readings/total", s.GetTotalReadingsPerSensor)

	mockService.On("ServiceGetTotalReadingsForEachSensor", mock.Anything).Return(map[string]int{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/readings/total", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_InvalidLimit(t *testing.T) {
	router, api, s, _ := setupSensorRouter()
	api.GET("/sensors/:name/health", func(c *gin.Context) {
		var params gen.GetSensorHealthHistoryByNameParams
		if limitStr := c.Query("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter"})
				return
			}
			params.Limit = &limit
		}
		s.GetSensorHealthHistoryByName(c, c.Param("name"), params)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1/health?limit=invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_ServiceError(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/:name/health", func(c *gin.Context) {
		var params gen.GetSensorHealthHistoryByNameParams
		if limitStr := c.Query("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter"})
				return
			}
			params.Limit = &limit
		}
		s.GetSensorHealthHistoryByName(c, c.Param("name"), params)
	})

	mockService.On("ServiceGetSensorHealthHistoryByName", mock.Anything, "s1", 10).Return([]gen.SensorHealthHistory{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ============================================================================
// Sensor Status Handlers
// ============================================================================

func TestGetSensorsByStatusHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/status/:status", func(c *gin.Context) {
		s.GetSensorsByStatus(c, gen.GetSensorsByStatusParamsStatus(c.Param("status")))
	})

	mockService.On("ServiceGetSensorsByStatus", mock.Anything, "pending").Return([]gen.Sensor{
		{Id: 1, Name: "auto-sensor", Status: "pending"},
	}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/status/pending", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "auto-sensor")
}

func TestGetSensorsByStatusHandler_Empty(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/status/:status", func(c *gin.Context) {
		s.GetSensorsByStatus(c, gen.GetSensorsByStatusParamsStatus(c.Param("status")))
	})

	mockService.On("ServiceGetSensorsByStatus", mock.Anything, "pending").Return(nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/status/pending", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "[]")
}

func TestGetSensorsByStatusHandler_Error(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.GET("/sensors/status/:status", func(c *gin.Context) {
		s.GetSensorsByStatus(c, gen.GetSensorsByStatusParamsStatus(c.Param("status")))
	})

	mockService.On("ServiceGetSensorsByStatus", mock.Anything, "pending").Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/status/pending", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestApproveSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/approve/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.ApproveSensor(c, id)
	})

	mockService.On("ServiceApproveSensor", mock.Anything, 1).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/approve/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "approved")
}

func TestApproveSensorHandler_InvalidID(t *testing.T) {
	router, api, s, _ := setupSensorRouter()
	api.POST("/sensors/approve/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.ApproveSensor(c, id)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/approve/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApproveSensorHandler_Error(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/approve/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.ApproveSensor(c, id)
	})

	mockService.On("ServiceApproveSensor", mock.Anything, 1).Return(errors.New("not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/approve/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDismissSensorHandler(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/dismiss/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.DismissSensor(c, id)
	})

	mockService.On("ServiceDismissSensor", mock.Anything, 1).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/dismiss/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "dismissed")
}

func TestDismissSensorHandler_InvalidID(t *testing.T) {
	router, api, s, _ := setupSensorRouter()
	api.POST("/sensors/dismiss/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.DismissSensor(c, id)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/dismiss/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDismissSensorHandler_Error(t *testing.T) {
	router, api, s, mockService := setupSensorRouter()
	api.POST("/sensors/dismiss/:id", func(c *gin.Context) {
		var id int
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
			return
		}
		s.DismissSensor(c, id)
	})

	mockService.On("ServiceDismissSensor", mock.Anything, 1).Return(errors.New("not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/dismiss/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAllMeasurementTypes(t *testing.T) {
router, _, s, mockService := setupSensorRouter()
router.GET("/api/measurement-types", func(c *gin.Context) {
var params gen.GetAllMeasurementTypesParams
s.GetAllMeasurementTypes(c, params)
})

mts := []gen.MeasurementType{{Name: "temperature"}}
mockService.On("ServiceGetAllMeasurementTypes", mock.Anything).Return(mts, nil)

w := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/api/measurement-types", nil)
router.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "temperature")
}

func TestGetAllMeasurementTypes_HasReadings(t *testing.T) {
router, _, s, mockService := setupSensorRouter()
router.GET("/api/measurement-types", func(c *gin.Context) {
hasReadings := c.Query("has_readings") == "true"
params := gen.GetAllMeasurementTypesParams{HasReadings: &hasReadings}
s.GetAllMeasurementTypes(c, params)
})

mts := []gen.MeasurementType{{Name: "humidity"}}
mockService.On("ServiceGetAllMeasurementTypesWithReadings", mock.Anything).Return(mts, nil)

w := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/api/measurement-types?has_readings=true", nil)
router.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "humidity")
}

func TestGetSensorMeasurementTypes(t *testing.T) {
router, _, s, mockService := setupSensorRouter()
router.GET("/api/sensors/:id/measurement-types", func(c *gin.Context) {
var id int
fmt.Sscan(c.Param("id"), &id)
s.GetSensorMeasurementTypes(c, id)
})

mts := []gen.MeasurementType{{Name: "temperature"}}
mockService.On("ServiceGetMeasurementTypesForSensor", mock.Anything, 1).Return(mts, nil)

w := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/api/sensors/1/measurement-types", nil)
router.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "temperature")
}
