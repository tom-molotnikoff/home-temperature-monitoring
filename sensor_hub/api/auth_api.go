package api

import (
	appProps "example/sensorHub/application_properties"
	gen "example/sensorHub/gen"
	"example/sensorHub/service"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)



type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) loginHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req loginRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	token, csrf, mustChange, err := s.authService.Login(ctx, req.Username, req.Password, ip, userAgent)
	if err != nil {
		switch e := err.(type) {
		case *service.TooManyAttemptsError:
			slog.Warn("rejecting login: too many attempts",
				"ip", c.ClientIP(),
				"username", req.Username,
				"retry_after", e.RetryAfterSeconds,
				"failed_by_user", e.FailedByUser,
				"failed_by_ip", e.FailedByIP,
				"threshold", e.Threshold,
				"exponent", e.Exponent,
			)
			c.Header("Retry-After", fmt.Sprintf("%d", e.RetryAfterSeconds))
			c.IndentedJSON(http.StatusTooManyRequests, gin.H{"message": "too many failed login attempts, retry later", "retry_after": e.RetryAfterSeconds, "failed_by_user": e.FailedByUser, "failed_by_ip": e.FailedByIP, "threshold": e.Threshold, "exponent": e.Exponent})
			return
		default:
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
			return
		}
	}
	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	secure := false
	if os.Getenv("SENSOR_HUB_PRODUCTION") == "true" || c.Request.TLS != nil {
		secure = true
	}

	ttlMinutes := 60 * 24 * 30
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionTTLMinutes > 0 {
		ttlMinutes = appProps.AppConfig.AuthSessionTTLMinutes
	}
	expires := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	// return csrf token in JSON (SPA should store it in memory and send via X-CSRF-Token header on state changes)
	c.IndentedJSON(http.StatusOK, gin.H{"must_change_password": mustChange, "csrf_token": csrf})
}

func (s *Server) logoutHandler(c *gin.Context) {
	ctx := c.Request.Context()
	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	token, err := c.Cookie(cookieName)
	secure := false
	if os.Getenv("SENSOR_HUB_PRODUCTION") == "true" || c.Request.TLS != nil {
		secure = true
	}
	if err == nil && token != "" {
		_ = s.authService.Logout(ctx, token)
		http.SetCookie(c.Writer, &http.Cookie{Name: cookieName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
	}
	c.Status(http.StatusOK)
}

func (s *Server) meHandler(c *gin.Context) {
	ctx := c.Request.Context()
	u, exists := c.Get("currentUser")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	user, ok := u.(*gen.User)
	if !ok || user == nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	token, _ := c.Cookie(cookieName)
	csrf := ""
	if token != "" {
		if t, err := s.authService.GetCSRFForToken(ctx, token); err == nil {
			csrf = t
		}
	}
	c.IndentedJSON(http.StatusOK, gin.H{"user": user, "csrf_token": csrf})
}

func (s *Server) listSessionsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	u, _ := c.Get("currentUser")
	user := u.(*gen.User)
	sessions, err := s.authService.ListSessionsForUser(ctx, user.Id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list sessions", "error": err.Error()})
		return
	}
	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	currentToken, _ := c.Cookie(cookieName)
	var currentSessionId int64
	if currentToken != "" {
		if sid, err := s.authService.GetSessionIdForToken(ctx, currentToken); err == nil {
			currentSessionId = sid
		}
	}

	out := make([]gin.H, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, gin.H{
			"id":               s.Id,
			"user_id":          s.UserId,
			"created_at":       s.CreatedAt,
			"expires_at":       s.ExpiresAt,
			"last_accessed_at": s.LastAccessedAt,
			"ip_address":       s.IpAddress,
			"user_agent":       s.UserAgent,
			"current":          s.Id == currentSessionId,
		})
	}
	c.IndentedJSON(http.StatusOK, out)
}

func (s *Server) revokeSessionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	if idStr == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "session id required"})
		return
	}

	var id int64
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid session id"})
		return
	}

	u, _ := c.Get("currentUser")
	user := u.(*gen.User)

	sessions, err := s.authService.ListSessionsForUser(ctx, user.Id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list sessions", "error": err.Error()})
		return
	}
	owned := false
	for _, s := range sessions {
		if s.Id == id {
			owned = true
			break
		}
	}
	isAdmin := false
	for _, r := range user.Roles {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	if !owned && !isAdmin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	revokerId := user.Id
	if err := s.authService.RevokeSessionByIdWithActor(ctx, id, &revokerId, nil); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to revoke session", "error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
