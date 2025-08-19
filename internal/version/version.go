package version

// Version information for the application.
// These variables can be set during build time using ldflags.

// Version represents the application version (default: dev).
var Version = "dev"

// Commit represents the Git commit hash (default: none).
var Commit = "none"

// BuildTime represents the build timestamp (default: unknown).
var BuildTime = "unknown"
