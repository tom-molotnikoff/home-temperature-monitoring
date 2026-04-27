package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// OAuthAPIServiceInterface defines what the API needs from OAuth service
type OAuthAPIServiceInterface interface {
	GetStatus(ctx context.Context) map[string]interface{}
	GetAuthURL(ctx context.Context, state string) (string, error)
	ExchangeCode(ctx context.Context, code string) error
	IsReady(ctx context.Context) bool
	Reload(ctx context.Context) error
}


// pendingStates stores CSRF states for OAuth flow
var pendingStates = struct {
	sync.RWMutex
	states map[string]bool
}{states: make(map[string]bool)}

func (s *Server) oauthStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()
	if s.oauthService == nil {
		c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}
	status := s.oauthService.GetStatus(ctx)
	c.IndentedJSON(http.StatusOK, status)
}

func (s *Server) oauthAuthorizeHandler(c *gin.Context) {
	ctx := c.Request.Context()
	if s.oauthService == nil {
		c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	// Generate CSRF state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate state"})
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Store state for validation
	pendingStates.Lock()
	pendingStates.states[state] = true
	pendingStates.Unlock()

	authURL, err := s.oauthService.GetAuthURL(ctx, state)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get auth URL", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"auth_url": authURL, "state": state})
}

// oauthSubmitCodeRequest is the request body for submitting an authorization code
type oauthSubmitCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// oauthSubmitCodeHandler handles manual submission of the authorization code
// This is used with the out-of-band OAuth flow where Google displays the code on screen
func (s *Server) oauthSubmitCodeHandler(c *gin.Context) {
	ctx := c.Request.Context()
	if s.oauthService == nil {
		c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	var req oauthSubmitCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}

	// Validate state
	pendingStates.Lock()
	valid := pendingStates.states[req.State]
	delete(pendingStates.states, req.State)
	pendingStates.Unlock()

	if !valid {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid or expired state"})
		return
	}

	if err := s.oauthService.ExchangeCode(ctx, req.Code); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to exchange code", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "OAuth authorization successful"})
}

// oauthReloadHandler reloads credentials and token from disk
func (s *Server) oauthReloadHandler(c *gin.Context) {
	ctx := c.Request.Context()
	if s.oauthService == nil {
		c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "OAuth not configured"})
		return
	}

	if err := s.oauthService.Reload(ctx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to reload", "error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "OAuth configuration reloaded"})
}
