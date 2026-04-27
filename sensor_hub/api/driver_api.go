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

func (s *Server) listDriversHandler(c *gin.Context) {
	allDrivers := drivers.All()
	typeFilter := c.Query("type") // "pull", "push", or "" (all)

	result := make([]driverInfoResponse, 0, len(allDrivers))
	for _, d := range allDrivers {
		if typeFilter == "pull" {
			if _, ok := d.(drivers.PullDriver); !ok {
				continue
			}
		} else if typeFilter == "push" {
			if _, ok := d.(drivers.PushDriver); !ok {
				continue
			}
		}
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

func (s *Server) RegisterDriverRoutes(router gin.IRouter) {
	driversGroup := router.Group("/drivers")
	{
		driversGroup.GET("", s.listDriversHandler)
	}
}
