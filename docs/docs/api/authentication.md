---
id: authentication
title: Authentication
sidebar_position: 1
---

# Authentication API

Sensor Hub uses session-based authentication. Clients authenticate by sending credentials to the login endpoint, which returns a session cookie and a CSRF token. The session cookie is included automatically in subsequent requests. The CSRF token must be sent as a header on state-changing requests.

> All paths below are relative to the `/api` base path (e.g. `POST /auth/login` is served at `POST /api/auth/login`).

## CSRF protection

All POST, PUT, PATCH, and DELETE requests (except `/api/auth/login` and `/api/auth/logout`) require the `X-CSRF-Token` header. The CSRF token is returned in the response body of the login and `/api/auth/me` endpoints.

## Session cookie

The session cookie is set with the following flags:

| Flag     | Value                                        |
|----------|----------------------------------------------|
| Name     | Configurable (default: `sensor_hub_session`) |
| HttpOnly | `true` (not accessible via JavaScript)       |
| Secure   | `true` in production (HTTPS only)            |
| SameSite | `Lax`                                        |

---

## POST /auth/login

Authenticate a user and create a new session.

Authentication: not required.
CSRF: not required.

### Request body

```json
{
  "username": "admin",
  "password": "your_password"
}
```

### Response (200 OK)

Sets the session cookie via the `Set-Cookie` header and returns the CSRF token.

```json
{
  "must_change_password": false,
  "csrf_token": "dGhpcyBpcyBhIGNzcmYgdG9rZW4"
}
```

If `must_change_password` is `true`, the user is restricted to the password change endpoint until they update their password.

### Response (401 Unauthorized)

Returned when the username or password is incorrect.

```json
{
  "message": "invalid credentials"
}
```

### Response (429 Too Many Requests)

Returned when the login rate limit has been exceeded. Includes a `Retry-After` header.

```json
{
  "message": "too many failed login attempts, retry later",
  "retry_after": 8,
  "failed_by_user": 6,
  "failed_by_ip": 3,
  "threshold": 5,
  "exponent": 1
}
```

---

## POST /auth/logout

Invalidate the current session and clear the session cookie.

Authentication: required.
CSRF: not required.

### Response (200 OK)

Empty response body. The session cookie is cleared.

---

## GET /auth/me

Return the authenticated user's profile and a fresh CSRF token. Use this endpoint on page load to verify the session and retrieve the current user's permissions.

Authentication: required.

### Response (200 OK)

```json
{
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "disabled": false,
    "must_change_password": false,
    "roles": ["admin"],
    "permissions": [
      "manage_sensors",
      "view_sensors",
      "view_readings",
      "manage_alerts",
      "view_alerts",
      "manage_users",
      "view_users",
      "manage_roles",
      "view_roles",
      "manage_properties",
      "view_properties",
      "manage_oauth",
      "view_notifications",
      "manage_notifications"
    ],
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-01-15T10:00:00Z"
  },
  "csrf_token": "dGhpcyBpcyBhIGNzcmYgdG9rZW4"
}
```

---

## Example: authenticated request flow

1. Log in to obtain the session cookie and CSRF token:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your_password"}' \
  -c cookies.txt
```

2. Use the session cookie and CSRF token for subsequent requests:

```bash
curl http://localhost:8080/auth/me \
  -b cookies.txt
```

3. Include the CSRF token on state-changing requests:

```bash
curl -X POST http://localhost:8080/sensors \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: <csrf_token_from_login>" \
  -b cookies.txt \
  -d '{"name": "Downstairs", "type": "temperature", "url": "http://192.168.1.50:5000"}'
```
