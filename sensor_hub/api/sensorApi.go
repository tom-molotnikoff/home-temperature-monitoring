package api

import (
	"example/sensorHub/service"
	"example/sensorHub/types"

	"github.com/gin-gonic/gin"
)

var sensorService service.SensorServiceInterface

func InitSensorAPI(s service.SensorServiceInterface) {
	sensorService = s
}

func addSensorHandler(ctx *gin.Context) {
	var sensor types.Sensor
	if err := ctx.BindJSON(&sensor); err != nil {
		ctx.IndentedJSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	err := sensorService.ServiceAddSensor(sensor)
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error adding sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(201, gin.H{"message": "Sensor added successfully"})
}

func updateSensorHandler(ctx *gin.Context) {
	var sensor types.Sensor
	if err := ctx.BindJSON(&sensor); err != nil {
		ctx.IndentedJSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	err := sensorService.ServiceUpdateSensorByName(sensor)
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error updating sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(200, gin.H{"message": "Sensor updated successfully"})
}

func deleteSensorHandler(ctx *gin.Context) {
	sensorName := ctx.Param("name")
	if sensorName == "" {
		ctx.IndentedJSON(400, gin.H{"message": "Sensor name is required"})
		return
	}
	err := sensorService.ServiceDeleteSensorByName(sensorName)
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error deleting sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(200, gin.H{"message": "Sensor deleted successfully"})
}

func getSensorByNameHandler(ctx *gin.Context) {
	sensorName := ctx.Param("name")
	if sensorName == "" {
		ctx.IndentedJSON(400, gin.H{"message": "Sensor name is required"})
		return
	}
	sensor, err := sensorService.ServiceGetSensorByName(sensorName)
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error retrieving sensor", "error": err.Error()})
		return
	}
	if sensor == nil {
		ctx.IndentedJSON(404, gin.H{"message": "Sensor not found"})
		return
	}
	ctx.IndentedJSON(200, sensor)
}

func getAllSensorsHandler(ctx *gin.Context) {
	sensors, err := sensorService.ServiceGetAllSensors()
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error retrieving sensors", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(200, sensors)
}

func getSensorsByTypeHandler(ctx *gin.Context) {
	sensorType := ctx.Param("type")
	if sensorType == "" {
		ctx.IndentedJSON(400, gin.H{"message": "Sensor type is required"})
		return
	}
	sensors, err := sensorService.ServiceGetSensorsByType(sensorType)
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error retrieving sensors by type", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(200, sensors)
}

func sensorExistsHandler(ctx *gin.Context) {
	sensorName := ctx.Param("name")
	if sensorName == "" {
		ctx.IndentedJSON(400, gin.H{"message": "Sensor name is required"})
		return
	}
	exists, err := sensorService.ServiceSensorExists(sensorName)
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error checking sensor existence", "error": err.Error()})
		return
	}
	if exists {
		ctx.Status(200)
	} else {
		ctx.Status(404)
	}
}

func collectAndStoreAllSensorReadingsHandler(ctx *gin.Context) {
	err := sensorService.ServiceCollectAndStoreAllSensorReadings()
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error collecting sensor readings", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(200, gin.H{"message": "Sensor readings collected and stored successfully"})
}

func collectFromSensorByNameHandler(ctx *gin.Context) {
	sensorName := ctx.Param("sensorName")
	if sensorName == "" {
		ctx.IndentedJSON(400, gin.H{"message": "Sensor name is required"})
		return
	}

	err := sensorService.ServiceCollectFromSensorByName(sensorName)
	if err != nil {
		ctx.IndentedJSON(500, gin.H{"message": "Error collecting from sensor", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(200, gin.H{"message": "Sensor reading collected successfully"})
}

func RegisterSensorRoutes(router *gin.Engine) {
	sensorsGroup := router.Group("/sensors")
	{
		sensorsGroup.POST("/", addSensorHandler)
		sensorsGroup.PUT("/:name", updateSensorHandler)
		sensorsGroup.DELETE("/:name", deleteSensorHandler)
		sensorsGroup.GET("/:name", getSensorByNameHandler)
		sensorsGroup.GET("/", getAllSensorsHandler)
		sensorsGroup.GET("/type/:type", getSensorsByTypeHandler)
		sensorsGroup.HEAD("/:name", sensorExistsHandler)
		sensorsGroup.POST("/collect", collectAndStoreAllSensorReadingsHandler)
		sensorsGroup.POST("/collect/:sensorName", collectFromSensorByNameHandler)
	}
}
