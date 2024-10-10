package rocrate

import (
	"database/sql"
	"dt-geo-db/cwl"
	"dt-geo-db/orms"
	"strings"
)

func WorkflowToRoCrate(wf string, cwl cwl.Cwl, db *sql.DB) (RoCrate, error) {
	datasets, err := orms.GetDTsForWF(db, wf)
	if err != nil {
		return RoCrate{}, err
	}

	steps, err := orms.GetSTsForWF(db, wf)
	if err != nil {
		return RoCrate{}, err
	}

	var graph []any

	// metadata object
	graph = append(graph, Metadata{
		ID:         "ro-crate-metadata.json",
		Type:       "CreativeWork",
		ConformsTo: []IDRef{{"https://w3id.org/ro/crate/1.1"}, {"https://w3id.org/workflowhub/workflow-ro-crate/1.0"}},
		About:      IDRef{"./"},
	})

	// main workflow
	var workflowHasPart []IDRef
	workflowHasPart = append(workflowHasPart, IDRef{wf + ".cwl"})

	for _, dataset := range datasets {
		workflowHasPart = append(workflowHasPart, IDRef{dataset.ID})
	}
	for _, step := range steps {
		workflowHasPart = append(workflowHasPart, IDRef{step.ID + ".cwl"})
	}

	graph = append(graph, Workflow{
		ID:          "./",
		Type:        "Dataset",
		Name:        "TODO",
		Description: "TODO",
		License:     "TODO",
		Author:      IDRef{"TODO"},
		ConformsTo:  []IDRef{{"https://w3id.org/ro/wfrun/process/0.4"}, {"https://w3id.org/ro/wfrun/workflow/0.4"}, {"https://w3id.org/workflowhub/workflow-ro-crate/1.0"}},
		HasPart:     workflowHasPart,
		MainEntity:  IDRef{wf + ".cwl"},
	})

	graph = append(graph, CreativeWork{
		ID:      "https://w3id.org/ro/wfrun/process/0.4",
		Type:    "CreativeWork",
		Name:    "Process Run Crate",
		Version: "0.4",
	})

	graph = append(graph, CreativeWork{
		ID:      "https://w3id.org/ro/wfrun/workflow/0.4",
		Type:    "CreativeWork",
		Name:    "Workflow Run Crate",
		Version: "0.4",
	})

	graph = append(graph, CreativeWork{
		ID:      "https://w3id.org/workflowhub/workflow-ro-crate/1.0",
		Type:    "CreativeWork",
		Name:    "Workflow RO-Crate",
		Version: "1.0",
	})

	var workflowInputs []IDRef
	var workflowOutputs []IDRef
	for s := range cwl.Inputs {
		workflowInputs = append(workflowInputs, IDRef{"#" + s + "-param"})
	}
	for s := range cwl.Outputs {
		workflowOutputs = append(workflowOutputs, IDRef{"#" + s + "-param"})
	}

	graph = append(graph, ComputationalWorkflowFile{
		ID:                  wf + ".cwl",
		Type:                []string{"File", "SoftwareSourceCode", "ComputationalWorkflow"},
		Name:                "TODO",
		Author:              IDRef{"TODO"},
		Creator:             IDRef{"TODO"},
		ProgrammingLanguage: IDRef{"https://about.workflowhub.eu/Workflow-RO-Crate/#cwl"},
		Input:               workflowInputs,
		Output:              workflowOutputs,
	})

	// Formal parameters
	for _, param := range workflowInputs {
		graph = append(graph, FormalParameter{
			ID:             param.ID,
			Type:           "FormalParameter",
			AdditionalType: "Dataset",
			ConformsTo:     IDRef{"https://bioschemas.org/profiles/FormalParameter/1.0-RELEASE"},
			Description:    "TODO",
			WorkExample:    IDRef{strings.ReplaceAll(strings.ReplaceAll(param.ID, "#", ""), "-param", "")},
			Name:           "TODO",
			ValueRequired:  true,
		})
	}
	for _, param := range workflowOutputs {
		graph = append(graph, FormalParameter{
			ID:             param.ID,
			Type:           "FormalParameter",
			AdditionalType: "Dataset",
			ConformsTo:     IDRef{"https://bioschemas.org/profiles/FormalParameter/1.0-RELEASE"},
			Description:    "TODO",
			WorkExample:    IDRef{strings.ReplaceAll(strings.ReplaceAll(param.ID, "#", ""), "-param", "")},
			Name:           "TODO",
			ValueRequired:  true,
		})
	}

	graph = append(graph, ComputerLanguage{
		ID:         "https://w3id.org/workflowhub/workflow-ro-crate#cwl",
		Type:       "ComputerLanguage",
		Identifier: "https://www.commonwl.org/",
		Name:       "Common Workflow Language",
		URL:        "https://www.commonwl.org/",
	})

	//TODO SoftwareApplication
	for _, step := range steps {
		graph = append(graph, SoftwareApplication{
			ID:              step.ID,
			Type:            "SoftwareApplication",
			Name:            "TODO",
			Description:     "TODO",
			SoftwareVersion: "TODO",
		})
	}

	//Datasets
	for _, dataset := range datasets {
		graph = append(graph, DatasetDetails{
			ID:            dataset.ID,
			Type:          "Dataset",
			Name:          "TODO",
			Abstract:      "TODO",
			URL:           "TODO",
			Author:        IDRef{"TODO"},
			ExampleOfWork: IDRef{"#" + dataset.ID + "-param"},
		})
	}

	graph = append(graph, Person{
		ID:          "TODO",
		Type:        "Person",
		Name:        "TODO",
		Affiliation: IDRef{"TODO"},
	})

	graph = append(graph, Organization{
		ID:   "TODO",
		Type: "Organization",
		Name: "TODO",
	})

	return RoCrate{
		Context: "https://w3id.org/ro/crate/1.1/context",
		Graph:   graph,
	}, nil
}
