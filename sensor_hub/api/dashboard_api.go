package api

import (
	"net/http"

	gen "example/sensorHub/gen"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListDashboards(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	dashboards, err := s.dashboardService.ServiceListDashboards(ctx, user.Id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error listing dashboards"})
		return
	}
	if dashboards == nil {
		dashboards = []gen.Dashboard{}
	}
	c.IndentedJSON(http.StatusOK, dashboards)
}

func (s *Server) GetDashboard(c *gin.Context, id int) {
	ctx := c.Request.Context()

	dashboard, err := s.dashboardService.ServiceGetDashboard(ctx, id)
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

func (s *Server) CreateDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	var req gen.CreateDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	if req.Name == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	id, err := s.dashboardService.ServiceCreateDashboard(ctx, user.Id, req)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating dashboard"})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) UpdateDashboard(c *gin.Context, id int) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	var req gen.UpdateDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := s.dashboardService.ServiceUpdateDashboard(ctx, user.Id, id, req); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard updated"})
}

func (s *Server) DeleteDashboard(c *gin.Context, id int) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	if err := s.dashboardService.ServiceDeleteDashboard(ctx, user.Id, id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard deleted"})
}

func (s *Server) ShareDashboard(c *gin.Context, id int) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	var req gen.ShareDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := s.dashboardService.ServiceShareDashboard(ctx, user.Id, id, req.TargetUserId); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard shared"})
}

func (s *Server) SetDefaultDashboard(c *gin.Context, id int) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	if err := s.dashboardService.ServiceSetDefaultDashboard(ctx, user.Id, id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Default dashboard set"})
}
