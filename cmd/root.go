package cmd

import (
	"dt-geo-converter/logger"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var debug bool

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "dt-geo-converter",
	Short: "A tool to manage the dt-geo database and workflow conversion",
	Long:  "A CLI tool to initialize the database with CSV data, convert workflows into CWL, generate workflow graphs, and produce RO‑Crate metadata.",
	// PersistentPreRun runs before any subcommand.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			logger.DebugEnabled = true
			logger.Info("Debug mode enabled")
		}
	},
	// If no subcommand is provided, show help.
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global persistent flag for debug mode.
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
