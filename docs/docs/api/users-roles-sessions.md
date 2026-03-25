---
id: users-roles-sessions
title: Users, Roles, and Sessions
sidebar_position: 4
---

# Users, Roles, and Sessions API

Endpoints for managing user accounts, roles, permissions, and sessions. All endpoints require authentication and the specified permission.

> All paths below are relative to the `/api` base path (e.g. `POST /users` is served at `POST /api/users`).

---

## User endpoints

### POST /users

Create a new user account.

Permission: `manage_users`

#### Request body

```json
{
  "username": "jdoe",
  "email": "jdoe@example.com",
  "password": "secure_password",
  "roles": ["user"]
}
```

#### Response (201 Created)

```json
{
  "id": 2
}
```

---

### GET /users

List all user accounts.

Permission: `view_users`

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "disabled": false,
    "must_change_password": false,
    "roles": ["admin"],
    "permissions": ["manage_sensors", "view_sensors", "view_readings"],
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-01-15T10:00:00Z"
  }
]
```

---

### PUT /users/password

Change a user's password.

Authentication: required (any authenticated user).

When called without a `user_id`, changes the current user's password. Administrators can change another user's password by providing their `user_id`.

#### Request body

```json
{
  "user_id": 1,
  "new_password": "new_secure_password"
}
```

The `user_id` field is optional. If omitted, the password is changed for the currently authenticated user.

#### Response (200 OK)

Empty response body.

---

### DELETE /users/:id

Delete a user account.

Permission: `manage_users`

You cannot delete your own account.

#### Path parameters

| Parameter | Type    | Description |
|-----------|---------|-------------|
| `id`      | integer | User ID     |

#### Response (200 OK)

Empty response body.

---

### PATCH /users/:id/must\_change

Set or clear the "must change password" flag on a user account. When set, the user is restricted to the password change endpoint until they update their password.

Permission: `manage_users`

#### Path parameters

| Parameter | Type    | Description |
|-----------|---------|-------------|
| `id`      | integer | User ID     |

#### Request body

```json
{
  "must_change": true
}
```

#### Response (200 OK)

Empty response body.

---

### POST /users/:id/roles

Set the roles for a user. This replaces the user's current roles with the specified list.

Permission: `manage_users`

#### Path parameters

| Parameter | Type    | Description |
|-----------|---------|-------------|
| `id`      | integer | User ID     |

#### Request body

```json
{
  "roles": ["admin", "user"]
}
```

#### Response (200 OK)

Empty response body.

---

## Role endpoints

### GET /roles

List all available roles.

Permission: `view_roles`

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "name": "admin"
  },
  {
    "id": 2,
    "name": "user"
  },
  {
    "id": 3,
    "name": "viewer"
  }
]
```

---

### GET /roles/permissions

List all available permissions in the system.

Permission: `view_roles`

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "name": "manage_sensors",
    "description": "Add and edit sensors"
  },
  {
    "id": 2,
    "name": "view_sensors",
    "description": "View sensor list and data"
  }
]
```

---

### GET /roles/:id/permissions

Get the permissions assigned to a specific role.

Permission: `view_roles`

#### Path parameters

| Parameter | Type    | Description |
|-----------|---------|-------------|
| `id`      | integer | Role ID     |

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "name": "manage_sensors",
    "description": "Add and edit sensors"
  }
]
```

---

### POST /roles/:id/permissions/:permissionId

Grant a permission to a role.

Permission: `manage_roles`

#### Path parameters

| Parameter      | Type    | Description   |
|----------------|---------|---------------|
| `id`           | integer | Role ID       |
| `permissionId` | integer | Permission ID |

#### Response (200 OK)

Empty response body.

---

### DELETE /roles/:id/permissions/:permissionId

Revoke a permission from a role.

Permission: `manage_roles`

#### Path parameters

| Parameter      | Type    | Description   |
|----------------|---------|---------------|
| `id`           | integer | Role ID       |
| `permissionId` | integer | Permission ID |

#### Response (200 OK)

Empty response body.

---

## Session endpoints

### GET /auth/sessions

List all active sessions for the current user.

Authentication: required.

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "user_id": 1,
    "created_at": "2026-01-15T10:00:00Z",
    "expires_at": "2026-02-14T10:00:00Z",
    "last_accessed_at": "2026-01-15T14:30:00Z",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
    "current": true
  },
  {
    "id": 2,
    "user_id": 1,
    "created_at": "2026-01-14T08:00:00Z",
    "expires_at": "2026-02-13T08:00:00Z",
    "last_accessed_at": "2026-01-14T18:00:00Z",
    "ip_address": "192.168.1.101",
    "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)",
    "current": false
  }
]
```

The `current` field indicates whether the session belongs to the request that is listing sessions.

---

### DELETE /auth/sessions/:id

Revoke a specific session. The affected client is immediately logged out.

Authentication: required.

#### Path parameters

| Parameter | Type    | Description |
|-----------|---------|-------------|
| `id`      | integer | Session ID  |

#### Response (200 OK)

Empty response body.
