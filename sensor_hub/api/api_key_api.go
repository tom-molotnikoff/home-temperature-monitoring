package api

import (
	gen "example/sensorHub/gen"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)



type createApiKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type updateExpiryRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
}

func (s *Server) createApiKeyHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req createApiKeyRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	user := c.MustGet("currentUser").(*gen.User)

	fullKey, err := s.apiKeyService.CreateApiKey(ctx, req.Name, user.Id, req.ExpiresAt)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to create api key", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{
		"key":     fullKey,
		"message": "Store this key securely. It will not be shown again.",
	})
}

func (s *Server) listApiKeysHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	keys, err := s.apiKeyService.ListApiKeysForUser(ctx, user.Id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list api keys", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, keys)
}

func (s *Server) updateApiKeyExpiryHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
		return
	}

	var req updateExpiryRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	if err := s.apiKeyService.UpdateApiKeyExpiry(ctx, id, user.Id, req.ExpiresAt); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to update expiry", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "expiry updated"})
}

func (s *Server) revokeApiKeyHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
		return
	}

	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	if err := s.apiKeyService.RevokeApiKey(ctx, id, user.Id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to revoke api key", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "api key revoked"})
}

func (s *Server) deleteApiKeyHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
		return
	}

	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*gen.User)

	if err := s.apiKeyService.DeleteApiKey(ctx, id, user.Id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete api key", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "api key deleted"})
}
