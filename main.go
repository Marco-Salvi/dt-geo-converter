package main

import (
	"database/sql"
	"dt-geo-db/cwl"
	"dt-geo-db/implicit"
	"encoding/csv"
	"fmt"
	"github.com/dominikbraun/graph/draw"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
)

func main() {
	wp := "wp5"

	err := os.Remove("./db.db")
	if err != nil {
		log.Println(err)
	}

	// Open the database
	db, err := sql.Open("sqlite3", "./db.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables
	err = createTables(db)
	if err != nil {
		log.Fatal(err)
	}

	// Import data from CSV files
	err = importDataFromCSV(db, wp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database created and CSV data imported successfully")

	// get all workflows ids from the db
	wfs, err := GetWorkflows(db)

	for _, wf := range wfs {
		path := "workflows/" + wp + "/" + wf + "/"

		gr, err := cwl.GetWorkflowExecutionOrder(db, wf, path)
		if err != nil {
			log.Fatal(err)
		}

		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Create(path + "steps.dot")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		_ = draw.DOT(gr, file)
	}

	test, err := implicit.GetWorkflowGraph("WF5301", db)
	if err != nil {
		panic(err)
	}
	err = test.SaveToFile(db)
	if err != nil {
		panic(err)
	}
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

func GetWorkflows(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM WF")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []string
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, name)
	}

	return workflows, nil
}
