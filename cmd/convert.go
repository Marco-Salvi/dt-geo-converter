package cmd

import (
	"dt-geo-converter/commands"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	convertDBFile string
	convertDir    string
	workflowID    string
	convertAll    bool
	convertUpdate bool
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert workflow(s) from the database into CWL and generate workflow graphs",
	Long: `Convert workflow(s) from the database into CWL and generate workflow graphs.

Usage:
  dt-geo-converter convert [--wf <workflow_id> | --all] [--db <dbfile>] [--dir <directory>] [--update]`,
	Run: func(cmd *cobra.Command, args []string) {
		if !convertAll && workflowID == "" {
			fmt.Println("Either --wf or --all must be specified.")
			cmd.Help()
			os.Exit(1)
		}
		if convertUpdate && convertDir == "" {
			fmt.Println("The --dir flag is required when using --update.")
			cmd.Help()
			os.Exit(1)
		}
		commands.ConvertWorkflows(convertDBFile, convertDir, workflowID, convertAll, convertUpdate)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVar(&convertDBFile, "db", "./db.db", "Path to the database file (optional)")
	convertCmd.Flags().StringVar(&convertDir, "dir", "", "Directory containing CSV files or subdirectories with CSV files (required when using --update)")
	convertCmd.Flags().StringVar(&workflowID, "wf", "", "Workflow ID to process. Use --all to process all workflows.")
	convertCmd.Flags().BoolVar(&convertAll, "all", false, "Convert all workflows in the database")
	convertCmd.Flags().BoolVar(&convertUpdate, "update", false, "Reset and reinitialize the database before conversion")
}
