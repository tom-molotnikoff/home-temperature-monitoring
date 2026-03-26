package api

import (
	"example/sensorHub/service"
	"example/sensorHub/types"
	"example/sensorHub/ws"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var tempService service.TemperatureServiceInterface

func InitTemperatureAPI(s service.TemperatureServiceInterface) {
	tempService = s
}

func getHourlyReadingsBetweenDatesHandler(c *gin.Context) {
	getReadingsBetweenDatesHelper(c, types.TableHourlyAverageTemperature)
}

func getReadingsBetweenDatesHandler(c *gin.Context) {
	getReadingsBetweenDatesHelper(c, types.TableTemperatureReadings)
}

func getReadingsBetweenDatesHelper(c *gin.Context, tableName string) {
	ctx := c.Request.Context()
	startDate := c.Query("start")
	endDate := c.Query("end")
	if startDate == "" || endDate == "" {
		slog.Warn("missing start or end date")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Start and end dates are required"})
		return
	}

	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		slog.Warn("invalid start date format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid start date format, expected YYYY-MM-DD"})
		return
	}

	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		slog.Warn("invalid end date format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid end date format, expected YYYY-MM-DD"})
		return
	}

	sensorName := c.Query("sensor")

	slog.Debug("fetching readings between dates", "start", startDate, "end", endDate, "sensor", sensorName, "table", tableName)
	readings, err := tempService.ServiceGetBetweenDates(ctx, tableName, startDate, endDate, sensorName)

	if err != nil {
		slog.Error("error fetching readings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, readings)
}

func currentTemperaturesWebSocket(c *gin.Context) {
	ctx := c.Request.Context()
	currentTemperatures, err := tempService.ServiceGetLatest(ctx)

	if err != nil {
		slog.Error("error fetching latest temperatures", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching latest temperatures"})
		return
	}

	createPushWebSocket(c, "current-temperatures")

	ws.BroadcastToTopic("current-temperatures", currentTemperatures)
}
