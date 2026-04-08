package api

import (
	"example/sensorHub/drivers"
	"net/http"

	"github.com/gin-gonic/gin"
)

type driverInfoResponse struct {
	Type                      string                    `json:"type"`
	DisplayName               string                    `json:"display_name"`
	Description               string                    `json:"description"`
	SupportedMeasurementTypes []string                  `json:"supported_measurement_types"`
	ConfigFields              []drivers.ConfigFieldSpec `json:"config_fields"`
}

func listDriversHandler(c *gin.Context) {
	allDrivers := drivers.All()
	result := make([]driverInfoResponse, 0, len(allDrivers))
	for _, d := range allDrivers {
		mtNames := make([]string, 0)
		for _, mt := range d.SupportedMeasurementTypes() {
			mtNames = append(mtNames, mt.Name)
		}
		result = append(result, driverInfoResponse{
			Type:                      d.Type(),
			DisplayName:               d.DisplayName(),
			Description:               d.Description(),
			SupportedMeasurementTypes: mtNames,
			ConfigFields:              d.ConfigFields(),
		})
	}
	c.IndentedJSON(http.StatusOK, result)
}

func RegisterDriverRoutes(router gin.IRouter) {
	driversGroup := router.Group("/drivers")
	{
		driversGroup.GET("", listDriversHandler)
	}
}
