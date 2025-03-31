package cmd

import (
	"dt-geo-converter/commands"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	initDBFile string
	initDir    string
	initUpdate bool
)

var initDBCmd = &cobra.Command{
	Use:   "init-db",
	Short: "Initialize the database with CSV data",
	Long: `Initialize the database with CSV data.

Usage:
  dt-geo-converter init-db --dir <directory> [--db <dbfile>] [--update]`,
	Run: func(cmd *cobra.Command, args []string) {
		if initDir == "" {
			fmt.Println("The --dir flag is required.")
			cmd.Help()
			os.Exit(1)
		}
		commands.InitDatabase(initDBFile, initDir, initUpdate)
	},
}

func init() {
	rootCmd.AddCommand(initDBCmd)
	initDBCmd.Flags().StringVar(&initDBFile, "db", "./db.db", "Path to the database file (optional)")
	initDBCmd.Flags().StringVar(&initDir, "dir", "", "Directory containing CSV files or subdirectories with CSV files (required)")
	initDBCmd.Flags().BoolVar(&initUpdate, "update", false, "Reset and reinitialize the database if it exists")
}
