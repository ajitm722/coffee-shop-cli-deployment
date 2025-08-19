package version

// Version information for the application.
// These variables can be set during build time using ldflags.
var (
	Version   = "dev"     // Application version (default: dev).
	Commit    = "none"    // Git commit hash (default: none).
	BuildTime = "unknown" // Build timestamp (default: unknown).
)
