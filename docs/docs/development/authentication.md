# Authentication and Authorisation

This guide explains how authentication, session management, CSRF protection,
and role-based access control work across the full stack.

## Authentication Methods

Sensor Hub supports two authentication methods, checked in this order:

1. **API Key** — `X-API-Key` header, used by the CLI and programmatic clients
2. **Session Cookie** — `sensor_hub_session` HTTP-only cookie, used by the React UI

### Session-Based Auth (Browser)

```
Browser                         Sensor Hub
  │                                │
  ├─ POST /api/auth/login ────────►│
  │  {username, password}          │ Validate credentials
  │                                │ Create session (token_hash + csrf_token)
  │◄── Set-Cookie: sensor_hub_session=<token>
  │    Body: {csrf_token, must_change_password}
  │                                │
  ├─ GET /api/sensors/ ───────────►│
  │  Cookie: sensor_hub_session    │ Validate session
  │                                │ Set currentUser in context
  │◄── 200 OK [{...}]             │
  │                                │
  ├─ POST /api/sensors/ ──────────►│
  │  Cookie: sensor_hub_session    │ Validate CSRF token
  │  X-CSRF-Token: <token>         │ Validate session
  │  {name, type, url}             │ Check manage_sensors permission
  │◄── 201 Created                │
```

**Token storage:** The session token is stored as a SHA256 hash in the
`sessions` table. The raw token is never persisted — only the hash. On each
request the incoming cookie value is hashed and looked up.

**CSRF token:** Generated when the session is created and stored alongside it.
Returned to the client in the login response body. The frontend stores it
in-memory (not in localStorage or cookies) and sends it as the `X-CSRF-Token`
header on all state-changing requests.

### API Key Auth (CLI / Programmatic)

```
CLI                             Sensor Hub
  │                                │
  ├─ GET /api/sensors/ ───────────►│
  │  X-API-Key: shk_abc123...     │ Hash key, look up in api_keys table
  │                                │ Check not expired or revoked
  │                                │ Load owning user and permissions
  │◄── 200 OK [{...}]             │
```

API key authentication bypasses CSRF checks entirely (API keys are not
vulnerable to CSRF attacks since they are not automatically sent by the
browser).

API keys are hashed with SHA256 before storage. The `key_prefix` field stores a
short displayable prefix (e.g. `shk_abc12...`) so users can identify their keys
without exposing the full value.

## CSRF Protection

CSRF middleware (`api/middleware/csrf_middleware.go`) runs on all routes under
`/api` and:

1. **Skips** GET, HEAD, OPTIONS requests (safe methods)
2. **Skips** requests with an `X-API-Key` header
3. **Skips** `/api/auth/login` and `/api/auth/logout`
4. For everything else: reads the session cookie, retrieves the expected CSRF
   token from the database, and compares it with the `X-CSRF-Token` header.
   Returns 403 if they do not match.

### Frontend CSRF Flow

```typescript
// On login: store the CSRF token in memory
const response = await Auth.login(username, password);
setCsrfToken(response.csrf_token);

// On every state-changing request: send it as a header
// (handled automatically by Client.ts)
headers['X-CSRF-Token'] = getCsrfToken();
```

The token is refreshed whenever the client calls `/api/auth/me` (which also
returns the current CSRF token).

## Must Change Password

When a user is created (including the initial admin), `must_change_password` is
set to `true`. While this flag is set, the auth middleware only allows access
to:

- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`
- `PUT /api/users/password`

All other endpoints return 403 Forbidden. The flag is cleared after the user
changes their password.

## Role-Based Access Control

### Roles

Three roles are created by the initial migration:

| Role | Description |
|------|-------------|
| `admin` | Full access to all features |
| `user` | Standard access (read + some write) |
| `viewer` | Read-only access |

### Permissions

There are permissions defined in the database:

| Permission | Description |
|-----------|-------------|
| `manage_users` | Create, update, disable, delete users |
| `view_users` | View user list |
| `manage_sensors` | Add, update sensors |
| `view_sensors` | View sensor list and details |
| `delete_sensors` | Delete sensors |
| `view_readings` | View temperature data |
| `trigger_readings` | Manually trigger sensor collection |
| `view_roles` | View role list |
| `manage_roles` | Modify role permissions |
| `view_properties` | View configuration |
| `manage_properties` | Update configuration |
| `view_alerts` | View alert rules and history |
| `manage_alerts` | Create, update, delete alert rules |
| `manage_oauth` | Configure OAuth settings |
| `view_notifications` | View in-app notifications |
| `view_notifications_user_mgmt` | View notification user settings |
| `view_notifications_config` | View notification channel preferences |
| `manage_notifications` | Modify notification preferences |
| `manage_api_keys` | Create and manage API keys |

The admin role is granted all permissions by default.

### How Permission Checks Work

The `RequirePermission` middleware:

1. Reads `currentUser` from the Gin context (set by `AuthRequired`)
2. Checks if the user's permissions (loaded during session validation) include
   the required permission
3. Returns 403 Forbidden if not

```go
// Route registration with permission check
sensorsGroup.POST("/", middleware.AuthRequired(),
    middleware.RequirePermission("manage_sensors"), addSensorHandler)
```

Permission comparison is case-insensitive (`strings.EqualFold`).

### Adding a New Permission

1. Create a new migration that inserts into the `permissions` table
2. Optionally grant it to the admin role via `role_permissions`
3. Use `middleware.RequirePermission("your_permission")` on the relevant routes

## Login Rate Limiting

Failed login attempts are tracked in the `failed_login_attempts` table. The
system applies exponential backoff when the number of failures within a
configurable window exceeds a threshold:

| Config Property | Default | Description |
|----------------|---------|-------------|
| `auth.login.backoff.window.minutes` | 15 | Time window for counting failures |
| `auth.login.backoff.threshold` | 5 | Failures before backoff kicks in |
| `auth.login.backoff.base.seconds` | 2 | Initial backoff duration |
| `auth.login.backoff.max.seconds` | 300 | Maximum backoff duration |

Rate limiting is applied per-username and per-IP-address independently.

## OAuth (Gmail SMTP)

OAuth 2.0 is used specifically for Gmail SMTP integration (sending alert
emails). It is not a user-facing login method.

The `pre_authorise_application/` directory contains a helper tool that generates
the required `credentials.json` and `token.json` files. Once configured, the
OAuth service periodically refreshes the access token in the background.
