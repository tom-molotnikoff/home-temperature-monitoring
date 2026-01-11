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

func updatePropertiesHandler(ctx *gin.Context) {
	var requestBody map[string]string

	if err := ctx.BindJSON(&requestBody); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	err := propertiesService.ServiceUpdateProperties(requestBody)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating properties", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusAccepted, gin.H{"message": "Property updated successfully"})
}

func getPropertiesHandler(ctx *gin.Context) {
	properties, err := propertiesService.ServiceGetProperties()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching properties", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, properties)
}

func propertiesWebSocketHandler(ctx *gin.Context) {
	properties, err := propertiesService.ServiceGetProperties()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching properties", "error": err.Error()})
		return
	}

	createPushWebSocket(ctx, "properties")

	ws.BroadcastToTopic("properties", properties)
}

func RegisterPropertiesRoutes(router *gin.Engine) {
	propertiesGroup := router.Group("/properties")
	{
		propertiesGroup.PATCH("/", middleware.AuthRequired(), middleware.RequirePermission("manage_properties"), updatePropertiesHandler)
		propertiesGroup.GET("/", middleware.AuthRequired(), middleware.RequirePermission("view_properties"), getPropertiesHandler)
		propertiesGroup.GET("/ws", middleware.AuthRequired(), middleware.RequirePermission("view_properties"), propertiesWebSocketHandler)
	}
}
