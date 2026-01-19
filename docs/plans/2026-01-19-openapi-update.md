# OpenAPI Specification Update Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Synchronize `sensor_hub/openapi.yaml` with the current API implementation to accurately document all routes, parameters, request bodies, and responses.

**Architecture:** The API uses Gin router with modular route registration. Each domain (alerts, auth, notifications, etc.) has a `*_routes.go` file defining endpoints and a `*_api.go` file with handlers. Authentication uses session cookies with middleware enforcement and RBAC permissions.

**Tech Stack:** Go (Gin framework), OpenAPI 3.1.0 YAML specification

---

## Summary of Changes Required

### Routes Currently in OpenAPI (already documented):
- `/temperature/readings/hourly/between` (GET)
- `/temperature/readings/between` (GET)
- `/temperature/ws/current-temperatures` (GET - WebSocket)
- `/sensors/ws/{type}` (GET - WebSocket)
- `/sensors` (POST, GET)
- `/sensors/{id}` (PUT)
- `/sensors/{name}` (GET)
- `/sensors/type/{type}` (GET)
- `/sensors/health/{name}` (GET)
- `/sensors/stats/total-readings` (GET)
- `/properties` (GET, PATCH)
- `/properties/ws` (GET - WebSocket)

### Routes Missing from OpenAPI (need to add):

**Alerts Domain (NEW - 6 endpoints):**
- `GET /alerts/` - List all alert rules
- `GET /alerts/:sensorId` - Get alert rule by sensor ID
- `GET /alerts/:sensorId/history` - Get alert history for sensor
- `POST /alerts/` - Create alert rule
- `PUT /alerts/:sensorId` - Update alert rule
- `DELETE /alerts/:sensorId` - Delete alert rule

**Auth Domain (NEW - 5 endpoints):**
- `POST /auth/login` - Login with credentials
- `POST /auth/logout` - Logout current session
- `GET /auth/me` - Get current user info
- `GET /auth/sessions` - List user sessions
- `DELETE /auth/sessions/:id` - Revoke session

**Notifications Domain (NEW - 10 endpoints):**
- `GET /notifications/` - List notifications
- `GET /notifications/unread-count` - Get unread count
- `POST /notifications/:id/read` - Mark as read
- `POST /notifications/:id/dismiss` - Dismiss notification
- `POST /notifications/bulk/read` - Bulk mark as read
- `POST /notifications/bulk/dismiss` - Bulk dismiss
- `GET /notifications/preferences` - Get channel preferences
- `POST /notifications/preferences` - Set channel preference
- `GET /notifications/ws` - WebSocket for real-time notifications

**OAuth Domain (NEW - 4 endpoints):**
- `GET /oauth/status` - Get OAuth status
- `GET /oauth/authorize` - Get authorization URL
- `POST /oauth/submit-code` - Submit authorization code
- `POST /oauth/reload` - Reload OAuth configuration

**Roles Domain (NEW - 5 endpoints):**
- `GET /roles/` - List all roles
- `GET /roles/permissions` - List all permissions
- `GET /roles/:id/permissions` - Get role permissions
- `POST /roles/:id/permissions` - Assign permission to role
- `DELETE /roles/:id/permissions/:pid` - Remove permission from role

**Users Domain (NEW - 6 endpoints):**
- `POST /users/` - Create user
- `GET /users/` - List users
- `PUT /users/password` - Change password
- `DELETE /users/:id` - Delete user
- `PATCH /users/:id/must_change` - Set must change password flag
- `POST /users/:id/roles` - Set user roles

**Sensors Domain (needs updates - 5 new endpoints):**
- `DELETE /sensors/:name` - Delete sensor (intentionally omitted but should be documented with warning)
- `HEAD /sensors/:name` - Check sensor exists
- `POST /sensors/collect` - Trigger collection for all sensors
- `POST /sensors/collect/:sensorName` - Trigger collection for specific sensor
- `POST /sensors/disable/:sensorName` - Disable sensor
- `POST /sensors/enable/:sensorName` - Enable sensor

### Updates to Existing Endpoints:
- Add authentication requirements (`security` section) to all endpoints
- Add permission requirements documentation
- Update response schemas where needed

