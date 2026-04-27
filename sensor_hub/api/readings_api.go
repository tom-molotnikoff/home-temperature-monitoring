package api

import (
	"errors"
	"example/sensorHub/service"
	"example/sensorHub/utils"
	"example/sensorHub/ws"
	gen "example/sensorHub/gen"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetReadingsBetweenDates implements gen.ServerInterface.
// Query parameters arrive pre-parsed in params; start/end are still normalised here.
func (s *Server) GetReadingsBetweenDates(c *gin.Context, params gen.GetReadingsBetweenDatesParams) {
	ctx := c.Request.Context()

	if params.Start == "" || params.End == "" {
		slog.Warn("missing start or end date")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Start and end dates are required"})
		return
	}

	startStr, err := utils.NormalizeDateTimeParam(params.Start, false)
	if err != nil {
		slog.Warn("invalid start date format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid start parameter, expected YYYY-MM-DD or ISO 8601 datetime"})
		return
	}

	endStr, err := utils.NormalizeDateTimeParam(params.End, true)
	if err != nil {
		slog.Warn("invalid end date format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid end parameter, expected YYYY-MM-DD or ISO 8601 datetime"})
		return
	}

	var sensorName, measurementType, overrideInterval, overrideFunction string
	if params.Sensor != nil {
		sensorName = *params.Sensor
	}
	if params.Type != nil {
		measurementType = *params.Type
	}
	if params.Aggregation != nil {
		overrideInterval = string(*params.Aggregation)
	}
	if params.AggregationFunction != nil {
		overrideFunction = string(*params.AggregationFunction)
	}

	slog.Debug("fetching readings between dates", "start", startStr, "end", endStr, "sensor", sensorName, "type", measurementType, "aggregation", overrideInterval, "aggregation_function", overrideFunction)
	response, err := s.readingsService.ServiceGetBetweenDates(ctx, startStr, endStr, sensorName, measurementType, overrideInterval, overrideFunction)

	if err != nil {
		var unsupported *service.ErrUnsupportedAggregationFunction
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

func (s *Server) currentReadingsWebSocket(c *gin.Context) {
	ctx := c.Request.Context()
	currentReadings, err := s.readingsService.ServiceGetLatest(ctx)

	if err != nil {
		slog.Error("error fetching latest readings", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching latest readings"})
		return
	}

	createPushWebSocket(c, "current-readings")

	ws.BroadcastToTopic("current-readings", currentReadings)
}
