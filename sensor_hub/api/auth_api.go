package api

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/service"
	"example/sensorHub/types"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var authService service.AuthServiceInterface

func InitAuthAPI(a service.AuthServiceInterface) {
	authService = a
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	ip := ctx.ClientIP()
	userAgent := ctx.Request.UserAgent()
	token, csrf, mustChange, err := authService.Login(req.Username, req.Password, ip, userAgent)
	if err != nil {
		switch e := err.(type) {
		case *service.TooManyAttemptsError:
			log.Printf("rejecting login from ip=%s username=%s: retry_after=%ds failed_by_user=%d failed_by_ip=%d threshold=%d exponent=%d", ctx.ClientIP(), req.Username, e.RetryAfterSeconds, e.FailedByUser, e.FailedByIP, e.Threshold, e.Exponent)
			ctx.Header("Retry-After", fmt.Sprintf("%d", e.RetryAfterSeconds))
			ctx.IndentedJSON(http.StatusTooManyRequests, gin.H{"message": "too many failed login attempts, retry later", "retry_after": e.RetryAfterSeconds, "failed_by_user": e.FailedByUser, "failed_by_ip": e.FailedByIP, "threshold": e.Threshold, "exponent": e.Exponent})
			return
		default:
			ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
			return
		}
	}
	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	secure := false
	if os.Getenv("SENSOR_HUB_PRODUCTION") == "true" || ctx.Request.TLS != nil {
		secure = true
	}

	ttlMinutes := 60 * 24 * 30
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionTTLMinutes > 0 {
		ttlMinutes = appProps.AppConfig.AuthSessionTTLMinutes
	}
	expires := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	// return csrf token in JSON (SPA should store it in memory and send via X-CSRF-Token header on state changes)
	ctx.IndentedJSON(http.StatusOK, gin.H{"must_change_password": mustChange, "csrf_token": csrf})
}

func logoutHandler(ctx *gin.Context) {
	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	token, err := ctx.Cookie(cookieName)
	secure := false
	if os.Getenv("SENSOR_HUB_PRODUCTION") == "true" || ctx.Request.TLS != nil {
		secure = true
	}
	if err == nil && token != "" {
		_ = authService.Logout(token)
		http.SetCookie(ctx.Writer, &http.Cookie{Name: cookieName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
	}
	ctx.Status(http.StatusOK)
}

func meHandler(ctx *gin.Context) {
	u, exists := ctx.Get("currentUser")
	if !exists {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	user, ok := u.(*types.User)
	if !ok || user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	token, _ := ctx.Cookie(cookieName)
	csrf := ""
	if token != "" {
		if t, err := authService.GetCSRFForToken(token); err == nil {
			csrf = t
		}
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"user": user, "csrf_token": csrf})
}

func listSessionsHandler(ctx *gin.Context) {
	u, _ := ctx.Get("currentUser")
	user := u.(*types.User)
	sessions, err := authService.ListSessionsForUser(user.Id)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list sessions", "error": err.Error()})
		return
	}
	cookieName := "sensor_hub_session"
	if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
		cookieName = appProps.AppConfig.AuthSessionCookieName
	}
	currentToken, _ := ctx.Cookie(cookieName)
	var currentSessionId int64
	if currentToken != "" {
		if sid, err := authService.GetSessionIdForToken(currentToken); err == nil {
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
	ctx.IndentedJSON(http.StatusOK, out)
}

func revokeSessionHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "session id required"})
		return
	}

	var id int64
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid session id"})
		return
	}

	u, _ := ctx.Get("currentUser")
	user := u.(*types.User)

	sessions, err := authService.ListSessionsForUser(user.Id)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list sessions", "error": err.Error()})
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
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	revokerId := user.Id
	if err := authService.RevokeSessionByIdWithActor(id, &revokerId, nil); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to revoke session", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}
