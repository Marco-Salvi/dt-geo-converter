package main

import (
	"dt-geo-db/commands"
	"dt-geo-db/logger"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Process global debug flag from os.Args.
	debug := false
	newArgs := []string{os.Args[0]}
	for _, arg := range os.Args[1:] {
		if arg == "--debug" || arg == "-d" {
			debug = true
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	os.Args = newArgs

	if debug {
		logger.DebugEnabled = true
		logger.Info("Debug mode enabled")
	}

	// Check for help flag.
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printMainHelp()
		return
	}

	// Ensure a subcommand is provided.
	if len(os.Args) < 2 {
		fmt.Println("Expected a subcommand: init-db, convert, generate-ro-crate, list")
		fmt.Println("Run 'dt-geo-converter --help' for usage information")
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "init-db":
		initCmd := flag.NewFlagSet("init-db", flag.ExitOnError)
		dbFile := initCmd.String("db", "./db.db", "Path to the database file (optional)")
		dir := initCmd.String("dir", "", "Directory containing CSV files or subdirectories with CSV files (required)")
		update := initCmd.Bool("update", false, "Reset and reinitialize the database if it exists")
		initCmd.Usage = func() {
			fmt.Println(`
Usage: dt-geo-converter init-db --dir <directory> [--db <dbfile>] [--update]

Options:
  --db       Path to the database file. Default is "./db.db".
  --dir      Directory containing CSV files or subdirectories with CSV files. (Required)
  --update   Reset and reinitialize the database if it exists.
`)
		}
		if err := initCmd.Parse(os.Args[2:]); err != nil {
			logger.Fatal("Error parsing flags for 'init-db':", err)
		}
		if *dir == "" {
			fmt.Println("The --dir flag is required.")
			initCmd.Usage()
			os.Exit(1)
		}
		commands.InitDatabase(*dbFile, *dir, *update)

	case "convert":
		convertCmd := flag.NewFlagSet("convert", flag.ExitOnError)
		dbFile := convertCmd.String("db", "./db.db", "Path to the database file (optional)")
		dir := convertCmd.String("dir", "", "Directory containing CSV files or subdirectories with CSV files (required when using --update)")
		workflowID := convertCmd.String("wf", "", "Workflow ID to process. Use --all to process all workflows.")
		all := convertCmd.Bool("all", false, "Convert all workflows in the database")
		update := convertCmd.Bool("update", false, "Reset and reinitialize the database before conversion")
		convertCmd.Usage = func() {
			fmt.Println(`
Usage: dt-geo-converter convert [--wf <workflow_id> | --all] [--db <dbfile>] [--dir <directory>] [--update]

Options:
  --wf       Workflow ID to process (required if --all is not set)
  --all      Process all workflows in the database
  --db       Path to the database file. Default is "./db.db".
  --dir      Directory containing CSV files or subdirectories with CSV files. Required when using --update.
  --update   Reset and reinitialize the database before conversion.
`)
		}
		if err := convertCmd.Parse(os.Args[2:]); err != nil {
			logger.Fatal("Error parsing flags for 'convert':", err)
		}
		if !*all && *workflowID == "" {
			fmt.Println("Either --wf or --all must be specified.")
			convertCmd.Usage()
			os.Exit(1)
		}
		if *update && *dir == "" {
			fmt.Println("The --dir flag is required when using --update.")
			convertCmd.Usage()
			os.Exit(1)
		}
		commands.ConvertWorkflows(*dbFile, *dir, *workflowID, *all, *update)

	case "generate-ro-crate":
		roCmd := flag.NewFlagSet("generate-ro-crate", flag.ExitOnError)
		cwlPath := roCmd.String("cwl", "", "Path to the CWL file (required)")
		workflowName := roCmd.String("name", "", "Name of the workflow (required)")
		output := roCmd.String("output", "ro-crate-metadata.json", "Output file for the RO‑Crate metadata")
		roCmd.Usage = func() {
			fmt.Println(`
Usage: dt-geo-converter generate-ro-crate --cwl <path_to_cwl_file> --name <workflow_name> [--output <filename>]

Options:
  --cwl      Path to the CWL file (required)
  --name     Name of the workflow (required)
  --output   Output file for the RO‑Crate metadata. Default is "ro-crate-metadata.json".
`)
		}
		if err := roCmd.Parse(os.Args[2:]); err != nil {
			logger.Fatal("Error parsing flags for 'generate-ro-crate':", err)
		}
		if *cwlPath == "" || *workflowName == "" {
			fmt.Println("Both --cwl and --name flags are required.")
			roCmd.Usage()
			os.Exit(1)
		}
		commands.GenerateRoCrate(*cwlPath, *workflowName, *output)

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		dbFile := listCmd.String("db", "./db.db", "Path to the database file (optional)")
		listCmd.Usage = func() {
			fmt.Println(`
Usage: dt-geo-converter list [--db <dbfile>]

Options:
  --db       Path to the database file. Default is "./db.db".
`)
		}
		if err := listCmd.Parse(os.Args[2:]); err != nil {
			logger.Fatal("Error parsing flags for 'list':", err)
		}
		commands.ListWorkflows(*dbFile)

	default:
		fmt.Printf("Unknown subcommand '%s'\n", subcommand)
		printMainHelp()
		os.Exit(1)
	}
}

func printMainHelp() {
	helpText := `
Usage:
  dt-geo-converter <command> [options]

Commands:
  init-db           Initialize the database with CSV data
  convert           Convert workflow(s) from the database into CWL and generate workflow graphs
  generate-ro-crate Generate RO‑Crate metadata package from a CWL file
  list              List workflows available in the database

Options:
  --help, -h        Show this help message

For more information on a specific command, run:
  dt-geo-converter <command> --help
`
	fmt.Println(helpText)
}
