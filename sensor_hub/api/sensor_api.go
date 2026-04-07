package api

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/drivers"
	"example/sensorHub/service"
	"example/sensorHub/types"
	"example/sensorHub/ws"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var sensorService service.SensorServiceInterface

func InitSensorAPI(s service.SensorServiceInterface) {
	sensorService = s
}

func addSensorHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var sensor types.Sensor
	if err := c.BindJSON(&sensor); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	err := sensorService.ServiceAddSensor(ctx, sensor)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error adding sensor", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Sensor added successfully"})
}

func updateSensorHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	if idStr == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor ID is required"})
		return
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	// Parse as raw map to support merge-patch semantics on config
	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Build sensor from body fields
	var sensor types.Sensor
	sensor.Id = idInt
	if name, ok := body["name"].(string); ok {
		sensor.Name = name
	}
	if driver, ok := body["sensor_driver"].(string); ok {
		sensor.SensorDriver = driver
	}
	if enabled, ok := body["enabled"].(bool); ok {
		sensor.Enabled = enabled
	}

	// Handle config with merge-patch semantics
	if rawConfig, exists := body["config"]; exists {
		sensor.Config = make(map[string]string)
		if configMap, ok := rawConfig.(map[string]interface{}); ok {
			for k, v := range configMap {
				if v == nil {
					// null means delete key — skip it
					continue
				}
				if strVal, ok := v.(string); ok {
					// Skip "****" for sensitive fields — means "keep existing"
					if strVal == "****" {
						continue
					}
					sensor.Config[k] = strVal
				}
			}
		}
	}

	err = sensorService.ServiceUpdateSensorById(ctx, sensor)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating sensor", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor updated successfully"})
}

func deleteSensorHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorName := c.Param("name")

	if sensorName == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	err := sensorService.ServiceDeleteSensorByName(ctx, sensorName)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting sensor", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}

func getSensorByNameHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorName := c.Param("name")
	if sensorName == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	sensor, err := sensorService.ServiceGetSensorByName(ctx, sensorName)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensor", "error": err.Error()})
		return
	}
	if sensor == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Sensor not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, maskSensitiveConfig(*sensor))
}

func getAllSensorsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensors, err := sensorService.ServiceGetAllSensors(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensors", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, maskSensitiveConfigSlice(sensors))
}

func getSensorsByDriverHandler(c *gin.Context) {
	ctx := c.Request.Context()
	driver := c.Param("driver")
	if driver == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor driver is required"})
		return
	}
	sensors, err := sensorService.ServiceGetSensorsByDriver(ctx, driver)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensors by driver", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, maskSensitiveConfigSlice(sensors))
}

func sensorExistsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorName := c.Param("name")
	if sensorName == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	exists, err := sensorService.ServiceSensorExists(ctx, sensorName)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error checking sensor existence", "error": err.Error()})
		return
	}
	if exists {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func collectAndStoreAllSensorReadingsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	err := sensorService.ServiceCollectAndStoreAllSensorReadings(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error collecting sensor readings", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor readings collected and stored successfully"})
}

func collectFromSensorByNameHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorName := c.Param("sensorName")
	if sensorName == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}

	err := sensorService.ServiceCollectFromSensorByName(ctx, sensorName)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error collecting from sensor", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor reading collected successfully"})
}

func disableSensorHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorName := c.Param("sensorName")
	if sensorName == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	err := sensorService.ServiceSetEnabledSensorByName(ctx, sensorName, false)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error disabling sensor", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor disabled successfully"})
}

func enableSensorHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorName := c.Param("sensorName")
	if sensorName == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	err := sensorService.ServiceSetEnabledSensorByName(ctx, sensorName, true)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error enabling sensor", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor enabled successfully"})
}

func sensorWebSocketHandler(c *gin.Context) {
	ctx := c.Request.Context()
	driver := c.Param("driver")
	if driver == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor driver is required"})
		return
	}

	topic := "sensors:" + driver
	createPushWebSocket(c, topic)

	sensors, err := sensorService.ServiceGetSensorsByDriver(ctx, driver)
	if err != nil {
		slog.Error("error retrieving sensors by driver for WebSocket broadcast", "driver", driver, "error", err)
		return
	}

	ws.BroadcastToTopic(topic, sensors)
}

func getSensorHealthHistoryByNameHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorName := c.Param("name")
	if sensorName == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", strconv.Itoa(appProps.AppConfig.HealthHistoryDefaultResponseNumber))
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter"})
		return
	}

	healthHistory, err := sensorService.ServiceGetSensorHealthHistoryByName(ctx, sensorName, limit)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensor health history", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, healthHistory)
}

func totalReadingsPerSensorHandler(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := sensorService.ServiceGetTotalReadingsForEachSensor(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving total readings per sensor", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, stats)
}

// maskSensitiveConfig returns a copy of the sensor with sensitive config fields masked.
func maskSensitiveConfig(sensor types.Sensor) types.Sensor {
	driver, ok := drivers.Get(sensor.SensorDriver)
	if !ok {
		return sensor
	}
	sensitiveKeys := make(map[string]bool)
	for _, f := range driver.ConfigFields() {
		if f.Sensitive {
			sensitiveKeys[f.Key] = true
		}
	}
	if len(sensitiveKeys) > 0 && len(sensor.Config) > 0 {
		masked := make(map[string]string, len(sensor.Config))
		for k, v := range sensor.Config {
			if sensitiveKeys[k] && v != "" {
				masked[k] = "****"
			} else {
				masked[k] = v
			}
		}
		sensor.Config = masked
	}
	return sensor
}

// maskSensitiveConfigSlice masks sensitive config fields in a slice of sensors.
func maskSensitiveConfigSlice(sensors []types.Sensor) []types.Sensor {
	result := make([]types.Sensor, len(sensors))
	for i, s := range sensors {
		result[i] = maskSensitiveConfig(s)
	}
	return result
}
