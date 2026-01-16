package middleware

import (
	"net/http"
	"strings"

	database "example/sensorHub/db"
	"example/sensorHub/types"

	"github.com/gin-gonic/gin"
)

var roleRepo database.RoleRepository

func InitPermissionMiddleware(r database.RoleRepository) {
	roleRepo = r
}

func RequirePermission(permission string) gin.HandlerFunc {
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

		// If permissions are already populated on the user (from ValidateSession), use them to avoid a DB lookup
		perms := user.Permissions
		if perms == nil {
			var err error
			perms, err = roleRepo.GetPermissionsForUser(user.Id)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}
		for _, p := range perms {
			if strings.EqualFold(p, permission) {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatus(http.StatusForbidden)
	}
}
