package cmd

import (
	"dt-geo-converter/commands"
	"dt-geo-converter/logger"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	initDBFile string
	initDir    string
	initUpdate bool
	initRemote bool
)

var initDBCmd = &cobra.Command{
	Use:   "init-db",
	Short: "Initialize the database with CSV data",
	Long:  "Initialize the database with CSV data from local directory or remote Google Sheets",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if database file already exists and handle accordingly
		if _, err := os.Stat(initDBFile); err == nil && !initUpdate {
			fmt.Printf("Database file already exists at %s. Use --update flag to overwrite it.\n", initDBFile)
			os.Exit(1)
		}

		if initRemote { // use the remote sheets
			var err error
			initDir, err = commands.DownloadRemoteSheets()
			if err != nil {
				fmt.Printf("Failed to download remote sheets: %v\n", err)
				os.Exit(1)
			}
			// Make sure we clean up the temp directory when done
			defer func() {
				if err := os.RemoveAll(initDir); err != nil {
					logger.Error("Warning: failed to clean up temporary directory %s: %v", initDir, err)
				}
				logger.Debug("Successfully removed temp dir")
			}()
		} else if initDir == "" { // if remote is not specified, the dir must be specified
			fmt.Println("The --dir flag is required if not using --remote.")
			_ = cmd.Help()
			os.Exit(1)
		}

		// Verify the directory exists
		if _, err := os.Stat(initDir); os.IsNotExist(err) {
			fmt.Printf("Directory does not exist: %s\n", initDir)
			os.Exit(1)
		}

		err := commands.InitDatabase(initDBFile, initDir, initUpdate)
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Database successfully initialized at: %s\n", initDBFile)
	},
}

func init() {
	rootCmd.AddCommand(initDBCmd)

	// Define and document all flags
	initDBCmd.Flags().StringVar(&initDBFile, "db", "./db.db", "Path to the database file")
	initDBCmd.Flags().StringVar(&initDir, "dir", "", "Directory containing CSV files or subdirectories with CSV files")
	initDBCmd.Flags().BoolVar(&initUpdate, "update", false, "Reset and reinitialize the database if it exists")
	initDBCmd.Flags().BoolVar(&initRemote, "remote", false, "Initialize the database using the remote spreadsheets in the DT-GEO Google Drive")

	// Mark flags as required or provide usage examples
	initDBCmd.MarkFlagsMutuallyExclusive("dir", "remote")
}
