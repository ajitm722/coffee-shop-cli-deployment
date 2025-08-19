package cmd

import (
	// Import necessary packages for server setup, configuration, and signal handling
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coffee/internal/config"
	"coffee/internal/db"
	"coffee/internal/server"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // NEW: for loading .env files
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// init initializes the serve command and binds flags to configuration values.
// Flags allow users to override configuration values via the CLI.
func init() {
	// NEW: config file flag
	serveCmd.Flags().String("config", "", "path to config file (e.g., ./config.yaml)")
	_ = viper.BindPFlag("config", serveCmd.Flags().Lookup("config"))

	serveCmd.Flags().String("addr", ":8080", "HTTP listen address")
	serveCmd.Flags().Duration("read-timeout", 5*time.Second, "HTTP server read timeout")
	serveCmd.Flags().Duration("write-timeout", 10*time.Second, "HTTP server write timeout")
	serveCmd.Flags().Duration("idle-timeout", 30*time.Second, "HTTP server idle timeout")
	serveCmd.Flags().Duration("shutdown-timeout", 5*time.Second, "graceful shutdown timeout")

	_ = viper.BindPFlag("addr", serveCmd.Flags().Lookup("addr"))
	_ = viper.BindPFlag("read_timeout", serveCmd.Flags().Lookup("read-timeout"))
	_ = viper.BindPFlag("write_timeout", serveCmd.Flags().Lookup("write-timeout"))
	_ = viper.BindPFlag("idle_timeout", serveCmd.Flags().Lookup("idle-timeout"))
	_ = viper.BindPFlag("shutdown_timeout", serveCmd.Flags().Lookup("shutdown-timeout"))

	rootCmd.AddCommand(serveCmd)
}

// serveCmd defines the "serve" command, which starts the HTTP API server.
// It loads configuration, initializes the server, and handles graceful shutdown.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// NEW: Load environment variables from .env file
		if err := godotenv.Load(".env"); err != nil {
			return err
		}

		cfg := config.Load()

		dbURL := os.Getenv("DB_URL")
		if dbURL == "" {
			return fmt.Errorf("DB_URL not set in environment")
		}

		db, err := db.Connect(dbURL)
		if err != nil {
			return err
		}
		defer db.Close()

		srv := server.NewServer(
			server.WithAddr(cfg.Addr),
			server.WithTimeouts(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout),
			server.WithShutdownTimeout(cfg.ShutdownTimeout),
			server.WithStartupConfigLog(),
			server.WithGinMode(gin.ReleaseMode), // <— explicit for runtime
			server.WithDB(db),                   // Pass the database connection
		)

		errCh := make(chan error, 1)
		go func() { errCh <- srv.Start() }()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-sigCh:
			srv.Logger().Info("shutdown signal received", "signal", sig.String())
		case err := <-errCh:
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return srv.Stop(ctx)
	},
}
