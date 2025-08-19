package server

import (
	"log/slog"
	"time"
)

type options struct {
	addr            string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
	logger          *slog.Logger
	logStartup      bool
}

type Option func(*options)

func WithAddr(addr string) Option {
	return func(o *options) { o.addr = addr }
}

func WithTimeouts(read, write, idle time.Duration) Option {
	return func(o *options) {
		o.readTimeout = read
		o.writeTimeout = write
		o.idleTimeout = idle
	}
}

// NEW: carry shutdown timeout into the server (for startup logging)
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *options) { o.shutdownTimeout = d }
}

// NEW: toggle logging of config at startup
func WithStartupConfigLog() Option {
	return func(o *options) { o.logStartup = true }
}

func WithLogger(l *slog.Logger) Option {
	return func(o *options) { o.logger = l }
}
