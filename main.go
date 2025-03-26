package main

import (
	"database/sql"
	"dt-geo-db/cwl"
	"dt-geo-db/implicit"
	"dt-geo-db/logger"
	"dt-geo-db/rocrate"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Process global debug flag by scanning os.Args.
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

	// Check for help flag first
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printMainHelp()
		return
	}

	// Ensure a subcommand is provided
	if len(os.Args) < 2 {
		fmt.Println("Expected 'convert' or 'generate-ro-crate' subcommands")
		fmt.Println("Run 'dt-geo-converter --help' for usage information")
		os.Exit(1)
	}

	// Initialize the logger based on the debug flag
	if debug {
		// Retain short file and date/time flags for debug mode
		// (Assuming the logger wrapper honors the standard log flags.)
		logger.DebugEnabled = true
		logger.Info("Debug mode enabled")
	}

	switch os.Args[1] {
	case "convert":
		convertCmd := flag.NewFlagSet("convert", flag.ExitOnError)
		dbFile := convertCmd.String("db", "./db.db", "Path to the database file, if it does not exist, it will be created")
		workflowID := convertCmd.String("wf", "WF5201", "Workflow ID to process")
		resetDB := convertCmd.Bool("rst", false, "Whether to reset the database before starting (necessary if changed the data in the csv)")

		convertCmd.Usage = func() {
			usageText := `
Usage: dt-geo-converter convert [options]

Options:
`
			fmt.Fprint(convertCmd.Output(), usageText)
			convertCmd.PrintDefaults()
			fmt.Println(`
Description:
  This subcommand initializes a database, imports data from CSV files and then generates an in-memory graph of the workflow that is used to generate CWL file descriptions, .dot files for the graphs, and an RO-Crate metadata template.
  Please review the generated logs carefully to see if the imported description has any issues.
`)
		}

		if err := convertCmd.Parse(os.Args[2:]); err != nil {
			logger.Fatal("Error parsing flags for 'convert':", err)
		}

		// Infer the work package from the workflow id
		workPackage := (*workflowID)[0:3]
		workPackage = strings.ReplaceAll(workPackage, "WF", "wp")
		logger.Info("Running conversion for workflow", *workflowID, "with work package", workPackage)
		runConvert(*dbFile, *workflowID, workPackage, *resetDB)

	case "generate-ro-crate":
		generateCmd := flag.NewFlagSet("generate-ro-crate", flag.ExitOnError)
		cwlFilePath := generateCmd.String("cwl", "", "Path to the CWL file")
		workflowName := generateCmd.String("name", "", "Name of the workflow")

		generateCmd.Usage = func() {
			usageText := `
Usage: dt-geo-converter generate-ro-crate -cwl <path_to_cwl_file> -name <workflow_name>

Options:
`
			fmt.Fprint(generateCmd.Output(), usageText)
			generateCmd.PrintDefaults()
			fmt.Println(`
Description:
  This subcommand generates an RO-Crate metadata package from a specified CWL file.
`)
		}

		if err := generateCmd.Parse(os.Args[2:]); err != nil {
			logger.Fatal("Error parsing flags for 'generate-ro-crate':", err)
		}

		if *cwlFilePath == "" {
			fmt.Println("Please provide a path to the CWL file using -cwl flag")
			generateCmd.Usage()
			os.Exit(1)
		}

		if *workflowName == "" {
			fmt.Println("Please provide the name of the workflow using -name flag")
			generateCmd.Usage()
			os.Exit(1)
		}

		logger.Info("Importing CWL file from", *cwlFilePath)
		cwlObj, err := cwl.ImportCWL(*cwlFilePath)
		if err != nil {
			logger.Fatal("Failed to import CWL file:", err)
		}

		logger.Info("Generating RO-Crate for workflow", *workflowName)
		ro, err := rocrate.GenerateRoCrate(*workflowName, cwlObj)
		if err != nil {
			logger.Fatal("Failed to generate RO-Crate:", err)
		}
		logger.Info("RO-Crate generated successfully")

		if err := ro.SaveToFile("ro-crate-metadata.json"); err != nil {
			logger.Fatal("Error saving RO-Crate to file:", err)
		}

	default:
		fmt.Printf("Unknown subcommand '%s'\n", os.Args[1])
		fmt.Println("Run 'dt-geo-converter --help' for usage information")
		os.Exit(1)
	}
}

