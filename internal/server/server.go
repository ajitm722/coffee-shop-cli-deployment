package server

import (
	"context" // context package for managing request contexts and timeouts.
	// sql package for database interactions, specifically PostgreSQL.
	"fmt"
	"log/slog" // slog package for structured logging.
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver for database interactions.
)

// Server represents the HTTP server and its configuration.
// It encapsulates the Gin engine, HTTP server, logger, and other settings.
type Server struct {
	http     *http.Server  // Underlying HTTP server.
	eng      *gin.Engine   // Gin engine for routing.
	log      *slog.Logger  // Structured logger for logging.
	addr     string        // Address to listen on.
	shutdown time.Duration // Timeout duration for graceful shutdown.
	logStart bool          // Whether to log configuration at startup.
}

// NewServer initializes a new Server instance with the provided options.
// It configures the Gin engine, HTTP server, and middleware.
func NewServer(opts ...Option) *Server {
	cfg := &options{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
	}
	// no defaults for timeouts here

	for _, o := range opts {
		o(cfg)
	}

	if cfg.ginMode != nil {
		gin.SetMode(*cfg.ginMode)
	}
	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(RequestIDMiddleware())

	api := e.Group("/v1")
	api.Use(AccessLogMiddleware(cfg.logger))
	registerHealthRoutes(api)
	if cfg.db != nil {
		registerMenuRoutes(api, cfg.db)       // Register menu routes if a database connection is provided.
		registerOrderRoutes(api, cfg.db)      // Register order routes if a database connection is provided.
		registerOrderListRoutes(api, cfg.db)  // Register order list routes if a database connection is provided.
		registerOrderClearRoutes(api, cfg.db) // Register order clear routes if a database connection is provided.
	}

	hs := &http.Server{
		Addr:         cfg.addr,
		Handler:      e,
		ReadTimeout:  cfg.readTimeout,
		WriteTimeout: cfg.writeTimeout,
		IdleTimeout:  cfg.idleTimeout,
	}

	return &Server{
		http:     hs,
		eng:      e,
		log:      cfg.logger,
		addr:     cfg.addr,
		shutdown: cfg.shutdownTimeout, // NEW
		logStart: cfg.logStartup,      // NEW
	}
}

// Start begins listening for incoming HTTP requests.
// It logs the server configuration if enabled and starts the HTTP server.
func (s *Server) Start() error {
	// NEW: log the final, resolved config at startup
	if s.logStart {
		s.log.Info("server.config",
			"addr", s.addr,
			"read_timeout", s.http.ReadTimeout.String(),
			"write_timeout", s.http.WriteTimeout.String(),
			"idle_timeout", s.http.IdleTimeout.String(),
			"shutdown_timeout", s.shutdown.String(),
		)
	}

	s.log.Info("starting http server", "addr", s.addr)
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server start: %w", err)
	}
	return nil
}

// Engine exposes the underlying Gin engine for tests.
// This allows test code to interact directly with the routing engine.
func (s *Server) Engine() *gin.Engine {
	return s.eng
}

// Logger exposes the structured logger (slog) for callers/tests.
// This allows external code to log messages using the server's logger.
func (s *Server) Logger() *slog.Logger {
	return s.log
}

// Stop performs a graceful shutdown with the provided context deadline.
// It ensures the server shuts down cleanly within the specified timeout.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("stopping http server")
	return s.http.Shutdown(ctx)
}
