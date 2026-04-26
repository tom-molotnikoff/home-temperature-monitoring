package api

import (
	appProps "example/sensorHub/application_properties"
	gen "example/sensorHub/gen"
	"example/sensorHub/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var userService service.UserServiceInterface

func InitUsersAPI(u service.UserServiceInterface) {
	userService = u
}

type createUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func createUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req createUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	user := gen.User{Username: req.Username, Email: req.Email, Roles: req.Roles}
	id, err := userService.CreateUser(ctx, user, req.Password)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to create user", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"id": id})
}

func listUsersHandler(c *gin.Context) {
	ctx := c.Request.Context()
	users, err := userService.ListUsers(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list users", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

type changePasswordRequest struct {
	UserId      int    `json:"user_id"`
	NewPassword string `json:"new_password"`
}

func changePasswordHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req changePasswordRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	currentUserObj, _ := c.Get("currentUser")
	currentUser := currentUserObj.(*gen.User)
	if currentUser == nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	targetUserId := req.UserId
	if targetUserId == 0 {
		targetUserId = currentUser.Id
	}
	if currentUser.Id != targetUserId {

		isAdmin := false
		for _, r := range currentUser.Roles {
			if r == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			c.Status(http.StatusForbidden)
			return
		}
	}

	keepToken := ""
	if currentUser.Id == targetUserId {
		cookieName := "sensor_hub_session"
		if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
			cookieName = appProps.AppConfig.AuthSessionCookieName
		}
		if t, err := c.Cookie(cookieName); err == nil {
			keepToken = t
		}
	}
	if err := userService.ChangePassword(ctx, targetUserId, req.NewPassword, keepToken); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to change password", "error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func deleteUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	if idStr == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user id required"})
		return
	}
	var id int
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}

	currentUserObj, _ := c.Get("currentUser")
	currentUser := currentUserObj.(*gen.User)
	isAdmin := false
	for _, r := range currentUser.Roles {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		c.Status(http.StatusForbidden)
		return
	}

	if currentUser.Id == id {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "cannot delete current user"})
		return
	}
	if err := userService.DeleteUser(ctx, id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete user", "error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

type mustChangeRequest struct {
	MustChange bool `json:"must_change"`
}

func setMustChangeHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	if idStr == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user id required"})
		return
	}
	var id int
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	var req mustChangeRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	currentUserObj, _ := c.Get("currentUser")
	currentUser := currentUserObj.(*gen.User)
	if currentUser == nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	if currentUser.Id != id {

		isAdmin := false
		for _, r := range currentUser.Roles {
			if r == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			c.Status(http.StatusForbidden)
			return
		}
	}
	if err := userService.SetMustChangeFlag(ctx, id, req.MustChange); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to update user flag", "error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

type setRolesRequest struct {
	Roles []string `json:"roles"`
}

func setRolesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	if idStr == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user id required"})
		return
	}
	var id int
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	var req setRolesRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	// admin only (middleware ensures currentUser present); we still check for demo
	currentUserObj, _ := c.Get("currentUser")
	currentUser := currentUserObj.(*gen.User)
	if currentUser == nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	isAdmin := false
	for _, r := range currentUser.Roles {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		c.Status(http.StatusForbidden)
		return
	}
	if err := userService.SetUserRoles(ctx, id, req.Roles); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to set roles", "error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
