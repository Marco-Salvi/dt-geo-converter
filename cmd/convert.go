package cmd

import (
	"dt-geo-converter/commands"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	convertDBFile string
	workflowID    string
	convertAll    bool
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert workflow(s) from the database into CWL and generate workflow graphs",
	Run: func(cmd *cobra.Command, args []string) {
		if !convertAll && workflowID == "" {
			fmt.Println("Either --wf or --all must be specified.")
			cmd.Help()
			os.Exit(1)
		}
		commands.ConvertWorkflows(convertDBFile, workflowID, convertAll)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVar(&convertDBFile, "db", "./db.db", "Path to the database file (optional)")
	convertCmd.Flags().StringVar(&workflowID, "wf", "", "Workflow ID to process. Use --all to process all workflows.")
	convertCmd.Flags().BoolVar(&convertAll, "all", false, "Convert all workflows in the database")
}
