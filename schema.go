package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// DatasetUsage represents a dataset and its specific role in a step or software system.
type DatasetUsage struct {
	DatasetName string `json:"dataset_name"`
	Role        string `json:"role"` // The role of the dataset (e.g., "input", "output", etc.)
}

// SoftwareSystem represents a software system that can be used within a step.
type SoftwareSystem struct {
	Name     string         `json:"name"`
	Datasets []DatasetUsage `json:"datasets,omitempty"` // Datasets used by the software system with roles
}

// Step represents a step in a workflow.
type Step struct {
	Name     string           `json:"name"`
	Software []SoftwareSystem `json:"software,omitempty"` // Software systems used in this step
	Datasets []DatasetUsage   `json:"datasets,omitempty"` // Datasets used by the step with roles
}

// Workflow represents the entire workflow with a collection of steps.
type Workflow struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

func GetWorkflowByName(db *sql.DB, workflowName string) (*Workflow, error) {
	// Fetch the workflow
	var wf Workflow
	wf.Name = workflowName

	// Fetch steps associated with the workflow
	stepsMap := make(map[string]*Step)
	stepRows, err := db.Query(`
        SELECT id1, relationship_type FROM ST_WF WHERE id2 = ?
    `, workflowName)
	if err != nil {
		return nil, fmt.Errorf("failed to query ST_WF: %v", err)
	}
	defer stepRows.Close()

	for stepRows.Next() {
		var stepName, relationshipType string
		if err := stepRows.Scan(&stepName, &relationshipType); err != nil {
			return nil, fmt.Errorf("failed to scan step row: %v", err)
		}

		step := &Step{Name: stepName}
		stepsMap[stepName] = step
		wf.Steps = append(wf.Steps, *step)
	}

	if len(wf.Steps) == 0 {
		log.Println("No steps found for workflow:", workflowName)
	}

	// Fetch software systems for each step
	for i := range wf.Steps {
		step := &wf.Steps[i]
		softwareRows, err := db.Query(`
            SELECT id1, relationship_type FROM SS_ST WHERE id2 = ?
        `, step.Name)

		if err != nil {
			return nil, fmt.Errorf("failed to query SS_ST: %v", err)
		}

		for softwareRows.Next() {
			var softwareName, relationshipType string
			if err := softwareRows.Scan(&softwareName, &relationshipType); err != nil {
				return nil, fmt.Errorf("failed to scan software row: %v", err)
			}

			software := SoftwareSystem{Name: softwareName}

			// Fetch datasets associated with the software system
			datasetRows, err := db.Query(`
                SELECT id1, relationship_type FROM DT_SS WHERE id2 = ?
            `, softwareName)
			if err != nil {
				return nil, fmt.Errorf("failed to query DT_SS: %v", err)
			}

			for datasetRows.Next() {
				var datasetName, relationshipType string
				if err := datasetRows.Scan(&datasetName, &relationshipType); err != nil {
					return nil, fmt.Errorf("failed to scan dataset row: %v", err)
				}

				datasetUsage := DatasetUsage{
					DatasetName: datasetName,
					Role:        relationshipType,
				}
				software.Datasets = append(software.Datasets, datasetUsage)
			}
			datasetRows.Close()

			step.Software = append(step.Software, software)
		}
		softwareRows.Close()

		// Fetch datasets directly associated with the step
		stepDatasetRows, err := db.Query(`
            SELECT id2, relationship_type FROM DT_ST WHERE id1 = ?
        `, step.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to query DT_ST: %v", err)
		}

		for stepDatasetRows.Next() {
			var datasetName, relationshipType string
			if err := stepDatasetRows.Scan(&datasetName, &relationshipType); err != nil {
				return nil, fmt.Errorf("failed to scan step dataset row: %v", err)
			}

			datasetUsage := DatasetUsage{
				DatasetName: datasetName,
				Role:        relationshipType,
			}
			step.Datasets = append(step.Datasets, datasetUsage)
		}
		stepDatasetRows.Close()
	}

	if len(wf.Steps) == 0 {
		log.Println("No steps found for workflow:", workflowName)
	}

	return &wf, nil
}
