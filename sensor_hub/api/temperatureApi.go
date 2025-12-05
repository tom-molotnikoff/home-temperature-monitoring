package api

import (
	"example/sensorHub/service"
	"example/sensorHub/types"
	"example/sensorHub/ws"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var tempService service.TemperatureServiceInterface

func InitTemperatureAPI(s service.TemperatureServiceInterface) {
	tempService = s
}

func getHourlyReadingsBetweenDatesHandler(ctx *gin.Context) {
	getReadingsBetweenDatesHelper(ctx, types.TableHourlyAverageTemperature)
}

func getReadingsBetweenDatesHandler(ctx *gin.Context) {
	getReadingsBetweenDatesHelper(ctx, types.TableTemperatureReadings)
}

func getReadingsBetweenDatesHelper(ctx *gin.Context, tableName string) {
	startDate := ctx.Query("start")
	endDate := ctx.Query("end")
	if startDate == "" || endDate == "" {
		log.Printf("Missing start or end date")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Start and end dates are required"})
		return
	}

	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Printf("Invalid start date format: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid start date format, expected YYYY-MM-DD"})
		return
	}

	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		log.Printf("Invalid end date format: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid end date format, expected YYYY-MM-DD"})
		return
	}

	log.Printf("Fetching readings between %s and %s from table %s", startDate, endDate, tableName)
	readings, err := tempService.ServiceGetBetweenDates(tableName, startDate, endDate)

	if err != nil {
		log.Printf("Error fetching readings: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, readings)
}

func currentTemperaturesWebSocket(c *gin.Context) {
	currentTemperatures, err := tempService.ServiceGetLatest()

	if err != nil {
		log.Printf("Error fetching latest temperatures: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching latest temperatures"})
		return
	}

	createPushWebSocket(c, "current-temperatures")

	ws.BroadcastToTopic("current-temperatures", currentTemperatures)
}

func RegisterTemperatureRoutes(router *gin.Engine) {
	temperatureGroup := router.Group("/temperature")
	{
		temperatureGroup.GET("/readings/between", getReadingsBetweenDatesHandler)
		temperatureGroup.GET("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
		temperatureGroup.GET("/ws/current-temperatures", currentTemperaturesWebSocket)
	}
}
