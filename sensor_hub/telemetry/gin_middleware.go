package telemetry

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// GinLoggerMiddleware returns a Gin middleware that logs HTTP requests via slog.
func GinLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		// Add trace ID if present
		spanCtx := trace.SpanContextFromContext(c.Request.Context())
		if spanCtx.HasTraceID() {
			attrs = append(attrs, "trace_id", spanCtx.TraceID().String())
		}

		// Add user ID if authenticated
		if user, exists := c.Get("currentUser"); exists && user != nil {
			attrs = append(attrs, "auth_method", c.GetString("authMethod"))
		}

		msg := c.Request.Method + " " + path

		switch {
		case status >= 500:
			logger.Error(msg, attrs...)
		case status >= 400:
			logger.Warn(msg, attrs...)
		default:
			logger.Info(msg, attrs...)
		}
	}
}
