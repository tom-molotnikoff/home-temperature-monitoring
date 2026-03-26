package api

import (
	"example/sensorHub/service"
	"example/sensorHub/types"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var apiKeyService service.ApiKeyServiceInterface

func InitApiKeyAPI(a service.ApiKeyServiceInterface) {
	apiKeyService = a
}

type createApiKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type updateExpiryRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
}

func createApiKeyHandler(ctx *gin.Context) {
	var req createApiKeyRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	user := ctx.MustGet("currentUser").(*types.User)

	fullKey, err := apiKeyService.CreateApiKey(req.Name, user.Id, req.ExpiresAt)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to create api key", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusCreated, gin.H{
		"key":     fullKey,
		"message": "Store this key securely. It will not be shown again.",
	})
}

func listApiKeysHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)

	keys, err := apiKeyService.ListApiKeysForUser(user.Id)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list api keys", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, keys)
}

func updateApiKeyExpiryHandler(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
		return
	}

	var req updateExpiryRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	user := ctx.MustGet("currentUser").(*types.User)

	if err := apiKeyService.UpdateApiKeyExpiry(id, user.Id, req.ExpiresAt); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to update expiry", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "expiry updated"})
}

func revokeApiKeyHandler(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
		return
	}

	user := ctx.MustGet("currentUser").(*types.User)

	if err := apiKeyService.RevokeApiKey(id, user.Id); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to revoke api key", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "api key revoked"})
}

func deleteApiKeyHandler(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
		return
	}

	user := ctx.MustGet("currentUser").(*types.User)

	if err := apiKeyService.DeleteApiKey(id, user.Id); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete api key", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "api key deleted"})
}
