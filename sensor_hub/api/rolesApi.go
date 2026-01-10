package api

import (
	"example/sensorHub/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var roleService service.RoleServiceInterface

func InitRolesAPI(rs service.RoleServiceInterface) {
	roleService = rs
}

func listRolesHandler(ctx *gin.Context) {
	roles, err := roleService.ListRoles()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list roles", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, roles)
}

func listPermissionsHandler(ctx *gin.Context) {
	perms, err := roleService.ListPermissions()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list permissions", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, perms)
}

func getRolePermissionsHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var roleId int
	_, err := fmt.Sscan(id, &roleId)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
		return
	}
	perms, err := roleService.ListPermissionsForRole(roleId)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list role permissions", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, perms)
}

func assignPermissionHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var roleId int
	_, err := fmt.Sscan(id, &roleId)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
		return
	}
	var req struct {
		PermissionId int `json:"permission_id"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	if err := roleService.AssignPermission(roleId, req.PermissionId); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to assign permission", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func removePermissionHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	pid := ctx.Param("pid")
	var roleId, permissionId int
	_, err := fmt.Sscan(id, &roleId)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
		return
	}
	_, err = fmt.Sscan(pid, &permissionId)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid permission id"})
		return
	}
	if err := roleService.RemovePermission(roleId, permissionId); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to remove permission", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}
