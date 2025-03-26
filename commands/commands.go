package commands

import (
	"database/sql"
	"dt-geo-db/cwl"
	"dt-geo-db/implicit"
	"dt-geo-db/logger"
	"dt-geo-db/rocrate"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// InitDatabase initializes (or re‑initializes) the database using CSV files.
// The 'dir' parameter must point to a folder that either contains the CSV files directly,
// or contains subdirectories where each has the expected CSV files.
func InitDatabase(dbFile, dir string, update bool) {
	// Check if the DB exists.
	dbExists := false
	if _, err := os.Stat(dbFile); err == nil {
		dbExists = true
	}

	// If update is requested, remove the existing database.
	if update && dbExists {
		if err := resetDatabase(dbFile); err != nil {
			logger.Fatal("Failed to reset database:", err)
		}
		dbExists = false
	}

	// Open the database.
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		logger.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create tables.
	if err := createTables(db); err != nil {
		logger.Fatal("Failed to create tables:", err)
	}

	// Determine if the provided directory contains CSV files directly
	// (by checking for a known file, e.g., "wf.csv").
	if _, err := os.Stat(filepath.Join(dir, "wf.csv")); err == nil {
		logger.Info("Importing CSV data from directory:", dir)
		if err := importDataFromCSV(db, dir); err != nil {
			logger.Fatal("Failed to import CSV data:", err)
		}
	} else {
		// Otherwise, assume it contains subdirectories.
		entries, err := os.ReadDir(dir)
		if err != nil {
			logger.Fatal("Error reading directory:", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(dir, entry.Name())
				logger.Info("Importing CSV data from subdirectory:", subDir)
				if err := importDataFromCSV(db, subDir); err != nil {
					logger.Error("Failed to import CSV data from", subDir, ":", err)
				}
			}
		}
	}
	logger.Info("Database initialized successfully.")
}

// ConvertWorkflows converts one or all workflows from the database.
// If 'update' is true, the database is re‑initialized using the CSV data from 'dir' before conversion.
func ConvertWorkflows(dbFile, dir, workflowID string, all bool, update bool) {
	// If update is requested, re‑initialize the database.
	if update {
		InitDatabase(dbFile, dir, true)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		logger.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	if all {
		// Query all workflow IDs.
		rows, err := db.Query("SELECT name FROM WF")
		if err != nil {
			logger.Fatal("Failed to query workflows:", err)
		}
		defer rows.Close()

		var workflows []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				logger.Error("Error scanning workflow:", err)
				continue
			}
			workflows = append(workflows, name)
		}

		for _, wf := range workflows {
			logger.Info("Processing workflow", wf)
			if err := processWorkflow(db, wf); err != nil {
				logger.Error("Failed to process workflow", wf, ":", err)
			} else {
				logger.Info("Workflow", wf, "processed successfully.")
			}
		}
		logger.Info("All workflows processed.")
	} else {
		if workflowID == "" {
			logger.Fatal("Workflow ID must be provided if not processing all workflows.")
		}
		logger.Info("Processing workflow", workflowID)
		if err := processWorkflow(db, workflowID); err != nil {
			logger.Fatal("Failed to process workflow:", err)
		}
		logger.Info("Workflow processed successfully.")
	}
}

// GenerateRoCrate generates an RO‑Crate metadata package from the specified CWL file.
func GenerateRoCrate(cwlPath, workflowName, output string) {
	logger.Info("Importing CWL file from", cwlPath)
	cwlObj, err := cwl.ImportCWL(cwlPath)
	if err != nil {
		logger.Fatal("Failed to import CWL file:", err)
	}

	logger.Info("Generating RO‑Crate for workflow", workflowName)
	ro, err := rocrate.GenerateRoCrate(workflowName, cwlObj)
	if err != nil {
		logger.Fatal("Failed to generate RO‑Crate:", err)
	}

	if err := ro.SaveToFile(output); err != nil {
		logger.Fatal("Error saving RO‑Crate to file:", err)
	}
	logger.Info("RO‑Crate generated and saved to", output)
}

// ListWorkflows prints out all workflows stored in the database.
func ListWorkflows(dbFile string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		logger.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT name, description, author FROM WF")
	if err != nil {
		logger.Fatal("Failed to query workflows:", err)
	}
	defer rows.Close()

	fmt.Println("Workflows in database:")
	for rows.Next() {
		var name, description, author string
		if err := rows.Scan(&name, &description, &author); err != nil {
			logger.Error("Error scanning workflow:", err)
			continue
		}
		fmt.Printf("ID: %s, Description: %s, Author: %s\n", name, description, author)
	}
}

// resetDatabase removes the existing database file.
func resetDatabase(dbFile string) error {
	if err := os.Remove(dbFile); err != nil {
		return fmt.Errorf("error removing database file %s: %w", dbFile, err)
	}
	logger.Info("Database file", dbFile, "removed successfully")
	return nil
}

// createTables creates the required tables in the database.
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

// importDataFromCSV imports CSV data from a given directory.
func importDataFromCSV(db *sql.DB, dir string) error {
	// Ensure the directory string is in lowercase.
	dir = strings.ToLower(dir)
	relationships := map[string]string{
		"WF_WF": filepath.Join(dir, "wf_wf.csv"),
		"ST_ST": filepath.Join(dir, "st_st.csv"),
		"SS_SS": filepath.Join(dir, "ss_ss.csv"),
		"DT_DT": filepath.Join(dir, "dt_dt.csv"),
		"SS_ST": filepath.Join(dir, "ss_st.csv"),
		"ST_WF": filepath.Join(dir, "st_wf.csv"),
		"DT_ST": filepath.Join(dir, "dt_st.csv"),
		"DT_SS": filepath.Join(dir, "dt_ss.csv"),
	}

	for table, file := range relationships {
		logger.Info("Importing table from file:", file)
		if err := importFromCSV(db, table, file); err != nil {
			return err
		}
	}

	if err := insertWF(db, filepath.Join(dir, "wf.csv")); err != nil {
		return err
	}

	return nil
}

// importFromCSV reads a CSV file and imports its data into the specified table.
func importFromCSV(db *sql.DB, tableName, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

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
		if len(row) < 3 {
			continue
		}
		id1 := strings.TrimSpace(row[0])
		relType := strings.TrimSpace(row[1])
		id2 := strings.TrimSpace(row[2])

		if id1 == "" || relType == "" || id2 == "" {
			continue
		}
		if _, err = stmt.Exec(id1, relType, id2); err != nil {
			logger.Error("Error inserting row:", row)
			return err
		}
	}
	logger.Debug("Finished importing data for table", tableName)
	return nil
}

// insertWF imports workflow data from a CSV file into the WF table.
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
		if len(row) < 3 {
			continue
		}
		name := strings.TrimSpace(row[0])
		description := strings.TrimSpace(row[1])
		author := strings.TrimSpace(row[2])

		if _, err = stmt.Exec(name, description, author); err != nil {
			logger.Error("Error inserting row:", row)
			return err
		}
	}
	logger.Debug("Workflow data imported successfully from", filename)
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
