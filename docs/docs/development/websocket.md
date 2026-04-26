# WebSocket

Sensor Hub uses WebSocket connections to push real-time updates to connected
browser clients. When sensor readings are collected, sensors are updated, or
notifications are created, the backend broadcasts the change over WebSocket so
the UI updates without polling.

## Topics

Topics are string identifiers. Services broadcast to specific topics, and
clients subscribe when their WebSocket connection opens.

## Message Formats

All messages are sent as JSON.
`

## Environment Configuration

In production, WebSocket connections go through the same origin. In development
with Vite, set `VITE_WEBSOCKET_BASE` to point to the Go backend directly (the
Vite dev server proxies HTTP requests but WebSocket connections need the direct
URL).