### New Schema Definitions Required:
- `AlertRule` - Alert rule configuration
- `AlertHistoryEntry` - Alert history record
- `User` - User information
- `SessionInfo` - Session information
- `RoleInfo` - Role information
- `PermissionInfo` - Permission information
- `Notification` - Notification object
- `UserNotification` - User-specific notification
- `ChannelPreference` - Notification channel preferences
- `LoginRequest` / `LoginResponse` - Auth types
- `CreateUserRequest` - User creation request
- `ChangePasswordRequest` - Password change request
- `OAuthStatus` - OAuth status response
- `OAuthAuthorizeResponse` - OAuth authorize response
- `OAuthSubmitCodeRequest` - OAuth code submission

---

## Task List

- [ ] **Task 1:** Add security schemes and global security requirements
- [ ] **Task 2:** Add Auth domain endpoints (`/auth/*`)
- [ ] **Task 3:** Add Users domain endpoints (`/users/*`)
- [ ] **Task 4:** Add Roles domain endpoints (`/roles/*`)
- [ ] **Task 5:** Add Alerts domain endpoints (`/alerts/*`)
- [ ] **Task 6:** Add Notifications domain endpoints (`/notifications/*`)
- [ ] **Task 7:** Add OAuth domain endpoints (`/oauth/*`)
- [ ] **Task 8:** Add missing Sensors endpoints
- [ ] **Task 9:** Update existing endpoints with security requirements
- [ ] **Task 10:** Add all new schema definitions
- [ ] **Task 11:** Validate OpenAPI spec and commit

---

### Task 1: Add Security Schemes and Global Security Requirements

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add security schemes to components section**

Add after `components:` section (before `schemas:`):

```yaml
  securitySchemes:
    cookieAuth:
      type: apiKey
      in: cookie
      name: sensor_hub_session
      description: Session cookie set after successful login
    csrfToken:
      type: apiKey
      in: header
      name: X-CSRF-Token
      description: CSRF token returned from login, required for state-changing operations
```

**Step 2: Add global security requirement at top level**

Add after `servers:` section:

```yaml
security:
  - cookieAuth: []
```

**Step 3: Verify the YAML is still valid**

Run: `cat sensor_hub/openapi.yaml | head -20`
Expected: Valid YAML structure with security schemes

**Step 4: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add authentication security schemes"
```

---

### Task 2: Add Auth Domain Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add /auth/login endpoint**

Add to `paths:` section:

```yaml
  /auth/login:
    post:
      tags:
        - auth
      summary: Authenticate user
      description: >-
        Authenticates a user with username and password. On success, sets a
        session cookie and returns a CSRF token. The CSRF token must be sent
        in the X-CSRF-Token header for all state-changing requests.
      operationId: login
      security: []  # No auth required for login
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: Login successful
          headers:
            Set-Cookie:
              schema:
                type: string
              description: Session cookie
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        '401':
          description: Invalid credentials
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '429':
          description: Too many failed attempts
          headers:
            Retry-After:
              schema:
                type: integer
              description: Seconds until retry is allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RateLimitResponse'
```

**Step 2: Add /auth/logout endpoint**

```yaml
  /auth/logout:
    post:
      tags:
        - auth
      summary: Logout current session
      description: Invalidates the current session and clears the session cookie.
      operationId: logout
      responses:
        '200':
          description: Logout successful
        '401':
          description: Not authenticated
```

**Step 3: Add /auth/me endpoint**

```yaml
  /auth/me:
    get:
      tags:
        - auth
      summary: Get current user info
      description: Returns the currently authenticated user's information and a fresh CSRF token.
      operationId: getCurrentUser
      responses:
        '200':
          description: Current user info
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MeResponse'
        '401':
          description: Not authenticated
```

**Step 4: Add /auth/sessions endpoint**

```yaml
  /auth/sessions:
    get:
      tags:
        - auth
      summary: List user sessions
      description: Returns all active sessions for the current user.
      operationId: listSessions
      responses:
        '200':
          description: List of sessions
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/SessionInfo'
        '401':
          description: Not authenticated
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 5: Add /auth/sessions/{id} endpoint**

```yaml
  /auth/sessions/{id}:
    delete:
      tags:
        - auth
      summary: Revoke a session
      description: >-
        Revokes a specific session by ID. Users can revoke their own sessions.
        Admins can revoke any session.
      operationId: revokeSession
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
          description: Session ID to revoke
      responses:
        '200':
          description: Session revoked
        '400':
          description: Invalid session ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Cannot revoke other user's session (non-admin)
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 6: Add auth tag to tags section**

Add to `components.tags`:

```yaml
    - name: auth
      description: Authentication and session management endpoints
```

**Step 7: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add auth domain endpoints"
```

---

### Task 3: Add Users Domain Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add /users endpoints (POST and GET)**

