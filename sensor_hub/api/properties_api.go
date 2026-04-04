package api

import (
	"example/sensorHub/api/middleware"
	"example/sensorHub/service"
	"example/sensorHub/ws"
	"net/http"

	"github.com/gin-gonic/gin"
)

var propertiesService service.PropertiesServiceInterface

func InitPropertiesAPI(s service.PropertiesServiceInterface) {
	propertiesService = s
}

func updatePropertiesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var requestBody map[string]string

	if err := c.BindJSON(&requestBody); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	err := propertiesService.ServiceUpdateProperties(ctx, requestBody)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating properties", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusAccepted, gin.H{"message": "Property updated successfully"})
}

func getPropertiesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	properties, err := propertiesService.ServiceGetProperties(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching properties", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, properties)
}

func propertiesWebSocketHandler(c *gin.Context) {
	ctx := c.Request.Context()
	properties, err := propertiesService.ServiceGetProperties(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching properties", "error": err.Error()})
		return
	}

	createPushWebSocket(c, "properties")

	ws.BroadcastToTopic("properties", properties)
}

func RegisterPropertiesRoutes(router gin.IRouter) {
	propertiesGroup := router.Group("/properties")
	{
		propertiesGroup.PATCH("", middleware.AuthRequired(), middleware.RequirePermission("manage_properties"), updatePropertiesHandler)
		propertiesGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_properties"), getPropertiesHandler)
		propertiesGroup.GET("/ws", middleware.AuthRequired(), middleware.RequirePermission("view_properties"), propertiesWebSocketHandler)
	}
}
