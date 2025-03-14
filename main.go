package main

import (
	"database/sql"
	"dt-geo-db/cwl"
	"dt-geo-db/implicit"
	"dt-geo-db/rocrate"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
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

	switch os.Args[1] {
	case "convert":
		convertCmd := flag.NewFlagSet("convert", flag.ExitOnError)
		dbFile := convertCmd.String("db", "./db.db", "Path to the database file, if it does not exist, it will be created")
		workflowID := convertCmd.String("wf", "WF5201", "Workflow ID to process")
		workPackage := convertCmd.String("wp", "wp5", "Work package identifier")
		resetDB := convertCmd.Bool("rst", false, "Whether to reset the database before starting (necessary if changed the data in the csv)")

		// Customize the usage message for 'convert'
		convertCmd.Usage = func() {
			usageText := `
Usage: dt-geo-converter convert [options]

Options:
`
			fmt.Fprint(convertCmd.Output(), usageText)
			convertCmd.PrintDefaults()
			fmt.Println(`
Description:
  This subcommand initializes a database, imports data from CSV files and it then generates an in-memoy graph of the workflow that is used to generate CWL files description, .dot files for the graphs and a ro-crate-metadata template.
  The tool has logging to warn the user when the imported description from the spreadsheets has some problems. It is IMPORTANT to look at the logging to see what is wrong.
  It will generate separate .cwl and .dot files for each step of the original workflow.
  The generated CWL files will probably have to be reviewed to ensure a correct workflow representation. In general the syntax should already be correct.
  The .dot files can be visualized using a web tool like https://dreampuf.github.io/GraphvizOnline. Can be very useful to understand the CWL workflow even if it is not correct.
  The generated ro-crate-metadata.json file will include the necessary objects that are used in the CWL workflow, but will need to be reviewed to add the necesary metadata.
`)
		}

		// Initialize the logger
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		// Parse flags for 'convert' subcommand
		err := convertCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Error parsing flags for 'convert': %v", err)
		}

		// Rest of your 'convert' logic
		runConvert(*dbFile, *workflowID, *workPackage, *resetDB)

	case "generate-ro-crate":
		generateCmd := flag.NewFlagSet("generate-ro-crate", flag.ExitOnError)
		cwlFilePath := generateCmd.String("cwl", "", "Path to the CWL file")
		workflowName := generateCmd.String("name", "", "Name of the workflow")

		// Customize the usage message for 'generate-ro-crate'
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

		// Parse flags for 'generate-ro-crate' subcommand
		err := generateCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Error parsing flags for 'generate-ro-crate': %v", err)
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

		// Implement the logic to generate RO-Crate from the CWL file
		cwl, err := cwl.ImportCWL(*cwlFilePath)
		if err != nil {
			log.Fatalf("Failed to import CWL file: %v", err)
		}

		ro, err := rocrate.GenerateRoCrate(*workflowName, cwl)
		if err != nil {
			log.Fatalf("Failed to generate RO-Crate: %v", err)
		}
		fmt.Println("RO-Crate generated successfully")

		err = ro.SaveToFile("ro-crate-metadata.json")
		if err != nil {
			log.Fatalf("Error saving RO-Crate to file: %v", err)
		}

	default:
		fmt.Printf("Unknown subcommand '%s'\n", os.Args[1])
		fmt.Println("Run 'dt-geo-workflow-converter --help' for usage information")
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
	// Check if the database exists
	dbExists := false
	if _, err := os.Stat(dbFile); err == nil {
		dbExists = true
	} else if !os.IsNotExist(err) {
		// An error other than "file does not exist" occurred
		log.Fatalf("Error checking database file %s: %v", dbFile, err)
	}

	// Reset the database if requested and it exists
	if resetDB && dbExists {
		if err := resetDatabase(dbFile); err != nil {
			log.Fatalf("Failed to reset database: %v", err)
		}
		dbExists = false // Database has been reset
	}

	// Open the database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("Failed to open database %s: %v", dbFile, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize the database if it doesn't exist or was reset
	if !dbExists {
		if err := initializeDatabase(db, workPackage); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		fmt.Println("Database created and CSV data imported successfully")
	} else {
		fmt.Println("Using existing database")
	}

	// Generate the workflow graph
	if err := processWorkflow(db, workflowID); err != nil {
		log.Fatalf("Failed to process workflow: %v", err)
	}

	fmt.Println("Workflow graph generated and saved successfully")
}

// resetDatabase removes the existing database file.
func resetDatabase(dbFile string) error {
	if err := os.Remove(dbFile); err != nil {
		return fmt.Errorf("error removing database file %s: %w", dbFile, err)
	}
	log.Printf("Database file %s removed successfully", dbFile)
	return nil
}

// initializeDatabase creates tables and imports data from CSV files.
func initializeDatabase(db *sql.DB, workPackage string) error {
	// Create tables
	if err := createTables(db); err != nil {
		return fmt.Errorf("error creating tables: %w", err)
	}

	// Import data from CSV files
	if err := importDataFromCSV(db, workPackage); err != nil {
		return fmt.Errorf("error importing data from CSV for work package %s: %w", workPackage, err)
	}

	return nil
}

// processWorkflow generates the workflow graph and saves it to files.
func processWorkflow(db *sql.DB, workflowID string) error {
	workflow, err := implicit.GetWorkflowGraph(workflowID, db)
	if err != nil {
		return fmt.Errorf("error getting workflow graph for ID %s: %w", workflowID, err)
	}

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
		_, err := db.Exec(schema)
		if err != nil {
			return err
		}
	}

	return nil
}

func importDataFromCSV(db *sql.DB, folder string) error {
	// Import relationship data using the generic function
	relationships := map[string]string{
		"WF_WF": folder + "/wf_wf.csv",
		"ST_ST": folder + "/st_st.csv",
		"SS_SS": folder + "/ss_ss.csv",
		"DT_DT": folder + "/dt_dt.csv",
		"SS_ST": folder + "/ss_st.csv",
		"ST_WF": folder + "/st_wf.csv",
		"DT_ST": folder + "/dt_st.csv",
		"DT_SS": folder + "/dt_ss.csv",
	}

	for table, file := range relationships {
		log.Println("importing table: " + file)
		err := importFromCSV(db, table, file)
		if err != nil {
			return err
		}
	}

	err := insertWF(db, folder+"/wf.csv")
	if err != nil {
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

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		id1, id2, relType := row[0], row[2], row[1]

		// clean the data
		if id1 == "" || id2 == "" || relType == "" {
			continue
		}
		// remove trailing spaces
		id1 = strings.Trim(id1, " ")
		id2 = strings.Trim(id2, " ")
		relType = strings.Trim(relType, " ")

		_, err = stmt.Exec(id1, relType, id2)
		if err != nil {
			fmt.Println("Error inserting row:", row)
			return err
		}
	}

	return nil
}

// insert wf data
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

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		name, description, author := row[0], row[1], row[2]

		// remove trailing spaces
		name = strings.Trim(name, " ")
		description = strings.Trim(description, " ")
		author = strings.Trim(author, " ")

		_, err = stmt.Exec(name, description, author)
		if err != nil {
			fmt.Println("Error inserting row:", row)
			return err
		}
	}

	return nil
}