```yaml
  /users:
    post:
      tags:
        - users
      summary: Create a new user
      description: Creates a new user account. Requires manage_users permission.
      operationId: createUser
      x-required-permission: manage_users
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                    description: ID of the created user
        '400':
          description: Invalid request body
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    get:
      tags:
        - users
      summary: List all users
      description: Returns a list of all users. Requires view_users permission.
      operationId: listUsers
      x-required-permission: view_users
      responses:
        '200':
          description: List of users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 2: Add /users/password endpoint**

```yaml
  /users/password:
    put:
      tags:
        - users
      summary: Change password
      description: >-
        Changes a user's password. Users can change their own password.
        Admins can change any user's password by specifying user_id.
      operationId: changePassword
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ChangePasswordRequest'
      responses:
        '200':
          description: Password changed successfully
        '400':
          description: Invalid request body
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Cannot change other user's password (non-admin)
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 3: Add /users/{id} DELETE endpoint**

```yaml
  /users/{id}:
    delete:
      tags:
        - users
      summary: Delete a user
      description: >-
        Deletes a user by ID. Requires manage_users permission.
        Cannot delete the currently logged-in user.
      operationId: deleteUser
      x-required-permission: manage_users
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: User ID to delete
      responses:
        '200':
          description: User deleted
        '400':
          description: Invalid user ID or attempting to delete self
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 4: Add /users/{id}/must_change endpoint**

```yaml
  /users/{id}/must_change:
    patch:
      tags:
        - users
      summary: Set must change password flag
      description: >-
        Sets whether a user must change their password on next login.
        Requires manage_users permission (or self for own account).
      operationId: setMustChangePassword
      x-required-permission: manage_users
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: User ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                must_change:
                  type: boolean
              required:
                - must_change
      responses:
        '200':
          description: Flag updated
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 5: Add /users/{id}/roles endpoint**

```yaml
  /users/{id}/roles:
    post:
      tags:
        - users
      summary: Set user roles
      description: Sets the roles for a user. Requires manage_users permission (admin only).
      operationId: setUserRoles
      x-required-permission: manage_users
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: User ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                roles:
                  type: array
                  items:
                    type: string
              required:
                - roles
      responses:
        '200':
          description: Roles updated
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions (admin only)
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 6: Add users tag**

```yaml
    - name: users
      description: User management endpoints
```

**Step 7: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add users domain endpoints"
```

---

### Task 4: Add Roles Domain Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add /roles endpoint**

```yaml
  /roles:
    get:
      tags:
        - roles
      summary: List all roles
      description: Returns all available roles. Requires view_roles permission.
      operationId: listRoles
      x-required-permission: view_roles
      responses:
        '200':
          description: List of roles
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/RoleInfo'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 2: Add /roles/permissions endpoint**

```yaml
  /roles/permissions:
    get:
      tags:
        - roles
      summary: List all permissions
      description: Returns all available permissions. Requires view_roles permission.
      operationId: listPermissions
      x-required-permission: view_roles
      responses:
        '200':
          description: List of permissions
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/PermissionInfo'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 3: Add /roles/{id}/permissions endpoints**

```yaml
  /roles/{id}/permissions:
    get:
      tags:
        - roles
      summary: Get permissions for a role
      description: Returns permissions assigned to a specific role. Requires view_roles permission.
      operationId: getRolePermissions
      x-required-permission: view_roles
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: Role ID
      responses:
        '200':
          description: List of permissions
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/PermissionInfo'
        '400':
          description: Invalid role ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    post:
      tags:
        - roles
      summary: Assign permission to role
      description: Assigns a permission to a role. Requires manage_roles permission.
      operationId: assignPermission
      x-required-permission: manage_roles
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: Role ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                permission_id:
                  type: integer
              required:
                - permission_id
      responses:
        '200':
          description: Permission assigned
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 4: Add /roles/{id}/permissions/{pid} endpoint**

```yaml
  /roles/{id}/permissions/{pid}:
    delete:
      tags:
        - roles
      summary: Remove permission from role
      description: Removes a permission from a role. Requires manage_roles permission.
      operationId: removePermission
      x-required-permission: manage_roles
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: Role ID
        - name: pid
          in: path
          required: true
          schema:
            type: integer
          description: Permission ID
      responses:
        '200':
          description: Permission removed
        '400':
          description: Invalid role or permission ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 5: Add roles tag**

```yaml
    - name: roles
      description: Role and permission management endpoints
```

