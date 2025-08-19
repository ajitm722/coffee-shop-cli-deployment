package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	http     *http.Server
	eng      *gin.Engine
	log      *slog.Logger
	addr     string
	shutdown time.Duration // NEW: store for logging
	logStart bool          // NEW: whether to log at startup
}

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
func (s *Server) Engine() *gin.Engine {
	return s.eng
}

// Logger exposes the structured logger (slog) for callers/tests.
func (s *Server) Logger() *slog.Logger {
	return s.log
}

// Stop performs a graceful shutdown with the provided context deadline.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("stopping http server")
	return s.http.Shutdown(ctx)
}
