package api

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/service"
	"example/sensorHub/types"
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
	interval := appProps.APPLICATION_PROPERTIES["current.temperature.websocket.interval"]
	if interval == "" {
		interval = "5" // Default to 5 seconds if not set
	}
	intervalDuration, err := time.ParseDuration(interval + "s")
	if err != nil {
		log.Printf("Invalid interval duration: %v, using default 5 seconds", err)
		intervalDuration = 5 * time.Second // Default to 5 seconds
	}

	getter := func() (any, error) {
		return tempService.ServiceGetLatest()
	}
	createWebSocket(c, getter, int(intervalDuration.Seconds()))
}

func RegisterTemperatureRoutes(router *gin.Engine) {
	temperatureGroup := router.Group("/temperature")
	{
		temperatureGroup.GET("/readings/between", getReadingsBetweenDatesHandler)
		temperatureGroup.GET("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
		temperatureGroup.GET("/ws/current-temperatures", currentTemperaturesWebSocket)
	}
}
