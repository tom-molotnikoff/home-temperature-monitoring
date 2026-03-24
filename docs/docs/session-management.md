---
id: session-management
title: Session Management
sidebar_position: 8
---

# Session Management

Sensor Hub uses session-based authentication. When a user logs in, the server creates a session and issues a session token as an HTTP-only cookie. All subsequent requests are authenticated using this cookie.

## Session details

Each session records:

- The client IP address
- The browser user agent
- Creation time
- Last accessed time
- Expiry time

Sessions expire after a configurable period (default: 30 days). The expiry is checked on each request; expired sessions are automatically invalidated.

A CSRF token is generated alongside each session to protect against cross-site request forgery on state-changing requests. The CSRF token must be sent as the `X-CSRF-Token` header on POST, PUT, PATCH, and DELETE requests. It is returned in the response body when logging in or calling the `/auth/me` endpoint.

## Viewing active sessions

The Sessions page in the web UI lists all active sessions for the currently logged-in user. Each entry shows:

- IP address
- Browser user agent
- When the session was created
- When it was last accessed
- When it expires
- Whether it is the current session

## Revoking sessions

You can revoke any of your own sessions from the Sessions page. Revoking a session immediately invalidates it. Any client using that session must log in again.

Session revocations are recorded in an audit trail that captures who revoked the session and when.

## Login rate limiting

To protect against brute-force attacks, Sensor Hub tracks failed login attempts per username and per IP address.

After exceeding a configurable threshold (default: 5 failed attempts within 15 minutes), an exponential backoff is applied. The API returns a `429 Too Many Requests` response with a `Retry-After` header indicating when the next login attempt is allowed.

The backoff starts at a base duration (default: 2 seconds) and increases exponentially with each subsequent failure, up to a maximum (default: 300 seconds / 5 minutes).

## Session configuration

The following properties control session behavior. See [Configuration Settings](configuration) for the full reference.

| Property                            | Default              | Description                                      |
|-------------------------------------|----------------------|--------------------------------------------------|
| `auth.session.ttl.minutes`          | `43200`              | Session duration in minutes (30 days)            |
| `auth.session.cookie.name`          | `sensor_hub_session` | Name of the session cookie                       |
| `auth.login.backoff.window.minutes` | `15`                 | Time window for counting failed login attempts   |
| `auth.login.backoff.threshold`      | `5`                  | Number of failed attempts before backoff applies |
| `auth.login.backoff.base.seconds`   | `2`                  | Base duration for exponential backoff            |
| `auth.login.backoff.max.seconds`    | `300`                | Maximum backoff duration                         |
