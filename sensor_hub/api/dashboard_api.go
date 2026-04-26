package api

import (
	"net/http"
	"strconv"

	"example/sensorHub/service"
	"example/sensorHub/types"
	gen "example/sensorHub/gen"

	"github.com/gin-gonic/gin"
)

var dashboardService service.DashboardServiceInterface

func InitDashboardAPI(s service.DashboardServiceInterface) {
	dashboardService = s
}

func listDashboardsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	dashboards, err := dashboardService.ServiceListDashboards(ctx, user.Id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error listing dashboards"})
		return
	}
	if dashboards == nil {
		dashboards = []gen.Dashboard{}
	}
	c.IndentedJSON(http.StatusOK, dashboards)
}

func getDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	dashboard, err := dashboardService.ServiceGetDashboard(ctx, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error getting dashboard"})
		return
	}
	if dashboard == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Dashboard not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, dashboard)
}

func createDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	var req types.CreateDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	id, err := dashboardService.ServiceCreateDashboard(ctx, user.Id, req)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating dashboard"})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"id": id})
}

func updateDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	var req types.UpdateDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := dashboardService.ServiceUpdateDashboard(ctx, user.Id, id, req); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard updated"})
}

func deleteDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	if err := dashboardService.ServiceDeleteDashboard(ctx, user.Id, id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard deleted"})
}

func shareDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	var req types.ShareDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := dashboardService.ServiceShareDashboard(ctx, user.Id, id, req.TargetUserId); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard shared"})
}

func setDefaultDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	if err := dashboardService.ServiceSetDefaultDashboard(ctx, user.Id, id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Default dashboard set"})
}