**Step 6: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add roles domain endpoints"
```

---

### Task 5: Add Alerts Domain Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add /alerts endpoints (GET all and POST)**

```yaml
  /alerts:
    get:
      tags:
        - alerts
      summary: List all alert rules
      description: Returns all configured alert rules. Requires view_alerts permission.
      operationId: getAllAlertRules
      x-required-permission: view_alerts
      responses:
        '200':
          description: List of alert rules
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/AlertRule'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    post:
      tags:
        - alerts
      summary: Create an alert rule
      description: Creates a new alert rule for a sensor. Requires manage_alerts permission.
      operationId: createAlertRule
      x-required-permission: manage_alerts
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AlertRule'
      responses:
        '201':
          description: Alert rule created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Invalid request body or validation error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 2: Add /alerts/{sensorId} endpoints (GET, PUT, DELETE)**

```yaml
  /alerts/{sensorId}:
    get:
      tags:
        - alerts
      summary: Get alert rule by sensor ID
      description: Returns the alert rule for a specific sensor. Requires view_alerts permission.
      operationId: getAlertRuleBySensorId
      x-required-permission: view_alerts
      parameters:
        - name: sensorId
          in: path
          required: true
          schema:
            type: integer
          description: Sensor ID
      responses:
        '200':
          description: Alert rule
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AlertRule'
        '400':
          description: Invalid sensor ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '404':
          description: Alert rule not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    put:
      tags:
        - alerts
      summary: Update alert rule
      description: Updates an existing alert rule. Requires manage_alerts permission.
      operationId: updateAlertRule
      x-required-permission: manage_alerts
      parameters:
        - name: sensorId
          in: path
          required: true
          schema:
            type: integer
          description: Sensor ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AlertRule'
      responses:
        '200':
          description: Alert rule updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    delete:
      tags:
        - alerts
      summary: Delete alert rule
      description: Deletes an alert rule. Requires manage_alerts permission.
      operationId: deleteAlertRule
      x-required-permission: manage_alerts
      parameters:
        - name: sensorId
          in: path
          required: true
          schema:
            type: integer
          description: Sensor ID
      responses:
        '200':
          description: Alert rule deleted
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Invalid sensor ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 3: Add /alerts/{sensorId}/history endpoint**

```yaml
  /alerts/{sensorId}/history:
    get:
      tags:
        - alerts
      summary: Get alert history
      description: Returns historical alert events for a sensor. Requires view_alerts permission.
      operationId: getAlertHistory
      x-required-permission: view_alerts
      parameters:
        - name: sensorId
          in: path
          required: true
          schema:
            type: integer
          description: Sensor ID
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 50
          description: Maximum number of records to return (default 50, max 100)
      responses:
        '200':
          description: Alert history
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/AlertHistoryEntry'
        '400':
          description: Invalid sensor ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 4: Add alerts tag**

```yaml
    - name: alerts
      description: Alert rule configuration and history endpoints
```

**Step 5: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add alerts domain endpoints"
```

---

### Task 6: Add Notifications Domain Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add /notifications list endpoint**

```yaml
  /notifications:
    get:
      tags:
        - notifications
      summary: List notifications
      description: Returns notifications for the current user. Requires view_notifications permission.
      operationId: listNotifications
      x-required-permission: view_notifications
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            default: 50
          description: Maximum number of notifications to return
        - name: offset
          in: query
          required: false
          schema:
            type: integer
            default: 0
          description: Offset for pagination
        - name: include_dismissed
          in: query
          required: false
          schema:
            type: string
            enum: ["true", "false"]
            default: "false"
          description: Include dismissed notifications
      responses:
        '200':
          description: List of notifications
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/UserNotification'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 2: Add /notifications/unread-count endpoint**

```yaml
  /notifications/unread-count:
    get:
      tags:
        - notifications
      summary: Get unread notification count
      description: Returns the count of unread notifications. Requires view_notifications permission.
      operationId: getUnreadCount
      x-required-permission: view_notifications
      responses:
        '200':
          description: Unread count
          content:
            application/json:
              schema:
                type: object
                properties:
                  count:
                    type: integer
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 3: Add /notifications/{id}/read and /notifications/{id}/dismiss endpoints**

```yaml
  /notifications/{id}/read:
    post:
      tags:
        - notifications
      summary: Mark notification as read
      description: Marks a specific notification as read. Requires view_notifications permission.
      operationId: markAsRead
      x-required-permission: view_notifications
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: Notification ID
      responses:
        '200':
          description: Marked as read
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Invalid notification ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /notifications/{id}/dismiss:
    post:
      tags:
        - notifications
      summary: Dismiss notification
      description: Dismisses a specific notification. Requires manage_notifications permission.
      operationId: dismissNotification
      x-required-permission: manage_notifications
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: Notification ID
      responses:
        '200':
          description: Dismissed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Invalid notification ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 4: Add bulk operation endpoints**