// printMainHelp prints the help information for the base CLI
func printMainHelp() {
	helpText := `
Usage:
  dt-geo-converter <command> [options]

Commands:
  convert            Initialize database, import data, and generate CWL workflow graphs
  generate-ro-crate  Generate RO-Crate metadata package template from a CWL file

Options:
  --help, -h         Show this help message

For more information on a specific command, run:
  dt-geo-converter <command> --help
`
	fmt.Println(helpText)
}

// Function to handle the 'convert' subcommand logic
func runConvert(dbFile, workflowID, workPackage string, resetDB bool) {
	logger.Info("Starting conversion process...")
	// Check if the database exists
	dbExists := false
	if _, err := os.Stat(dbFile); err == nil {
		dbExists = true
		logger.Debug("Database file", dbFile, "exists")
	} else if !os.IsNotExist(err) {
		logger.Fatal("Error checking database file", dbFile, ":", err)
	}

	// Reset the database if requested and it exists
	if resetDB && dbExists {
		logger.Info("Reset flag detected. Resetting database", dbFile)
		if err := resetDatabase(dbFile); err != nil {
			logger.Fatal("Failed to reset database:", err)
		}
		dbExists = false // Database has been reset
	}

	// Open the database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		logger.Fatal("Failed to open database", dbFile, ":", err)
	}
	logger.Debug("Opened database", dbFile)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error closing database:", err)
		} else {
			logger.Debug("Database closed successfully")
		}
	}()

	// Initialize the database if it doesn't exist or was reset
	if !dbExists {
		logger.Info("Initializing database and importing CSV data")
		if err := initializeDatabase(db, workPackage); err != nil {
			logger.Fatal("Failed to initialize database:", err)
		}
		logger.Debug("Database created and CSV data imported successfully")
	} else {
		logger.Debug("Using existing database", dbFile)
	}

	// Generate the workflow graph
	logger.Info("Processing workflow graph for", workflowID)
	if err := processWorkflow(db, workflowID); err != nil {
		logger.Fatal("Failed to process workflow:", err)
	}
	logger.Info("Workflow graph generated and saved successfully")
	logger.Info("Conversion process completed successfully")
}

// resetDatabase removes the existing database file.
func resetDatabase(dbFile string) error {
	if err := os.Remove(dbFile); err != nil {
		return fmt.Errorf("error removing database file %s: %w", dbFile, err)
	}
	logger.Info("Database file", dbFile, "removed successfully")
	return nil
}

// initializeDatabase creates tables and imports data from CSV files.
func initializeDatabase(db *sql.DB, workPackage string) error {
	logger.Debug("Creating tables in the database")
	if err := createTables(db); err != nil {
		return fmt.Errorf("error creating tables: %w", err)
	}
	logger.Debug("Tables created successfully")

	logger.Info("Importing CSV data for work package", workPackage)
	if err := importDataFromCSV(db, workPackage); err != nil {
		return fmt.Errorf("error importing data from CSV for work package %s: %w", workPackage, err)
	}

	return nil
}

// processWorkflow generates the workflow graph and saves it to files.
func processWorkflow(db *sql.DB, workflowID string) error {
	logger.Info("Loading workflow graph for ID", workflowID)
	workflow, err := implicit.GetWorkflowGraph(workflowID, db)
	if err != nil {
		return fmt.Errorf("error getting workflow graph for ID %s: %w", workflowID, err)
	}

	logger.Debug("Saving workflow graph to file")
	if err := workflow.SaveToFile(db); err != nil {
		return fmt.Errorf("error saving workflow to file: %w", err)
	}

	return nil
}

