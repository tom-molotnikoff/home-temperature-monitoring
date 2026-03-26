---
id: api-keys
title: API Keys
sidebar_position: 6
---

# API Keys

API keys provide token-based authentication for the CLI tool and programmatic access.
All endpoints require the `manage_api_keys` permission.

## Authentication

Include the API key in the `X-API-Key` header:

```
X-API-Key: shk_your_key_here
```

API key requests are exempt from CSRF validation.

## Create API Key

```
POST /api-keys/
```

**Request body:**

```json
{
  "name": "My CLI Key",
  "expires_at": "2026-01-01T00:00:00Z"  // optional
}
```

**Response** (`201 Created`):

```json
{
  "key": "shk_a1b2c3d4e5f6...",
  "message": "Store this key securely. It will not be shown again."
}
```

:::warning
The full key is returned **only once** at creation time. It cannot be retrieved later.
:::

## List API Keys

```
GET /api-keys/
```

Returns all keys belonging to the authenticated user.

**Response** (`200 OK`):

```json
[
  {
    "id": 1,
    "name": "My CLI Key",
    "key_prefix": "shk_a1b2",
    "user_id": 1,
    "expires_at": "2026-01-01T00:00:00Z",
    "revoked": false,
    "last_used_at": "2025-06-15T10:30:00Z",
    "created_at": "2025-06-01T12:00:00Z",
    "updated_at": "2025-06-01T12:00:00Z"
  }
]
```

## Update Key Expiry

```
PATCH /api-keys/:id/expiry
```

**Request body:**

```json
{
  "expires_at": "2027-01-01T00:00:00Z"
}
```

Set `expires_at` to `null` to remove the expiry.

**Response** (`200 OK`):

```json
{
  "message": "expiry updated"
}
```

## Revoke API Key

```
POST /api-keys/:id/revoke
```

Revoked keys can no longer be used for authentication.

**Response** (`200 OK`):

```json
{
  "message": "api key revoked"
}
```

## Delete API Key

```
DELETE /api-keys/:id
```

Permanently removes the key.

**Response** (`200 OK`):

```json
{
  "message": "api key deleted"
}
```
