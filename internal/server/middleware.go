package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ctxKey is a custom type used to define context keys.
// This prevents accidental collisions with other context values.
type ctxKey string

const requestIDKey ctxKey = "request_id" // Key used to store the request ID in the context.

// RequestIDMiddleware generates or retrieves a unique request ID for each incoming HTTP request.
// It ensures that the request ID is available in the context and response headers for tracking purposes.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		ctx := context.WithValue(c.Request.Context(), requestIDKey, rid)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

// AccessLogMiddleware logs details about each HTTP request, including method, path, status, latency, and client IP.
// It also includes the request ID for correlation and logs errors if any occur during request processing.
func AccessLogMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		rid, _ := c.Request.Context().Value(requestIDKey).(string)
		attrs := []any{
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", dur.Milliseconds(),
			"request_id", rid,
			"client_ip", c.ClientIP(),
		}
		if len(c.Errors) > 0 {
			log.Error("request", append(attrs, "error", c.Errors.String())...)
			return
		}
		log.Info("request", attrs...)
	}
}

// RequestIDFromContext retrieves the request ID from the given context.
// This is useful for accessing the request ID in other parts of the application.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(requestIDKey).(string)
	return v, ok
}
