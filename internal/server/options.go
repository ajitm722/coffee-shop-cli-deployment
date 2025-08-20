package server

import (
	"database/sql"
	"log/slog"
	"time"
)

// options struct holds configuration values for the server.
// These values are set using functional options.
type options struct {
	addr            string        // Address to listen on.
	readTimeout     time.Duration // Timeout for reading requests.
	writeTimeout    time.Duration // Timeout for writing responses.
	idleTimeout     time.Duration // Timeout for idle connections.
	shutdownTimeout time.Duration // Timeout for graceful shutdown.
	logger          *slog.Logger  // Structured logger for logging.
	logStartup      bool          // Whether to log configuration at startup.
	ginMode         *string       // Gin mode (e.g., release, test).
	db              *sql.DB       // Database connection.
	accessLog       bool          // Whether to log access details for each request.
}

// Option defines a functional option for configuring the server.
type Option func(*options)

// WithAddr sets the server's listening address.
func WithAddr(addr string) Option {
	return func(o *options) { o.addr = addr }
}

// WithTimeouts sets the read, write, and idle timeouts for the server.
func WithTimeouts(read, write, idle time.Duration) Option {
	return func(o *options) {
		o.readTimeout = read
		o.writeTimeout = write
		o.idleTimeout = idle
	}
}

// WithShutdownTimeout sets the timeout for graceful shutdown.
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *options) { o.shutdownTimeout = d }
}

// WithStartupConfigLog enables logging of server configuration at startup.
func WithStartupConfigLog() Option {
	return func(o *options) { o.logStartup = true }
}

// WithLogger sets the structured logger for the server.
func WithLogger(l *slog.Logger) Option {
	return func(o *options) { o.logger = l }
}

// WithGinMode sets the Gin mode (e.g., release, test).
func WithGinMode(mode string) Option {
	return func(o *options) { o.ginMode = &mode }
}

// WithDB sets the database connection for the server.
func WithDB(db *sql.DB) Option {
	return func(o *options) { o.db = db }
}

// WithAccessLog enables AccessLogMiddleware for /v1 routes.
func WithAccessLog() Option {
	return func(o *options) { o.accessLog = true }
}
