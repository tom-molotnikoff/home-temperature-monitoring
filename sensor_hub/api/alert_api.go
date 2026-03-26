package api

import (
	"example/sensorHub/alerting"
	"example/sensorHub/service"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var alertManagementService service.AlertManagementServiceInterface

func InitAlertAPI(s service.AlertManagementServiceInterface) {
	alertManagementService = s
}

func getAllAlertRulesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	rules, err := alertManagementService.ServiceGetAllAlertRules(ctx)
	if err != nil {
		slog.Error("error fetching alert rules", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert rules", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, rules)
}

func getAlertRuleBySensorIDHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorIDStr := c.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	rule, err := alertManagementService.ServiceGetAlertRuleBySensorID(ctx, sensorID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Alert rule not found", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, rule)
}

func createAlertRuleHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var rule alerting.AlertRule
	if err := c.BindJSON(&rule); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Validate the rule
	if err := rule.Validate(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceCreateAlertRule(ctx, &rule); err != nil {
		slog.Error("error creating alert rule", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating alert rule", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Alert rule created successfully"})
}

func updateAlertRuleHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorIDStr := c.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	var rule alerting.AlertRule
	if err := c.BindJSON(&rule); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Ensure the sensor ID from URL matches the rule
	rule.SensorID = sensorID

	// Validate the rule
	if err := rule.Validate(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceUpdateAlertRule(ctx, &rule); err != nil {
		slog.Error("error updating alert rule", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating alert rule", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule updated successfully"})
}

func deleteAlertRuleHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorIDStr := c.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceDeleteAlertRule(ctx, sensorID); err != nil {
		slog.Error("error deleting alert rule", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting alert rule", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule deleted successfully"})
}

func getAlertHistoryHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorIDStr := c.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	// Default limit is 50, max 100
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	history, err := alertManagementService.ServiceGetAlertHistory(ctx, sensorID, limit)
	if err != nil {
		slog.Error("error fetching alert history", "sensor_id", sensorID, "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert history", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, history)
}
