# API Test Coverage Expansion Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Expand API test coverage from ~50% to >85% by adding error cases, edge cases, authorization checks, and input validation tests.

**Architecture:** Add comprehensive negative tests and edge cases to existing test files. Focus on: invalid JSON, missing fields, authorization failures, service errors, boundary conditions, and response body validation.

**Tech Stack:** Go, Gin, Testify

---

## Current Gaps Analysis

**Auth API (60-81% coverage):**
- Missing: Invalid JSON body, missing required fields
- Missing: Logout error cases (missing cookie, service errors)
- Missing: MeHandler error cases (missing cookie, CSRF fetch failure)
- Missing: RevokeSession edge cases (invalid ID format, not owned session, admin revoking others)
- Missing: ListSessions error case

**User API (50-61% coverage):**
- Missing: Invalid JSON for all handlers
- Missing: Authorization failures (non-admin operations)
- Missing: CreateUser error cases (service failure, invalid data)
- Missing: ChangePassword edge cases (UserId=0 behavior, admin changing others, non-admin forbidden)
- Missing: DeleteUser edge cases (invalid ID format, self-deletion prevention, non-admin forbidden)
- Missing: SetMustChange/SetRoles error cases (invalid ID, invalid JSON, non-admin forbidden)

**Roles API (57-63% coverage):**
- Missing: Service error cases for all handlers
- Missing: Invalid ID format tests
- Missing: Invalid JSON tests

**Sensor API (50-60% coverage):**
- Missing: Invalid JSON for all handlers
- Missing: Service error cases
- Missing: Invalid ID/name format tests
- Missing: Missing required fields tests
- Missing: SensorExists error case, NotFound case

**Properties API (55-60% coverage):**
- Missing: Invalid JSON test
- Missing: Service error cases

---

### Task 1: Expand Auth API Tests

**Files:**
- Modify: `sensor_hub/api/auth_api_test.go`

**Step 1: Add login error cases**

Add tests for:
- Invalid JSON body
- Missing username field
- Missing password field
- MustChangePassword=true response

```go
func TestLoginHandler_InvalidJSON(t *testing.T) {
	router, _ := setupAuthRouter()
	router.POST("/auth/login", loginHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString("invalid-json"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestLoginHandler_MustChangePassword(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/login", loginHandler)

	reqBody := loginRequest{Username: "user", Password: "password"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("Login", "user", "password", mock.Anything, mock.Anything).Return("token", "csrf", true, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "must_change_password")
	assert.Contains(t, w.Body.String(), "true")
}
```

**Step 2: Add logout error cases**

