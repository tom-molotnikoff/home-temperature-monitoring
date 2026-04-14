package api

import (
	"errors"
	"example/sensorHub/service"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"example/sensorHub/ws"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

var readingsService service.ReadingsServiceInterface

func InitReadingsAPI(s service.ReadingsServiceInterface) {
	readingsService = s
}

func getReadingsBetweenDatesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	startDate := c.Query("start")
	endDate := c.Query("end")
	if startDate == "" || endDate == "" {
		slog.Warn("missing start or end date")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Start and end dates are required"})
		return
	}

	startStr, err := utils.NormalizeDateTimeParam(startDate, false)
	if err != nil {
		slog.Warn("invalid start date format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid start parameter, expected YYYY-MM-DD or ISO 8601 datetime"})
		return
	}

	endStr, err := utils.NormalizeDateTimeParam(endDate, true)
	if err != nil {
		slog.Warn("invalid end date format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid end parameter, expected YYYY-MM-DD or ISO 8601 datetime"})
		return
	}

	sensorName := c.Query("sensor")
	measurementType := c.Query("type")
	overrideInterval := c.Query("aggregation")
	overrideFunction := c.Query("aggregation_function")

	slog.Debug("fetching readings between dates", "start", startStr, "end", endStr, "sensor", sensorName, "type", measurementType, "aggregation", overrideInterval, "aggregation_function", overrideFunction)
	response, err := readingsService.ServiceGetBetweenDates(ctx, startStr, endStr, sensorName, measurementType, overrideInterval, overrideFunction)

	if err != nil {
		var unsupported *types.ErrUnsupportedAggregationFunction
		if errors.As(err, &unsupported) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		slog.Error("error fetching readings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, response)
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