```yaml
  /notifications/bulk/read:
    post:
      tags:
        - notifications
      summary: Mark all as read
      description: Marks all notifications as read for the current user. Requires view_notifications permission.
      operationId: bulkMarkAsRead
      x-required-permission: view_notifications
      responses:
        '200':
          description: All marked as read
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /notifications/bulk/dismiss:
    post:
      tags:
        - notifications
      summary: Dismiss all notifications
      description: Dismisses all notifications for the current user. Requires manage_notifications permission.
      operationId: bulkDismiss
      x-required-permission: manage_notifications
      responses:
        '200':
          description: All dismissed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 5: Add preferences endpoints**

```yaml
  /notifications/preferences:
    get:
      tags:
        - notifications
      summary: Get notification preferences
      description: Returns channel preferences for notifications. Requires view_notifications permission.
      operationId: getChannelPreferences
      x-required-permission: view_notifications
      responses:
        '200':
          description: Channel preferences
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/ChannelPreference'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    post:
      tags:
        - notifications
      summary: Set notification preference
      description: Sets a channel preference for a notification category. Requires manage_notifications permission.
      operationId: setChannelPreference
      x-required-permission: manage_notifications
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ChannelPreference'
      responses:
        '200':
          description: Preference saved
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Invalid request (category required)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 6: Add /notifications/ws WebSocket endpoint**

```yaml
  /notifications/ws:
    get:
      tags:
        - notifications
      summary: WebSocket endpoint — subscribe to notifications
      description: >-
        Upgrades to WebSocket for real-time notification updates.
        Requires view_notifications permission.
      operationId: notificationsWebSocket
      x-required-permission: view_notifications
      responses:
        '101':
          description: Switching Protocols — connection upgraded to WebSocket
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
      x-websocket:
        description: |
          Receives real-time notification updates for the authenticated user.
          Messages are UserNotification objects.
```

**Step 7: Add notifications tag**

```yaml
    - name: notifications
      description: In-app notification management and real-time updates
```

**Step 8: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add notifications domain endpoints"
```

---

### Task 7: Add OAuth Domain Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add /oauth/status endpoint**

```yaml
  /oauth/status:
    get:
      tags:
        - oauth
      summary: Get OAuth status
      description: Returns the current OAuth configuration status. Requires manage_oauth permission.
      operationId: getOAuthStatus
      x-required-permission: manage_oauth
      responses:
        '200':
          description: OAuth status
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OAuthStatus'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '503':
          description: OAuth not configured
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 2: Add /oauth/authorize endpoint**

```yaml
  /oauth/authorize:
    get:
      tags:
        - oauth
      summary: Get OAuth authorization URL
      description: >-
        Returns an OAuth authorization URL with CSRF state token.
        Used to initiate OAuth flow for Gmail integration.
        Requires manage_oauth permission.
      operationId: getOAuthAuthorizeUrl
      x-required-permission: manage_oauth
      responses:
        '200':
          description: Authorization URL and state
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OAuthAuthorizeResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Failed to generate state or URL
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '503':
          description: OAuth not configured
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 3: Add /oauth/submit-code endpoint**

```yaml
  /oauth/submit-code:
    post:
      tags:
        - oauth
      summary: Submit OAuth authorization code
      description: >-
        Submits the authorization code received from OAuth provider.
        Used with out-of-band OAuth flow. Requires manage_oauth permission.
      operationId: submitOAuthCode
      x-required-permission: manage_oauth
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/OAuthSubmitCodeRequest'
      responses:
        '200':
          description: Authorization successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Invalid request or expired state
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Failed to exchange code
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '503':
          description: OAuth not configured
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 4: Add /oauth/reload endpoint**

```yaml
  /oauth/reload:
    post:
      tags:
        - oauth
      summary: Reload OAuth configuration
      description: >-
        Reloads OAuth credentials and token from disk.
        Requires manage_oauth permission.
      operationId: reloadOAuth
      x-required-permission: manage_oauth
      responses:
        '200':
          description: Configuration reloaded
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Failed to reload
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '503':
          description: OAuth not configured
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 5: Add oauth tag**

```yaml
    - name: oauth
      description: OAuth configuration for Gmail integration
