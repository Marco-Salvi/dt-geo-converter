package cmd

import (
	"dt-geo-converter/commands"

	"github.com/spf13/cobra"
)

var listDBFile string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflows available in the database",
	Run: func(cmd *cobra.Command, args []string) {
		commands.ListWorkflows(listDBFile)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listDBFile, "db", "./db.db", "Path to the database file (optional)")
}
