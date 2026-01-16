package api

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/service"
	"example/sensorHub/types"
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

func createUserHandler(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	user := types.User{Username: req.Username, Email: req.Email, Roles: req.Roles}
	id, err := userService.CreateUser(user, req.Password)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to create user", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusCreated, gin.H{"id": id})
}

func listUsersHandler(ctx *gin.Context) {
	users, err := userService.ListUsers()
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to list users", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, users)
}

type changePasswordRequest struct {
	UserId      int    `json:"user_id"`
	NewPassword string `json:"new_password"`
}

func changePasswordHandler(ctx *gin.Context) {
	var req changePasswordRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	currentUserObj, _ := ctx.Get("currentUser")
	currentUser := currentUserObj.(*types.User)
	if currentUser == nil {
		ctx.Status(http.StatusUnauthorized)
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
			ctx.Status(http.StatusForbidden)
			return
		}
	}

	keepToken := ""
	if currentUser.Id == targetUserId {
		cookieName := "sensor_hub_session"
		if appProps.AppConfig != nil && appProps.AppConfig.AuthSessionCookieName != "" {
			cookieName = appProps.AppConfig.AuthSessionCookieName
		}
		if t, err := ctx.Cookie(cookieName); err == nil {
			keepToken = t
		}
	}
	if err := userService.ChangePassword(targetUserId, req.NewPassword, keepToken); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to change password", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func deleteUserHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user id required"})
		return
	}
	var id int
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}

	currentUserObj, _ := ctx.Get("currentUser")
	currentUser := currentUserObj.(*types.User)
	isAdmin := false
	for _, r := range currentUser.Roles {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		ctx.Status(http.StatusForbidden)
		return
	}

	if currentUser.Id == id {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "cannot delete current user"})
		return
	}
	if err := userService.DeleteUser(id); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete user", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

type mustChangeRequest struct {
	MustChange bool `json:"must_change"`
}

func setMustChangeHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user id required"})
		return
	}
	var id int
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	var req mustChangeRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	currentUserObj, _ := ctx.Get("currentUser")
	currentUser := currentUserObj.(*types.User)
	if currentUser == nil {
		ctx.Status(http.StatusUnauthorized)
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
			ctx.Status(http.StatusForbidden)
			return
		}
	}
	if err := userService.SetMustChangeFlag(id, req.MustChange); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to update user flag", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

type setRolesRequest struct {
	Roles []string `json:"roles"`
}

func setRolesHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user id required"})
		return
	}
	var id int
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	var req setRolesRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	// admin only (middleware ensures currentUser present); we still check for demo
	currentUserObj, _ := ctx.Get("currentUser")
	currentUser := currentUserObj.(*types.User)
	if currentUser == nil {
		ctx.Status(http.StatusUnauthorized)
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
		ctx.Status(http.StatusForbidden)
		return
	}
	if err := userService.SetUserRoles(id, req.Roles); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to set roles", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}
