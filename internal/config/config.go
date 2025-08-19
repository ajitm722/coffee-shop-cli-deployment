package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Load merges defaults + config file (if present) + env (COFFEE_*) + flags.
func Load() Config {
	v := viper.GetViper()

	// ------------- Defaults -------------
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

	if err := v.ReadInConfig(); err == nil {
		log.Printf("loaded config file: %s", v.ConfigFileUsed())
	} else {
		log.Printf("no config file found; using defaults/env/flags")
	}

	// ------------- Env -------------
	v.SetEnvPrefix("coffee") // COFFEE_*
	v.AutomaticEnv()

	return Config{
		Addr:            v.GetString("addr"),
		ReadTimeout:     v.GetDuration("read_timeout"),
		WriteTimeout:    v.GetDuration("write_timeout"),
		IdleTimeout:     v.GetDuration("idle_timeout"),
		ShutdownTimeout: v.GetDuration("shutdown_timeout"),
	}
}
