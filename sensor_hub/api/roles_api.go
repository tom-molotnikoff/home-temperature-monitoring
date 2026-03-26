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

func listRolesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	roles, err := roleService.ListRoles(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list roles", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, roles)
}

func listPermissionsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	perms, err := roleService.ListPermissions(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list permissions", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, perms)
}

func getRolePermissionsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var roleId int
	_, err := fmt.Sscan(id, &roleId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
		return
	}
	perms, err := roleService.ListPermissionsForRole(ctx, roleId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list role permissions", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, perms)
}

func assignPermissionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var roleId int
	_, err := fmt.Sscan(id, &roleId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
		return
	}
	var req struct {
		PermissionId int `json:"permission_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	if err := roleService.AssignPermission(ctx, roleId, req.PermissionId); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to assign permission", "error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func removePermissionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	pid := c.Param("pid")
	var roleId, permissionId int
	_, err := fmt.Sscan(id, &roleId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
		return
	}
	_, err = fmt.Sscan(pid, &permissionId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid permission id"})
		return
	}
	if err := roleService.RemovePermission(ctx, roleId, permissionId); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to remove permission", "error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
