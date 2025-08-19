package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd defines the root command for the CLI application.
// It serves as the entry point for all subcommands.
var rootCmd = &cobra.Command{
	Use:   "coffee",
	Short: "Coffee shop API (CLI-first)",
	Long:  "Minimal hexagonal Go service. Step 1 only: boot + config.",
}

// Execute runs the root command and handles errors.
// It ensures the application exits with an appropriate status code.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
