package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version holds the current version of the tool. It is overridden at build time.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of dt-geo-converter",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
