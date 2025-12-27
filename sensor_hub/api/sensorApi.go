package api

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/service"
	"example/sensorHub/types"
	"example/sensorHub/ws"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var sensorService service.SensorServiceInterface

func InitSensorAPI(s service.SensorServiceInterface) {
	sensorService = s
}

func addSensorHandler(ctx *gin.Context) {
	var sensor types.Sensor
	if err := ctx.BindJSON(&sensor); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	err := sensorService.ServiceAddSensor(sensor)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error adding sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusCreated, gin.H{"message": "Sensor added successfully"})
}

func updateSensorHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor ID is required"})
		return
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}
	var sensor types.Sensor
	if err := ctx.BindJSON(&sensor); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	sensor.Id = idInt
	err = sensorService.ServiceUpdateSensorById(sensor)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor updated successfully"})
}

func deleteSensorHandler(ctx *gin.Context) {
	sensorName := ctx.Param("name")

	if sensorName == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	err := sensorService.ServiceDeleteSensorByName(sensorName)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}

func getSensorByNameHandler(ctx *gin.Context) {
	sensorName := ctx.Param("name")
	if sensorName == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	sensor, err := sensorService.ServiceGetSensorByName(sensorName)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensor", "error": err.Error()})
		return
	}
	if sensor == nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "Sensor not found"})
		return
	}
	ctx.IndentedJSON(http.StatusOK, sensor)
}

func getAllSensorsHandler(ctx *gin.Context) {
	sensors, err := sensorService.ServiceGetAllSensors()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensors", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, sensors)
}

func getSensorsByTypeHandler(ctx *gin.Context) {
	sensorType := ctx.Param("type")
	if sensorType == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor type is required"})
		return
	}
	sensors, err := sensorService.ServiceGetSensorsByType(sensorType)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensors by type", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, sensors)
}

func sensorExistsHandler(ctx *gin.Context) {
	sensorName := ctx.Param("name")
	if sensorName == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	exists, err := sensorService.ServiceSensorExists(sensorName)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error checking sensor existence", "error": err.Error()})
		return
	}
	if exists {
		ctx.Status(http.StatusOK)
	} else {
		ctx.Status(http.StatusNotFound)
	}
}

func collectAndStoreAllSensorReadingsHandler(ctx *gin.Context) {
	err := sensorService.ServiceCollectAndStoreAllSensorReadings()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error collecting sensor readings", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor readings collected and stored successfully"})
}

func collectFromSensorByNameHandler(ctx *gin.Context) {
	sensorName := ctx.Param("sensorName")
	if sensorName == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}

	err := sensorService.ServiceCollectFromSensorByName(sensorName)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error collecting from sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor reading collected successfully"})
}

func disableSensorHandler(ctx *gin.Context) {
	sensorName := ctx.Param("sensorName")
	if sensorName == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	err := sensorService.ServiceSetEnabledSensorByName(sensorName, false)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error disabling sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor disabled successfully"})
}

func enableSensorHandler(ctx *gin.Context) {
	sensorName := ctx.Param("sensorName")
	if sensorName == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}
	err := sensorService.ServiceSetEnabledSensorByName(sensorName, true)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error enabling sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Sensor enabled successfully"})
}

func sensorWebSocketHandler(ctx *gin.Context) {
	sensorType := ctx.Param("type")
	if sensorType == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor type is required"})
		return
	}

	sensors, err := sensorService.ServiceGetSensorsByType(sensorType)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensors by type", "error": err.Error()})
		return
	}

	if len(sensors) == 0 {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "No sensors found for the specified type"})
		return
	}

	topic := "sensors:" + sensorType
	createPushWebSocket(ctx, topic)

	ws.BroadcastToTopic(topic, sensors)
}

func getSensorHealthHistoryByNameHandler(ctx *gin.Context) {
	sensorName := ctx.Param("name")
	if sensorName == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Sensor name is required"})
		return
	}

	limitStr := ctx.DefaultQuery("limit", strconv.Itoa(appProps.AppConfig.HealthHistoryDefaultResponseNumber))
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter"})
		return
	}

	healthHistory, err := sensorService.ServiceGetSensorHealthHistoryByName(sensorName, limit)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving sensor health history", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, healthHistory)
}

func totalReadingsPerSensorHandler(ctx *gin.Context) {
	stats, err := sensorService.ServiceGetTotalReadingsForEachSensor()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving total readings per sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, stats)
}

func RegisterSensorRoutes(router *gin.Engine) {
	sensorsGroup := router.Group("/sensors")
	{
		sensorsGroup.POST("/", addSensorHandler)
		sensorsGroup.PUT("/:id", updateSensorHandler)
		sensorsGroup.DELETE("/:name", deleteSensorHandler)
		sensorsGroup.GET("/:name", getSensorByNameHandler)
		sensorsGroup.GET("/", getAllSensorsHandler)
		sensorsGroup.GET("/type/:type", getSensorsByTypeHandler)
		sensorsGroup.HEAD("/:name", sensorExistsHandler)
		sensorsGroup.POST("/collect", collectAndStoreAllSensorReadingsHandler)
		sensorsGroup.POST("/collect/:sensorName", collectFromSensorByNameHandler)
		sensorsGroup.POST("/disable/:sensorName", disableSensorHandler)
		sensorsGroup.POST("/enable/:sensorName", enableSensorHandler)
		sensorsGroup.GET("/ws/:type", sensorWebSocketHandler)
		sensorsGroup.GET("/health/:name", getSensorHealthHistoryByNameHandler)
		sensorsGroup.GET("/stats/total-readings", totalReadingsPerSensorHandler)
	}
}
