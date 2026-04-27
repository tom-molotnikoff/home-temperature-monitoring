package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHealth implements gen.ServerInterface.
func (s *Server) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
