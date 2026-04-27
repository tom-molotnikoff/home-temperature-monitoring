package api

import (
	"example/sensorHub/alerting"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)



func (s *Server) getAllAlertRulesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	rules, err := s.alertService.ServiceGetAllAlertRules(ctx)
	if err != nil {
		slog.Error("error fetching alert rules", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert rules", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, rules)
}

func (s *Server) getAlertRuleByIDHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule ID", "error": err.Error()})
		return
	}

	rule, err := s.alertService.ServiceGetAlertRuleByID(ctx, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert rule", "error": err.Error()})
		return
	}
	if rule == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Alert rule not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, rule)
}

func (s *Server) getAlertRulesBySensorIDHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorIDStr := c.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	rules, err := s.alertService.ServiceGetAlertRulesBySensorID(ctx, sensorID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert rules", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, rules)
}

func (s *Server) createAlertRuleHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var rule alerting.AlertRule
	if err := c.BindJSON(&rule); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	if err := rule.Validate(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := s.alertService.ServiceCreateAlertRule(ctx, &rule); err != nil {
		slog.Error("error creating alert rule", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating alert rule", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Alert rule created successfully"})
}

func (s *Server) updateAlertRuleHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule ID", "error": err.Error()})
		return
	}

	var rule alerting.AlertRule
	if err := c.BindJSON(&rule); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	rule.ID = id

	if err := rule.Validate(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := s.alertService.ServiceUpdateAlertRule(ctx, &rule); err != nil {
		slog.Error("error updating alert rule", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating alert rule", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule updated successfully"})
}

func (s *Server) deleteAlertRuleHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule ID", "error": err.Error()})
		return
	}

	if err := s.alertService.ServiceDeleteAlertRule(ctx, id); err != nil {
		slog.Error("error deleting alert rule", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting alert rule", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule deleted successfully"})
}

func (s *Server) getAlertHistoryHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sensorIDStr := c.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	history, err := s.alertService.ServiceGetAlertHistory(ctx, sensorID, limit)
	if err != nil {
		slog.Error("error fetching alert history", "sensor_id", sensorID, "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert history", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, history)
}