func createTables(db *sql.DB) error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS WF (
	name TEXT PRIMARY KEY,
	description TEXT,
	author TEXT
);`,
		`CREATE TABLE IF NOT EXISTS WF_WF (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	FOREIGN KEY (id1) REFERENCES WF(name),
	FOREIGN KEY (id2) REFERENCES WF(name),
	PRIMARY KEY (id1, relationship_type, id2)
);`,
		`CREATE TABLE IF NOT EXISTS ST_ST (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	PRIMARY KEY (id1, relationship_type, id2)
);`,
		`CREATE TABLE IF NOT EXISTS SS_SS (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	PRIMARY KEY (id1, relationship_type, id2)
);`,
		`CREATE TABLE IF NOT EXISTS DT_DT (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	PRIMARY KEY (id1, relationship_type, id2)
);`,
		`CREATE TABLE IF NOT EXISTS ST_WF (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	FOREIGN KEY (id1) REFERENCES WF(name),
	PRIMARY KEY (id1, relationship_type, id2)
);`,
		`CREATE TABLE IF NOT EXISTS SS_ST (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	PRIMARY KEY (id1, relationship_type, id2)
);`,
		`CREATE TABLE IF NOT EXISTS DT_ST (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	PRIMARY KEY (id1, relationship_type, id2)
);`,
		`CREATE TABLE IF NOT EXISTS DT_SS (
	id1 TEXT,
	relationship_type TEXT NOT NULL,
	id2 TEXT,
	PRIMARY KEY (id1, relationship_type, id2)
);`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return err
		}
		logger.Debug("Executed schema:", schema)
	}

	return nil
}

func importDataFromCSV(db *sql.DB, dir string) error {
	dir = strings.ToLower(dir)
	relationships := map[string]string{
		"WF_WF": dir + "/wf_wf.csv",
		"ST_ST": dir + "/st_st.csv",
		"SS_SS": dir + "/ss_ss.csv",
		"DT_DT": dir + "/dt_dt.csv",
		"SS_ST": dir + "/ss_st.csv",
		"ST_WF": dir + "/st_wf.csv",
		"DT_ST": dir + "/dt_st.csv",
		"DT_SS": dir + "/dt_ss.csv",
	}

	for table, file := range relationships {
		logger.Info("Importing table from file:", file)
		if err := importFromCSV(db, table, file); err != nil {
			return err
		}
	}

	if err := insertWF(db, dir+"/wf.csv"); err != nil {
		return err
	}

	return nil
}

func importFromCSV(db *sql.DB, tableName, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	query := fmt.Sprintf("INSERT INTO %s (id1, relationship_type, id2) VALUES (?, ?, ?)", tableName)
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	logger.Debug("Inserting rows into table", tableName)
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		id1, id2, relType := row[0], row[2], row[1]

		// Clean the data
		if id1 == "" || id2 == "" || relType == "" {
			continue
		}
		id1 = strings.Trim(id1, " ")
		id2 = strings.Trim(id2, " ")
		relType = strings.Trim(relType, " ")

		if _, err = stmt.Exec(id1, relType, id2); err != nil {
			logger.Error("Error inserting row:", row)
			return err
		}
	}

	logger.Debug("Finished importing data for table", tableName)
	return nil
}

func insertWF(db *sql.DB, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	query := "INSERT INTO WF (name, description, author) VALUES (?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	logger.Debug("Inserting workflow data from", filename)
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		name, description, author := row[0], row[1], row[2]
		name = strings.Trim(name, " ")
		description = strings.Trim(description, " ")
		author = strings.Trim(author, " ")

		if _, err = stmt.Exec(name, description, author); err != nil {
			logger.Error("Error inserting row:", row)
			return err
		}
	}

	logger.Debug("Workflow data imported successfully from", filename)
	return nil
}
