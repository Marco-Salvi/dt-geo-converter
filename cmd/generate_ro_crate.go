package cmd

import (
	"dt-geo-converter/commands"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cwlPath      string
	workflowName string
	outputFile   string
)

var generateRoCrateCmd = &cobra.Command{
	Use:   "generate-ro-crate",
	Short: "Generate RO‑Crate metadata package from a CWL file",
	Run: func(cmd *cobra.Command, args []string) {
		if cwlPath == "" || workflowName == "" {
			fmt.Println("Both --cwl and --name flags are required.")
			cmd.Help()
			os.Exit(1)
		}
		commands.GenerateRoCrate(cwlPath, workflowName, outputFile)
	},
}

func init() {
	rootCmd.AddCommand(generateRoCrateCmd)
	generateRoCrateCmd.Flags().StringVar(&cwlPath, "cwl", "", "Path to the CWL file (required)")
	generateRoCrateCmd.Flags().StringVar(&workflowName, "name", "", "Name of the workflow (required)")
	generateRoCrateCmd.Flags().StringVar(&outputFile, "output", "ro-crate-metadata.json", "Output file for the RO‑Crate metadata")
}
