package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds the server configuration values.
// These values are loaded from defaults, config files, environment variables, and flags.
type Config struct {
	Addr            string        // Address to listen on.
	ReadTimeout     time.Duration // Timeout for reading requests.
	WriteTimeout    time.Duration // Timeout for writing responses.
	IdleTimeout     time.Duration // Timeout for idle connections.
	ShutdownTimeout time.Duration // Timeout for graceful shutdown.
}

// Load merges defaults, config file, environment variables, and flags into a Config struct.
// It ensures the application has all necessary configuration values.
func Load() Config {
	v := viper.GetViper()

	// ------------- Defaults -------------
	// Set default values for the configuration.
	v.SetDefault("addr", ":8080")
	v.SetDefault("read_timeout", 5*time.Second)
	v.SetDefault("write_timeout", 10*time.Second)
	v.SetDefault("idle_timeout", 30*time.Second)
	v.SetDefault("shutdown_timeout", 5*time.Second)

	// ------------- Config file -------------
	// If a --config (or COFFEE_CONFIG) was provided, use that exact file.
	if cfgFile := v.GetString("config"); cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		// Otherwise, search common locations for "config.yaml".
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/coffee")
	}

	// Attempt to read the configuration file.
	if err := v.ReadInConfig(); err == nil {
		log.Printf("loaded config file: %s", v.ConfigFileUsed())
	} else {
		log.Printf("no config file found; using defaults/env/flags")
	}

	// ------------- Env -------------
	// Load environment variables with the prefix "COFFEE_".
	v.SetEnvPrefix("coffee") // COFFEE_*
	v.AutomaticEnv()

	// Return the final configuration values.
	return Config{
		Addr:            v.GetString("addr"),
		ReadTimeout:     v.GetDuration("read_timeout"),
		WriteTimeout:    v.GetDuration("write_timeout"),
		IdleTimeout:     v.GetDuration("idle_timeout"),
		ShutdownTimeout: v.GetDuration("shutdown_timeout"),
	}
}
