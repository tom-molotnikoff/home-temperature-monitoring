package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// OAuthAPIServiceInterface defines what the API needs from OAuth service
type OAuthAPIServiceInterface interface {
	GetStatus() map[string]interface{}
	GetAuthURL(state string) (string, error)
	ExchangeCode(code string) error
	IsReady() bool
	Reload() error
}

var oauthAPIService OAuthAPIServiceInterface

// pendingStates stores CSRF states for OAuth flow
var pendingStates = struct {
	sync.RWMutex
	states map[string]bool
}{states: make(map[string]bool)}

func InitOAuthAPI(s OAuthAPIServiceInterface) {
	oauthAPIService = s
}

func oauthStatusHandler(ctx *gin.Context) {
	if oauthAPIService == nil {
		ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}
	status := oauthAPIService.GetStatus()
	ctx.IndentedJSON(http.StatusOK, status)
}

func oauthAuthorizeHandler(ctx *gin.Context) {
	if oauthAPIService == nil {
		ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	// Generate CSRF state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate state"})
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Store state for validation
	pendingStates.Lock()
	pendingStates.states[state] = true
	pendingStates.Unlock()

	authURL, err := oauthAPIService.GetAuthURL(state)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get auth URL", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"auth_url": authURL, "state": state})
}

// oauthSubmitCodeRequest is the request body for submitting an authorization code
type oauthSubmitCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// oauthSubmitCodeHandler handles manual submission of the authorization code
// This is used with the out-of-band OAuth flow where Google displays the code on screen
func oauthSubmitCodeHandler(ctx *gin.Context) {
	if oauthAPIService == nil {
		ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	var req oauthSubmitCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}

	// Validate state
	pendingStates.Lock()
	valid := pendingStates.states[req.State]
	delete(pendingStates.states, req.State)
	pendingStates.Unlock()

	if !valid {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid or expired state"})
		return
	}

	if err := oauthAPIService.ExchangeCode(req.Code); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to exchange code", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "OAuth authorization successful"})
}

// oauthReloadHandler reloads credentials and token from disk
func oauthReloadHandler(ctx *gin.Context) {
	if oauthAPIService == nil {
		ctx.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	if err := oauthAPIService.Reload(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to reload", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "OAuth configuration reloaded"})
}
