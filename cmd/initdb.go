package cmd

import (
	"dt-geo-converter/commands"
	"dt-geo-converter/logger"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	initDBFile string
	initDir    string
	initUpdate bool
	initRemote string
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

		if initRemote != "" {

			remotes, err := parseRemoteFlag()
			if err != nil {
				fmt.Printf("Failed to parse the --remote flag: %v\n", err)
				os.Exit(1)
			}
			initDir, err = commands.DownloadRemoteSheets(remotes)
			if err != nil {
				fmt.Printf("Failed to download remote sheets: %v\n", err)
				os.Exit(1)
			}
			// Ensure we clean up the temporary directory when done.
			defer func() {
				if err := os.RemoveAll(initDir); err != nil {
					logger.Error("Warning: failed to clean up temporary directory %s: %v", initDir, err)
				}
				logger.Debug("Successfully removed temp dir")
			}()
		} else if initDir == "" {
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
	// Dynamically retrieve available work packages for help text.
	availableWPs := commands.GetAvailableWPs()
	helpMsg := fmt.Sprintf(
		"Initialize the database using the remote spreadsheets in the DT-GEO Google Drive. "+
			"Allowed values: 'all' or a comma-separated list from [%s]. Use '--remote all' to load all available work packages.",
		strings.Join(availableWPs, ", "))

	rootCmd.AddCommand(initDBCmd)
	// Define and document all flags
	initDBCmd.Flags().StringVar(&initDBFile, "db", "./db.db", "Path to the database file")
	initDBCmd.Flags().StringVar(&initDir, "dir", "", "Directory containing CSV files or subdirectories with CSV files")
	initDBCmd.Flags().BoolVar(&initUpdate, "update", false, "Reset and reinitialize the database if it exists")
	initDBCmd.Flags().StringVar(&initRemote, "remote", "", helpMsg)

	// Mark flags as mutually exclusive
	initDBCmd.MarkFlagsMutuallyExclusive("dir", "remote")
}

func parseRemoteFlag() ([]string, error) {
	// Clean up the remote flag value and split it
	remoteVal := strings.TrimSpace(initRemote)
	var remotes []string

	// If the value is "all" (case-insensitive), use that directly.
	if strings.EqualFold(remoteVal, "all") {
		remotes = []string{"all"}
	} else {
		// Split on commas and trim spaces from each remote key.
		parts := strings.Split(remoteVal, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				remotes = append(remotes, trimmed)
			}
		}
		if len(remotes) == 0 {
			return nil, fmt.Errorf("error: invalid --remote flag value")
		}
		// Retrieve available work packages dynamically.
		availableWPs := commands.GetAvailableWPs()
		allowed := make(map[string]bool)
		for _, wp := range availableWPs {
			allowed[wp] = true
		}
		// Validate each provided remote value.
		for _, r := range remotes {
			if !allowed[r] {
				return nil, fmt.Errorf("error: unknown remote value '%s'. Allowed values are 'all' or one of [%s]", r, strings.Join(availableWPs, ", "))
			}
		}
	}
	return remotes, nil
}
