package cwl

import (
	"dt-geo-db/workflow"
	"fmt"
	"log"
)

func ConvertWorkflowToCWL(workflow *workflow.Workflow) Cwl {
	cwl := Cwl{
		CWLVersion: "v1.2",
		Class:      "Workflow",
		Inputs:     make(map[string]IOType),
		Outputs:    make(map[string]Output),
		Steps:      make(map[string]Step),
	}

	// Maps to track dataset roles and relationships
	datasetProducers := make(map[string]string)   // dataset -> producing step
	datasetConsumers := make(map[string][]string) // dataset -> list of consuming steps
	datasetTypes := make(map[string]IOType)       // dataset -> type (assumed Directory)
	datasetUpdated := make(map[string]bool)       // dataset -> is updated

	// Build step order mapping based on dependencies
	stepOrder := determineStepOrder(workflow, datasetProducers, datasetConsumers)

	// First pass: Collect all dataset information
	for _, step := range workflow.Steps {
		for _, software := range step.Software {
			softwareName := software.Name
			if softwareName == "" {
				softwareName = step.Name
			}

			for _, datasetUsage := range software.Datasets {
				datasetName := datasetUsage.DatasetName
				role := datasetUsage.Role
				datasetTypes[datasetName] = Directory // Assuming all datasets are Directory

				// Record consumers
				if role == "is input to" || role == "is updated by" {
					datasetConsumers[datasetName] = append(datasetConsumers[datasetName], softwareName)
				}

				// Record producers
				if role == "is output from" || role == "is the output from" {
					datasetProducers[datasetName] = softwareName
				}

				// For updated datasets, they are both inputs and outputs
				if role == "is updated by" {
					datasetUpdated[datasetName] = true
					datasetProducers[datasetName] = softwareName
					datasetConsumers[datasetName] = append(datasetConsumers[datasetName], softwareName)
				}
			}
		}
	}

	// Second pass: Determine workflow inputs
	for datasetName := range datasetConsumers {
		_, isProduced := datasetProducers[datasetName]
		isUpdated := datasetUpdated[datasetName]
		if !isProduced || isUpdated {
			cwl.Inputs[datasetName] = datasetTypes[datasetName]
		}
	}

	// Determine workflow outputs
	for datasetName := range datasetProducers {
		cwl.Outputs[datasetName] = Output{
			Type:         datasetTypes[datasetName],
			OutputSource: fmt.Sprintf("%s/%s", datasetProducers[datasetName], datasetName),
		}
	}

	// Third pass: Build steps
	for _, step := range workflow.Steps {
		for _, software := range step.Software {
			softwareName := software.Name
			if softwareName == "" {
				softwareName = step.Name
			}

			run := Run{
				Class:   "Operation",
				Inputs:  make(map[string]string),
				Outputs: make(map[string]string),
			}
			stepIn := make(map[string]string)
			var stepOut []string

			for _, datasetUsage := range software.Datasets {
				datasetName := datasetUsage.DatasetName
				role := datasetUsage.Role

				if role == "is updated by" {
					log.Println("WARNING: this workflow contains a dataset that is updated by a software step.\nAt this time, this converter does not support workflows with datasets that are updated by software steps.\nThe output CWL WILL NOT be correct, but it might be a starting point to model the workflow.")
				}

				// Determine if dataset is input, output, or updated
				isInput := role == "is input to" || role == "is updated by"
				isOutput := role == "is output from" || role == "is the output from" || role == "is updated by"

				if isInput {
					// Add to run inputs
					run.Inputs[datasetName] = string(datasetTypes[datasetName])

					// Determine source
					if producer, exists := datasetProducers[datasetName]; exists && producer != softwareName && stepOrder[producer] < stepOrder[softwareName] {
						// Dataset is produced by a previous step
						stepIn[datasetName] = fmt.Sprintf("%s/%s", producer, datasetName)
					} else {
						// Dataset is a workflow input
						stepIn[datasetName] = datasetName
					}
				}

				if isOutput {
					// Add to run outputs
					run.Outputs[datasetName] = string(datasetTypes[datasetName])

					// Add to step 'out'
					if !contains(stepOut, datasetName) {
						stepOut = append(stepOut, datasetName)
					}
				}
			}

			cwl.Steps[softwareName] = Step{
				Run: run,
				In:  stepIn,
				Out: stepOut,
			}
		}
	}

	return cwl
}

func determineStepOrder(workflow *workflow.Workflow, datasetProducers map[string]string, datasetConsumers map[string][]string) map[string]int {
	// Build initial step order mapping
	stepOrder := make(map[string]int)
	dependencies := make(map[string]map[string]bool)

	// Initialize dependencies map
	for _, step := range workflow.Steps {
		for _, software := range step.Software {
			softwareName := software.Name
			if softwareName == "" {
				softwareName = step.Name
			}
			dependencies[softwareName] = make(map[string]bool)
		}
	}

	// Build dependencies based on dataset usage
	for datasetName, consumers := range datasetConsumers {
		producer, hasProducer := datasetProducers[datasetName]
		if hasProducer {
			for _, consumer := range consumers {
				if producer != consumer {
					dependencies[consumer][producer] = true
				}
			}
		}
	}

	// Perform topological sort
	visited := make(map[string]bool)
	tempMarked := make(map[string]bool)
	order := []string{}
	hasCycle := false

	var visit func(string) bool
	visit = func(n string) bool {
		if tempMarked[n] {
			hasCycle = true
			return false // Cycle detected
		}
		if !visited[n] {
			tempMarked[n] = true
			for dep := range dependencies[n] {
				if !visit(dep) {
					return false
				}
			}
			tempMarked[n] = false
			visited[n] = true
			order = append(order, n)
		}
		return true
	}

	for stepName := range dependencies {
		if !visited[stepName] {
			if !visit(stepName) {
				// Handle cycle detection if necessary
				if hasCycle {
					log.Println("Cycle detected in workflow dependencies")
					return nil
				}
			}
		}
	}

	// Assign order
	for idx, stepName := range order {
		stepOrder[stepName] = idx
	}

	return stepOrder
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, elem := range slice {
		if elem == item {
			return true
		}
	}
	return false
}
