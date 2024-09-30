package main

import (
	"database/sql"
	"dt-geo-db/implicit"
	"encoding/csv"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
)

func main() {
	// Define command-line flags
	dbFile := flag.String("db", "./db.db", "Path to the database file")
	workflowID := flag.String("wf", "WF5201", "Workflow ID to process")
	workPackage := flag.String("wp", "wp5", "Work package identifier")
	resetDB := flag.Bool("rst", false, "Reset the database before starting")

	// Customize the usage message
	flag.Usage = func() {
		usageText := `
Usage: dt-geo-workflow-converter [options]

Options:
`
		fmt.Fprint(flag.CommandLine.Output(), usageText)
		flag.PrintDefaults()
		fmt.Println(`
Description:
  This program initializes a database, imports data from CSV files,
  generates a workflow graph based on the specified workflow ID,
  and saves the graph to files.
`)
	}

	// Parse the command-line flags
	flag.Parse()

	// Initialize the logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Check if the database exists
	dbExists := false
	if _, err := os.Stat(*dbFile); err == nil {
		dbExists = true
	} else if !os.IsNotExist(err) {
		// An error other than "file does not exist" occurred
		log.Fatalf("Error checking database file %s: %v", *dbFile, err)
	}

	// Reset the database if requested and it exists
	if *resetDB && dbExists {
		if err := resetDatabase(*dbFile); err != nil {
			log.Fatalf("Failed to reset database: %v", err)
		}
		dbExists = false // Database has been reset
	}

	// Open the database
	db, err := sql.Open("sqlite3", *dbFile)
	if err != nil {
		log.Fatalf("Failed to open database %s: %v", *dbFile, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize the database if it doesn't exist or was reset
	if !dbExists {
		if err := initializeDatabase(db, *workPackage); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		fmt.Println("Database created and CSV data imported successfully")
	} else {
		fmt.Println("Using existing database")
	}

	// Generate the workflow graph
	if err := processWorkflow(db, *workflowID); err != nil {
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
