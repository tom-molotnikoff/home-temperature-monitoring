package middleware

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/service"
	"example/sensorHub/types"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

var authService service.AuthServiceInterface

func InitAuthMiddleware(a service.AuthServiceInterface) {
	authService = a
}

func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookieName := "sensor_hub_session"
		if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
			cookieName = appProps.AppConfig.AuthSessionCookieName
		}
		token, err := ctx.Cookie(cookieName)
		if err != nil || token == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user, err := authService.ValidateSession(token)
		if err != nil || user == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if user.MustChangePassword {
			allowed := map[string]struct{}{
				"POST:/auth/login":    {},
				"POST:/auth/logout":   {},
				"GET:/auth/me":        {},
				"PUT:/users/password": {},
			}
			method := ctx.Request.Method
			cleanPath := path.Clean(ctx.Request.URL.Path)
			key := method + ":" + cleanPath
			if _, ok := allowed[key]; !ok {
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		ctx.Set("currentUser", user)
		ctx.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		u, exists := ctx.Get("currentUser")
		if !exists {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user := u.(*types.User)
		if user == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		isAdmin := false
		for _, r := range user.Roles {
			if r == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Next()
	}
}