```

**Step 6: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add oauth domain endpoints"
```

---

### Task 8: Add Missing Sensors Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add DELETE /sensors/{name} endpoint**

Note: This was intentionally omitted from original spec. Document with warning.

```yaml
  /sensors/{name}:
    get:
      # ... existing GET endpoint stays ...
    delete:
      tags:
        - sensors
      summary: Delete sensor by name
      description: >-
        **WARNING:** Deletes a sensor and ALL associated historical data.
        This operation is irreversible. Consider using disable instead.
        Requires delete_sensors permission (separate from manage_sensors).
      operationId: deleteSensorByName
      x-required-permission: delete_sensors
      parameters:
        - name: name
          in: path
          required: true
          schema:
            type: string
          description: Sensor name
      responses:
        '200':
          description: Sensor deleted
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Sensor name required
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    head:
      tags:
        - sensors
      summary: Check if sensor exists
      description: Returns 200 if sensor exists, 404 if not. Requires view_sensors permission.
      operationId: sensorExists
      x-required-permission: view_sensors
      parameters:
        - name: name
          in: path
          required: true
          schema:
            type: string
          description: Sensor name
      responses:
        '200':
          description: Sensor exists
        '400':
          description: Sensor name required
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '404':
          description: Sensor not found
        '500':
          description: Server error
```

**Step 2: Add /sensors/collect endpoints**

```yaml
  /sensors/collect:
    post:
      tags:
        - sensors
      summary: Trigger reading collection for all sensors
      description: >-
        Triggers immediate collection of readings from all enabled sensors.
        Requires trigger_readings permission.
      operationId: collectAllSensorReadings
      x-required-permission: trigger_readings
      responses:
        '200':
          description: Collection triggered
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /sensors/collect/{sensorName}:
    post:
      tags:
        - sensors
      summary: Trigger reading collection for specific sensor
      description: >-
        Triggers immediate collection of a reading from a specific sensor.
        Requires trigger_readings permission.
      operationId: collectFromSensor
      x-required-permission: trigger_readings
      parameters:
        - name: sensorName
          in: path
          required: true
          schema:
            type: string
          description: Sensor name
      responses:
        '200':
          description: Collection triggered
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Sensor name required
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 3: Add /sensors/disable and /sensors/enable endpoints**

```yaml
  /sensors/disable/{sensorName}:
    post:
      tags:
        - sensors
      summary: Disable a sensor
      description: >-
        Disables a sensor, stopping automatic collection. Historical data is preserved.
        Requires manage_sensors permission.
      operationId: disableSensor
      x-required-permission: manage_sensors
      parameters:
        - name: sensorName
          in: path
          required: true
          schema:
            type: string
          description: Sensor name
      responses:
        '200':
          description: Sensor disabled
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Sensor name required
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /sensors/enable/{sensorName}:
    post:
      tags:
        - sensors
      summary: Enable a sensor
      description: >-
        Enables a sensor for automatic collection.
        Requires manage_sensors permission.
      operationId: enableSensor
      x-required-permission: manage_sensors
      parameters:
        - name: sensorName
          in: path
          required: true
          schema:
            type: string
          description: Sensor name
      responses:
        '200':
          description: Sensor enabled
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessage'
        '400':
          description: Sensor name required
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Not authenticated
        '403':
          description: Insufficient permissions
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Step 4: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add missing sensors endpoints"
```

---

### Task 9: Update Existing Endpoints with Security Requirements

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add x-required-permission to all existing endpoints**

Update each existing endpoint to include the permission required. For example:

For `/temperature/readings/between`:
```yaml
      x-required-permission: view_readings
```

For `/temperature/readings/hourly/between`:
```yaml
      x-required-permission: view_readings
```

For `/temperature/ws/current-temperatures`:
```yaml
      x-required-permission: view_readings
```

For `/sensors` (POST):
```yaml
      x-required-permission: manage_sensors
```

For `/sensors` (GET):
```yaml
      x-required-permission: view_sensors
```

For `/sensors/{id}` (PUT):
```yaml
      x-required-permission: manage_sensors
```

For `/sensors/{name}` (GET):
```yaml
      x-required-permission: view_sensors
```

For `/sensors/type/{type}`:
```yaml
      x-required-permission: view_sensors
```

For `/sensors/health/{name}`:
```yaml
      x-required-permission: view_sensors