```go
func TestLogoutHandler_MissingCookie(t *testing.T) {
	router, _ := setupAuthRouter()
	router.POST("/auth/logout", logoutHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogoutHandler_ServiceError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.POST("/auth/logout", logoutHandler)

	mockService.On("Logout", "valid-token").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 3: Add meHandler error cases**

```go
func TestMeHandler_MissingCookie(t *testing.T) {
	router, _ := setupAuthRouter()
	router.GET("/auth/me", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "me"})
		meHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMeHandler_CSRFError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.GET("/auth/me", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Username: "me"})
		meHandler(c)
	})

	mockService.On("GetCSRFForToken", "valid-token").Return("", errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 4: Add listSessions error case**

```go
func TestListSessionsHandler_ServiceError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.GET("/auth/sessions", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		listSessionsHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/sessions", nil)
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 5: Add revokeSession edge cases**

```go
func TestRevokeSessionHandler_InvalidID(t *testing.T) {
	router, _ := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeSessionHandler_MissingID(t *testing.T) {
	router, _ := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeSessionHandler_NotOwnedSession(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100, UserId: 1}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/200", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRevokeSessionHandler_AdminRevokingOthers(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100, UserId: 1}}, nil)
	revoker := 1
	mockService.On("RevokeSessionByIdWithActor", int64(200), &revoker, (*string)(nil)).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/200", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRevokeSessionHandler_ListError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/100", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRevokeSessionHandler_RevokeError(t *testing.T) {
	router, mockService := setupAuthRouter()
	router.DELETE("/auth/sessions/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		revokeSessionHandler(c)
	})

	mockService.On("ListSessionsForUser", 1).Return([]db.SessionInfo{{Id: 100, UserId: 1}}, nil)
	revoker := 1
	mockService.On("RevokeSessionByIdWithActor", int64(100), &revoker, (*string)(nil)).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/auth/sessions/100", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 6: Run tests**

Run: `cd sensor_hub && go test ./api -run TestLogin -v`
Expected: All new tests PASS

**Step 7: Run tests**

Run: `cd sensor_hub && go test ./api -run TestLogout -v`
Expected: All new tests PASS

**Step 8: Run tests**

Run: `cd sensor_hub && go test ./api -run TestMe -v`
Expected: All new tests PASS

**Step 9: Run tests**

Run: `cd sensor_hub && go test ./api -run TestListSessions -v`
Expected: All new tests PASS

**Step 10: Run tests**

Run: `cd sensor_hub && go test ./api -run TestRevokeSession -v`
Expected: All new tests PASS

---

### Task 2: Expand User API Tests

**Files:**
- Modify: `sensor_hub/api/users_api_test.go`

**Step 1: Add createUser error cases**

```go
func TestCreateUserHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.POST("/users", createUserHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUserHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.POST("/users", createUserHandler)

	reqBody := createUserRequest{Username: "newuser", Password: "password"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("CreateUser", mock.AnythingOfType("types.User"), "password").Return(0, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 2: Add listUsers error case**

```go
func TestListUsersHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.GET("/users", listUsersHandler)

	mockService.On("ListUsers").Return([]types.User{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 3: Add changePassword edge cases**

```go
func TestChangePasswordHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		changePasswordHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangePasswordHandler_DefaultsToCurrentUser(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 0, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", 1, "newpass", "valid-token").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChangePasswordHandler_AdminChangingOthers(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 2, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", 2, "newpass", "").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChangePasswordHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 2, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestChangePasswordHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1})
		changePasswordHandler(c)
	})

	reqBody := changePasswordRequest{UserId: 1, NewPassword: "newpass"}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", 1, "newpass", "valid-token").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/password", bytes.NewBuffer(jsonBody))
	req.AddCookie(&http.Cookie{Name: "sensor_hub_session", Value: "valid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 4: Add deleteUser edge cases**

```go
func TestDeleteUserHandler_InvalidID(t *testing.T) {
	router, _ := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		deleteUserHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		deleteUserHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteUserHandler_CannotDeleteSelf(t *testing.T) {
	router, _ := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		deleteUserHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		deleteUserHandler(c)
	})

	mockService.On("DeleteUser", 2).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 5: Add setMustChange edge cases**

```go
func TestSetMustChangeHandler_InvalidID(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setMustChangeHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/invalid/must-change-password", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetMustChangeHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setMustChangeHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/must-change-password", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetMustChangeHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		setMustChangeHandler(c)
	})

	reqBody := mustChangeRequest{MustChange: true}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/must-change-password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSetMustChangeHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/:id/must-change-password", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setMustChangeHandler(c)
	})

	reqBody := mustChangeRequest{MustChange: true}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("SetMustChangeFlag", 2, true).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/must-change-password", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 6: Add setRoles edge cases**

```go
func TestSetRolesHandler_InvalidID(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setRolesHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/invalid/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetRolesHandler_InvalidJSON(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setRolesHandler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/roles", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetRolesHandler_NonAdminForbidden(t *testing.T) {
	router, _ := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"user"}})
		setRolesHandler(c)
	})

	reqBody := setRolesRequest{Roles: []string{"admin"}}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/roles", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSetRolesHandler_ServiceError(t *testing.T) {
	router, mockService := setupUserRouter()
	router.PUT("/users/:id/roles", func(c *gin.Context) {
		c.Set("currentUser", &types.User{Id: 1, Roles: []string{"admin"}})
		setRolesHandler(c)
	})

	reqBody := setRolesRequest{Roles: []string{"admin"}}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("SetUserRoles", 2, []string{"admin"}).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/2/roles", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 7: Run all user tests**

Run: `cd sensor_hub && go test ./api -run TestCreateUser -v`
Expected: All tests PASS

**Step 8: Run all user tests**

Run: `cd sensor_hub && go test ./api -run TestListUsers -v`
Expected: All tests PASS

**Step 9: Run all user tests**

Run: `cd sensor_hub && go test ./api -run TestChangePassword -v`
Expected: All tests PASS

**Step 10: Run all user tests**

Run: `cd sensor_hub && go test ./api -run TestDeleteUser -v`
Expected: All tests PASS

**Step 11: Run all user tests**

Run: `cd sensor_hub && go test ./api -run TestSetMustChange -v`
Expected: All tests PASS

**Step 12: Run all user tests**

Run: `cd sensor_hub && go test ./api -run TestSetRoles -v`
Expected: All tests PASS

---

### Task 3: Expand Roles API Tests

**Files:**
- Modify: `sensor_hub/api/roles_api_test.go`

**Step 1: Add error cases for all role handlers**

```go
func TestListRolesHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/roles", listRolesHandler)

	mockService.On("ListRoles").Return([]db.RoleInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListPermissionsHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/permissions", listPermissionsHandler)

	mockService.On("ListPermissions").Return([]db.PermissionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetRolePermissionsHandler_InvalidID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.GET("/roles/:id/permissions", getRolePermissionsHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles/invalid/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetRolePermissionsHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.GET("/roles/:id/permissions", getRolePermissionsHandler)

	mockService.On("ListPermissionsForRole", 1).Return([]db.PermissionInfo{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/roles/1/permissions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssignPermissionHandler_InvalidID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.POST("/roles/:id/permissions", assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/roles/invalid/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignPermissionHandler_InvalidJSON(t *testing.T) {
	router, _ := setupRoleRouter()
	router.POST("/roles/:id/permissions", assignPermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/roles/1/permissions", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignPermissionHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.POST("/roles/:id/permissions", assignPermissionHandler)

	reqBody := struct {
		PermissionId int `json:"permission_id"`
	}{PermissionId: 10}
	jsonBody, _ := json.Marshal(reqBody)

	mockService.On("AssignPermission", 1, 10).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/roles/1/permissions", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRemovePermissionHandler_InvalidRoleID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.DELETE("/roles/:id/permissions/:pid", removePermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/roles/invalid/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemovePermissionHandler_InvalidPermissionID(t *testing.T) {
	router, _ := setupRoleRouter()
	router.DELETE("/roles/:id/permissions/:pid", removePermissionHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/roles/1/permissions/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemovePermissionHandler_ServiceError(t *testing.T) {
	router, mockService := setupRoleRouter()
	router.DELETE("/roles/:id/permissions/:pid", removePermissionHandler)

	mockService.On("RemovePermission", 1, 10).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/roles/1/permissions/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 2: Run all role tests**

Run: `cd sensor_hub && go test ./api -run "TestList.*Handler" -v`
Expected: All tests PASS

**Step 3: Run all role tests**

Run: `cd sensor_hub && go test ./api -run "Test.*Permission.*Handler" -v`
Expected: All tests PASS

---

### Task 4: Expand Sensor API Tests

**Files:**
- Modify: `sensor_hub/api/sensor_api_test.go`

**Step 1: Add error cases for sensor CRUD**

```go
func TestAddSensorHandler_InvalidJSON(t *testing.T) {
	router, _ := setupSensorRouter()
	router.POST("/sensors", addSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors", addSensorHandler)

	sensor := types.Sensor{Name: "test-sensor", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)

	mockService.On("ServiceAddSensor", sensor).Return(errors.New("validation error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorByNameHandler_NotFound(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", "notfound").Return((*types.Sensor)(nil), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSensorByNameHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", "s1").Return((*types.Sensor)(nil), errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateSensorHandler_InvalidID(t *testing.T) {
	router, _ := setupSensorRouter()
	router.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/sensors/invalid", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_InvalidJSON(t *testing.T) {
	router, _ := setupSensorRouter()
	router.PUT("/sensors/:id", updateSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/sensors/1", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)
	
	expectedSensor := sensor
	expectedSensor.Id = 1

	mockService.On("ServiceUpdateSensorById", expectedSensor).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/sensors/1", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.DELETE("/sensors/:name", deleteSensorHandler)

	mockService.On("ServiceDeleteSensorByName", "s1").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 2: Add error cases for sensor operations**

```go
func TestGetAllSensorsHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors", getAllSensorsHandler)

	mockService.On("ServiceGetAllSensors").Return([]types.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorsByTypeHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/type/:type", getSensorsByTypeHandler)

	mockService.On("ServiceGetSensorsByType", "Temperature").Return([]types.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/type/Temperature", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSensorExistsHandler_NotFound(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", "notfound").Return(false, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSensorExistsHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", "s1").Return(false, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectAndStoreAllSensorReadingsHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/collect", collectAndStoreAllSensorReadingsHandler)

	mockService.On("ServiceCollectAndStoreAllSensorReadings").Return(errors.New("collection error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectFromSensorByNameHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/collect", collectFromSensorByNameHandler)

	mockService.On("ServiceCollectFromSensorByName", "s1").Return(errors.New("sensor offline"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEnableSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/enable", enableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", "s1", true).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDisableSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/disable", disableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", "s1", false).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTotalReadingsPerSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/readings/total", totalReadingsPerSensorHandler)

	mockService.On("ServiceGetTotalReadingsForEachSensor").Return(map[string]int{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/readings/total", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_InvalidLimit(t *testing.T) {
	router, _ := setupSensorRouter()
	router.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1/health?limit=invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	mockService.On("ServiceGetSensorHealthHistoryByName", "s1", 10).Return([]types.SensorHealthHistory{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 3: Run all sensor tests**

Run: `cd sensor_hub && go test ./api -run "Test.*Sensor.*Handler" -v`
Expected: All tests PASS

---

### Task 5: Expand Properties API Tests

**Files:**
- Modify: `sensor_hub/api/properties_api_test.go`

**Step 1: Add error cases**

```go
func TestUpdatePropertiesHandler_InvalidJSON(t *testing.T) {
	router, _ := setupPropertiesRouter()
	router.PATCH("/properties", updatePropertiesHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/properties", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePropertiesHandler_ServiceError(t *testing.T) {
	router, mockService := setupPropertiesRouter()
	router.PATCH("/properties", updatePropertiesHandler)

	props := map[string]string{"key": "value"}
	jsonBody, _ := json.Marshal(props)

	mockService.On("ServiceUpdateProperties", props).Return(errors.New("validation error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/properties", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetPropertiesHandler_ServiceError(t *testing.T) {
	router, mockService := setupPropertiesRouter()
	router.GET("/properties", getPropertiesHandler)

	mockService.On("ServiceGetProperties").Return(map[string]interface{}{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/properties", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
```

**Step 2: Run all properties tests**

Run: `cd sensor_hub && go test ./api -run "Test.*Properties.*Handler" -v`
Expected: All tests PASS

---

### Task 6: Verify Final Coverage

**Files:**
- N/A

**Step 1: Run all tests**

Run: `cd sensor_hub && go test ./api ./api/middleware -v`
Expected: All tests PASS

**Step 2: Check coverage**

Run: `cd sensor_hub && go test -coverprofile=coverage.out ./api ./api/middleware && go tool cover -func=coverage.out`
Expected: 
- Middleware: >90%
- API handlers: >85% (except websocket handlers which are 0%)
- Overall API package: >75%

**Step 3: Generate coverage report**

Run: `cd sensor_hub && go tool cover -html=coverage.out -o coverage.html`

---

## Summary

This plan expands test coverage by adding:
- **Invalid JSON tests** for all handlers that parse request bodies
- **Service error tests** for all service layer calls
- **Authorization tests** for admin-only operations
- **Edge case tests** for ID parsing, missing parameters, boundary conditions
- **Response validation** beyond just status codes

Expected outcome: Increase coverage from ~50% to >85% for all API handlers.
