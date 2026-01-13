package api

import (
	"example/sensorHub/alerting"
	"example/sensorHub/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var alertManagementService service.AlertManagementServiceInterface

func InitAlertAPI(s service.AlertManagementServiceInterface) {
	alertManagementService = s
}

func getAllAlertRulesHandler(ctx *gin.Context) {
	rules, err := alertManagementService.ServiceGetAllAlertRules()
	if err != nil {
		log.Printf("Error fetching alert rules: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert rules", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, rules)
}

func getAlertRuleBySensorIDHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	rule, err := alertManagementService.ServiceGetAlertRuleBySensorID(sensorID)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "Alert rule not found", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, rule)
}

func createAlertRuleHandler(ctx *gin.Context) {
	var rule alerting.AlertRule
	if err := ctx.BindJSON(&rule); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Validate the rule
	if err := rule.Validate(); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceCreateAlertRule(&rule); err != nil {
		log.Printf("Error creating alert rule: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating alert rule", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusCreated, gin.H{"message": "Alert rule created successfully"})
}

func updateAlertRuleHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	var rule alerting.AlertRule
	if err := ctx.BindJSON(&rule); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Ensure the sensor ID from URL matches the rule
	rule.SensorID = sensorID

	// Validate the rule
	if err := rule.Validate(); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceUpdateAlertRule(&rule); err != nil {
		log.Printf("Error updating alert rule: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating alert rule", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule updated successfully"})
}

func deleteAlertRuleHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceDeleteAlertRule(sensorID); err != nil {
		log.Printf("Error deleting alert rule: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting alert rule", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule deleted successfully"})
}

func getAlertHistoryHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	// Default limit is 50, max 100
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	history, err := alertManagementService.ServiceGetAlertHistory(sensorID, limit)
	if err != nil {
		log.Printf("Error fetching alert history: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert history", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, history)
}