```

For `/sensors/stats/total-readings`:
```yaml
      x-required-permission: view_sensors
```

For `/sensors/ws/{type}`:
```yaml
      # No specific permission check in code beyond auth
```

For `/properties` (GET):
```yaml
      x-required-permission: view_properties
```

For `/properties` (PATCH):
```yaml
      x-required-permission: manage_properties
```

For `/properties/ws`:
```yaml
      x-required-permission: view_properties
```

**Step 2: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add security requirements to existing endpoints"
```

---

### Task 10: Add All New Schema Definitions

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add Auth-related schemas**

Add to `components.schemas`:

```yaml
    LoginRequest:
      type: object
      description: Login request body
      properties:
        username:
          type: string
          description: Username
        password:
          type: string
          description: Password
      required:
        - username
        - password

    LoginResponse:
      type: object
      description: Successful login response
      properties:
        must_change_password:
          type: boolean
          description: Whether user must change password
        csrf_token:
          type: string
          description: CSRF token to include in X-CSRF-Token header for state changes

    RateLimitResponse:
      type: object
      description: Rate limit exceeded response
      properties:
        message:
          type: string
        retry_after:
          type: integer
          description: Seconds until retry is allowed
        failed_by_user:
          type: integer
        failed_by_ip:
          type: integer
        threshold:
          type: integer
        exponent:
          type: integer

    MeResponse:
      type: object
      description: Current user info response
      properties:
        user:
          $ref: '#/components/schemas/User'
        csrf_token:
          type: string
          description: Current CSRF token

    SessionInfo:
      type: object
      description: Session information
      properties:
        id:
          type: integer
          format: int64
        user_id:
          type: integer
        created_at:
          type: string
          format: date-time
        expires_at:
          type: string
          format: date-time
        last_accessed_at:
          type: string
          format: date-time
        ip_address:
          type: string
        user_agent:
          type: string
        current:
          type: boolean
          description: Whether this is the current session
```

**Step 2: Add User-related schemas**

```yaml
    User:
      type: object
      description: User information
      properties:
        id:
          type: integer
        username:
          type: string
        email:
          type: string
        disabled:
          type: boolean
        must_change_password:
          type: boolean
        roles:
          type: array
          items:
            type: string
        permissions:
          type: array
          items:
            type: string
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
      required:
        - id
        - username

    CreateUserRequest:
      type: object
      description: Create user request body
      properties:
        username:
          type: string
        email:
          type: string
        password:
          type: string
        roles:
          type: array
          items:
            type: string
      required:
        - username
        - password

    ChangePasswordRequest:
      type: object
      description: Change password request body
      properties:
        user_id:
          type: integer
          description: User ID (optional, defaults to current user)
        new_password:
          type: string
      required:
        - new_password
```

**Step 3: Add Role/Permission schemas**

```yaml
    RoleInfo:
      type: object
      description: Role information
      properties:
        id:
          type: integer
        name:
          type: string
      required:
        - id
        - name

    PermissionInfo:
      type: object
      description: Permission information
      properties:
        id:
          type: integer
        name:
          type: string
        description:
          type: string
      required:
        - id
        - name
```

**Step 4: Add Alert schemas**

```yaml
    AlertRule:
      type: object
      description: Alert rule configuration
      properties:
        ID:
          type: integer
        SensorID:
          type: integer
        SensorName:
          type: string
        AlertType:
          type: string
          enum: [numeric_range, status_based]
          description: Type of alert (threshold-based or status-based)
        HighThreshold:
          type: number
          format: float
          nullable: true
        LowThreshold:
          type: number
          format: float
          nullable: true
        TriggerStatus:
          type: string
          description: Status that triggers alert (for status_based type)
        Enabled:
          type: boolean
        RateLimitHours:
          type: integer
          description: Minimum hours between alerts
        LastAlertSentAt:
          type: string
          format: date-time
          nullable: true
      required:
        - SensorID
        - AlertType
        - Enabled

    AlertHistoryEntry:
      type: object
      description: Historical alert event
      properties:
        id:
          type: integer
        sensor_id:
          type: integer
        alert_type:
          type: string
        reading_value:
          type: string
        sent_at:
          type: string
          format: date-time
      required:
        - id
        - sensor_id
        - alert_type
        - sent_at
```

**Step 5: Add Notification schemas**

