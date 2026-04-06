package api

import (
	"example/sensorHub/service"
	"example/sensorHub/ws"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var readingsService service.ReadingsServiceInterface

func InitReadingsAPI(s service.ReadingsServiceInterface) {
	readingsService = s
}

func getHourlyReadingsBetweenDatesHandler(c *gin.Context) {
	getReadingsBetweenDatesHelper(c, true)
}

func getReadingsBetweenDatesHandler(c *gin.Context) {
	getReadingsBetweenDatesHelper(c, false)
}

func getReadingsBetweenDatesHelper(c *gin.Context, hourly bool) {
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
	measurementType := c.Query("type")

	slog.Debug("fetching readings between dates", "start", startDate, "end", endDate, "sensor", sensorName, "type", measurementType, "hourly", hourly)
	readings, err := readingsService.ServiceGetBetweenDates(ctx, startDate, endDate, sensorName, measurementType, hourly)

	if err != nil {
		slog.Error("error fetching readings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, readings)
}

func currentReadingsWebSocket(c *gin.Context) {
	ctx := c.Request.Context()
	currentReadings, err := readingsService.ServiceGetLatest(ctx)

	if err != nil {
		slog.Error("error fetching latest readings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching latest readings"})
		return
	}

	createPushWebSocket(c, "current-readings")

	ws.BroadcastToTopic("current-readings", currentReadings)
}
