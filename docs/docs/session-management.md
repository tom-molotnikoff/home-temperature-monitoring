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

## Revoking sessions

You can revoke any of your own sessions from the Sessions page or through the REST API. Revoking a session immediately invalidates it. Any client using that session must log in again.

## Login rate limiting

To protect against brute-force attacks, Sensor Hub tracks failed login attempts per username and per IP address.

After exceeding a configurable threshold (default: 5 failed attempts within 15 minutes), an exponential backoff is applied.

The backoff starts at a base duration and increases exponentially with each subsequent failure, up to a maximum.

## Session configuration

See [Configuration Settings](configuration) for the configuration reference.