```yaml
    Notification:
      type: object
      description: Notification base object
      properties:
        id:
          type: integer
        category:
          type: string
          enum: [threshold_alert, user_management, config_change]
        severity:
          type: string
          enum: [info, warning, error]
        title:
          type: string
        message:
          type: string
        metadata:
          type: object
          additionalProperties: true
        created_at:
          type: string
          format: date-time
      required:
        - id
        - category
        - severity
        - title
        - message

    UserNotification:
      type: object
      description: User-specific notification with read/dismiss state
      properties:
        id:
          type: integer
        user_id:
          type: integer
        notification_id:
          type: integer
        is_read:
          type: boolean
        is_dismissed:
          type: boolean
        read_at:
          type: string
          format: date-time
          nullable: true
        dismissed_at:
          type: string
          format: date-time
          nullable: true
        notification:
          $ref: '#/components/schemas/Notification'

    ChannelPreference:
      type: object
      description: Notification channel preference
      properties:
        user_id:
          type: integer
        category:
          type: string
          enum: [threshold_alert, user_management, config_change]
        email_enabled:
          type: boolean
        inapp_enabled:
          type: boolean
      required:
        - category
```

**Step 6: Add OAuth schemas**

```yaml
    OAuthStatus:
      type: object
      description: OAuth configuration status
      additionalProperties: true
      example:
        ready: true
        has_credentials: true
        has_token: true

    OAuthAuthorizeResponse:
      type: object
      description: OAuth authorization URL response
      properties:
        auth_url:
          type: string
          description: URL to redirect user for authorization
        state:
          type: string
          description: CSRF state token
      required:
        - auth_url
        - state

    OAuthSubmitCodeRequest:
      type: object
      description: OAuth code submission request
      properties:
        code:
          type: string
          description: Authorization code from OAuth provider
        state:
          type: string
          description: CSRF state token from authorize response
      required:
        - code
        - state
```

**Step 7: Add SuccessMessage schema**

```yaml
    SuccessMessage:
      type: object
      description: Generic success response
      properties:
        message:
          type: string
      required:
        - message
```

**Step 8: Update tags section**

Ensure all new tags are in the `tags` section:

```yaml
  tags:
    - name: temperature
      description: Endpoints related to temperature readings and real-time updates
    - name: sensors
      description: Endpoints related to sensor metadata and control
    - name: properties
      description: Endpoints to read and update application, SMTP, and database properties
    - name: auth
      description: Authentication and session management endpoints
    - name: users
      description: User management endpoints
    - name: roles
      description: Role and permission management endpoints
    - name: alerts
      description: Alert rule configuration and history endpoints
    - name: notifications
      description: In-app notification management and real-time updates
    - name: oauth
      description: OAuth configuration for Gmail integration
```

**Step 9: Commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): add all new schema definitions"
```

---

### Task 11: Validate OpenAPI Spec and Final Commit

**Files:**
- Review: `sensor_hub/openapi.yaml`

**Step 1: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('sensor_hub/openapi.yaml'))"`
Expected: No output (success) or Python error (fix and retry)

**Step 2: (Optional) Validate with OpenAPI linter if available**

If `npx` is available:
```bash
npx @redocly/cli lint sensor_hub/openapi.yaml
```

**Step 3: Review the file visually**

```bash
wc -l sensor_hub/openapi.yaml
head -50 sensor_hub/openapi.yaml
```

Expected: File should be significantly larger than original (~780 lines → ~2000+ lines)

**Step 4: Final commit**

```bash
git add sensor_hub/openapi.yaml
git commit -m "docs(openapi): complete API specification update

Updated OpenAPI spec to reflect current API state:
- Added security schemes (cookie auth, CSRF token)
- Added auth domain endpoints (login, logout, me, sessions)
- Added users domain endpoints (CRUD, password, roles)
- Added roles domain endpoints (list, permissions)
- Added alerts domain endpoints (rules, history)
- Added notifications domain endpoints (list, preferences, WebSocket)
- Added oauth domain endpoints (status, authorize, submit-code, reload)
- Added missing sensors endpoints (delete, head, collect, enable/disable)
- Added permission requirements to all endpoints
- Added all required schema definitions
"
```

**Step 5: Verify git status is clean**

Run: `git --no-pager status`
Expected: Nothing to commit, working tree clean

---

## Execution Notes

- Work through tasks sequentially - each builds on the previous
- Test YAML validity after each major edit
- Commit frequently to enable easy rollback
- The schema names use PascalCase to match Go struct field names where JSON tags use that convention
- Some endpoints have both authentication and permission requirements
- WebSocket endpoints are documented with `x-websocket` extension for tooling awareness
