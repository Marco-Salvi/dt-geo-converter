package rocrate

import (
	"dt-geo-db/cwl"
	"dt-geo-db/logger"
)

// GenerateRoCrate creates a RO-Crate metadata package from the given CWL.
func GenerateRoCrate(wf string, originalCwl cwl.Cwl) (RoCrate, error) {
	var graph []any

	logger.Info("Starting RO-Crate generation for workflow", wf)

	// Add metadata objects.
	graph = append(graph, Metadata{
		ID:         "ro-crate-metadata.json",
		Type:       "CreativeWork",
		ConformsTo: []IDRef{{"https://w3id.org/ro/crate/1.1"}, {"https://w3id.org/workflowhub/workflow-ro-crate/1.0"}},
		About:      IDRef{"./"},
	})
	// graph = append(graph, CreativeWork{
	// 	ID:      "https://w3id.org/ro/wfrun/process/0.4",
	// 	Type:    "CreativeWork",
	// 	Name:    "Process Run Crate",
	// 	Version: "0.4",
	// })
	// graph = append(graph, CreativeWork{
	// 	ID:      "https://w3id.org/ro/wfrun/workflow/0.4",
	// 	Type:    "CreativeWork",
	// 	Name:    "Workflow Run Crate",
	// 	Version: "0.4",
	// })
	graph = append(graph, CreativeWork{
		ID:      "https://w3id.org/workflowhub/workflow-ro-crate/1.0",
		Type:    "CreativeWork",
		Name:    "Workflow RO-Crate",
		Version: "1.0",
	})
	graph = append(graph, ComputerLanguage{
		ID:         "https://w3id.org/workflowhub/workflow-ro-crate#cwl",
		Type:       "ComputerLanguage",
		Identifier: "https://www.commonwl.org/",
		Name:       "Common Workflow Language",
		URL:        "https://www.commonwl.org/",
	})
	// Add organization and person templates.
	graph = append(graph, Organization{
		ID:   "TODO: this is just a template. You can copy-paste this template to have more than one in the RO-Crate",
		Type: "Organization",
		Name: "TODO",
	})
	graph = append(graph, Person{
		ID:          "TODO: this is just a template. You can copy-paste this template to have more than one in the RO-Crate",
		Type:        "Person",
		Name:        "TODO",
		Affiliation: IDRef{"TODO"},
	})

	// Add workflow and related items.
	logger.Debug("Adding workflow details to RO-Crate")
	err := addWorkflowToRoCrate(&graph, wf, originalCwl)
	if err != nil {
		logger.Error("Error adding workflow to RO-Crate:", err)
		return RoCrate{}, err
	}

	logger.Info("RO-Crate generation completed successfully")
	return RoCrate{
		Context: "https://w3id.org/ro/crate/1.1/context",
		Graph:   graph,
	}, nil
}

// addWorkflowToRoCrate adds the workflow and its parts to the RO-Crate graph.
func addWorkflowToRoCrate(rocrate *[]any, wf string, originalCwl cwl.Cwl) error {
	datasets := getAllDTs(originalCwl)

	// Check if the RO-Crate already has a workflow item (with id "./").
	hasWorkflow := false
	for _, item := range *rocrate {
		if _, ok := item.(Workflow); ok {
			hasWorkflow = true
			break
		}
	}
	if !hasWorkflow {
		logger.Debug("Workflow item not found in RO-Crate. Adding workflow item for", wf)
		var workflowHasPart []IDRef
		workflowHasPart = append(workflowHasPart, IDRef{wf})
		for _, dataset := range datasets {
			workflowHasPart = append(workflowHasPart, IDRef{dataset})
		}
		for id, step := range originalCwl.Steps {
			if s, ok := step.Run.(string); ok {
				workflowHasPart = append(workflowHasPart, IDRef{s})
				logger.Debug("Detected sub-workflow step:", s)
			} else {
				workflowHasPart = append(workflowHasPart, IDRef{id})
				logger.Debug("Adding step:", id)
			}
		}
		*rocrate = append(*rocrate, Workflow{
			ID:          "./",
			Type:        "Dataset",
			Name:        "TODO",
			Description: "TODO",
			License:     "TODO",
			Author:      IDRef{"TODO"},
			ConformsTo: []IDRef{
				// {"https://w3id.org/ro/wfrun/process/0.4"},
				// {"https://w3id.org/ro/wfrun/workflow/0.4"},
				{"https://w3id.org/workflowhub/workflow-ro-crate/1.0"},
			},
			HasPart:    workflowHasPart,
			MainEntity: IDRef{wf},
		})
		logger.Info("Workflow item added to RO-Crate")
	} else {
		logger.Debug("Workflow item already exists in RO-Crate. Skipping addition.")
	}

	// Add the computational workflow file item.
	workflowInputsMap := make(map[string]string)
	var workflowInputs []IDRef
	workflowOutputsMap := make(map[string]string)
	var workflowOutputs []IDRef
	for s := range originalCwl.Inputs {
		workflowInputsMap[s] = "#" + s + "->" + wf
		workflowInputs = append(workflowInputs, IDRef{workflowInputsMap[s]})
		logger.Debug("Mapping input dataset", s, "to", workflowInputsMap[s])
	}
	for s := range originalCwl.Outputs {
		workflowOutputsMap[s] = "#" + wf + "->" + s
		workflowOutputs = append(workflowOutputs, IDRef{workflowOutputsMap[s]})
		logger.Debug("Mapping output dataset", s, "to", workflowOutputsMap[s])
	}
	*rocrate = append(*rocrate, ComputationalWorkflowFile{
		ID:                  wf,
		Type:                []string{"File", "SoftwareSourceCode", "ComputationalWorkflow"},
		Name:                "TODO",
		Author:              IDRef{"TODO"},
		Creator:             IDRef{"TODO"},
		ProgrammingLanguage: IDRef{"https://about.workflowhub.eu/Workflow-RO-Crate/#cwl"},
		Input:               workflowInputs,
		Output:              workflowOutputs,
	})
	logger.Info("ComputationalWorkflowFile item added for workflow", wf)

	// Add formal parameters for each input and output.
	for id, param := range workflowInputsMap {
		if parameterExists(*rocrate, param) {
			logger.Debug("Formal parameter", param, "already exists. Skipping.")
			continue
		}
		*rocrate = append(*rocrate, FormalParameter{
			ID:             param,
			Type:           "FormalParameter",
			AdditionalType: "Dataset",
			ConformsTo:     IDRef{"https://bioschemas.org/profiles/FormalParameter/1.0-RELEASE"},
			Description:    "TODO",
			WorkExample:    IDRef{id},
			Name:           id,
			ValueRequired:  true,
		})
		logger.Debug("Added formal parameter for input", id)
	}
	for id, param := range workflowOutputsMap {
		if parameterExists(*rocrate, param) {
			logger.Debug("Formal parameter", param, "already exists. Skipping.")
			continue
		}
		*rocrate = append(*rocrate, FormalParameter{
			ID:             param,
			Type:           "FormalParameter",
			AdditionalType: "Dataset",
			ConformsTo:     IDRef{"https://bioschemas.org/profiles/FormalParameter/1.0-RELEASE"},
			Description:    "TODO",
			WorkExample:    IDRef{id},
			Name:           id,
			ValueRequired:  true,
		})
		logger.Debug("Added formal parameter for output", id)
	}

	// Add a SoftwareSourceCode item for each step.
	for id, step := range originalCwl.Steps {
		if sw, ok := step.Run.(string); ok {
			logger.Info("Processing sub-workflow step", id, "with CWL file", sw)
			subWorkflowCwl, err := cwl.ImportCWL(sw)
			if err != nil {
				logger.Error("Error importing sub-workflow CWL for", sw, ":", err)
				return err
			}
			err = addWorkflowToRoCrate(rocrate, sw, subWorkflowCwl)
			if err != nil {
				logger.Error("Error adding sub-workflow", sw, "to RO-Crate:", err)
				return err
			}
		} else {
			if sscExists(*rocrate, id) {
				logger.Debug("SoftwareSourceCode item for", id, "already exists. Skipping.")
				continue
			}
			*rocrate = append(*rocrate, SoftwareSourceCode{
				ID:                  id,
				Type:                "SoftwareSourceCode",
				Name:                "TODO",
				Description:         "TODO",
				Author:              IDRef{"TODO"},
				License:             "TODO",
				ProgrammingLanguage: "TODO",
				Url:                 "TODO: add the url to the software definition from the EPOS APIs",
			})
			logger.Debug("Added SoftwareSourceCode item for step", id)
		}
	}

	// Add DatasetDetails items.
	for _, dataset := range datasets {
		if datasetExists(*rocrate, dataset) {
			logger.Debug("Dataset", dataset, "already exists. Skipping.")
			continue
		}
		paramPresent := false
		for _, item := range *rocrate {
			if param, ok := item.(FormalParameter); ok && param.ID == "#"+dataset {
				paramPresent = true
				break
			}
		}
		if paramPresent {
			*rocrate = append(*rocrate, DatasetDetails{
				ID:            dataset,
				Type:          "Dataset",
				Name:          "TODO",
				URL:           "TODO",
				ExampleOfWork: &IDRef{"#" + dataset},
			})
			logger.Debug("Added DatasetDetails (with formal parameter) for dataset", dataset)
		} else {
			*rocrate = append(*rocrate, DatasetDetails{
				ID:   dataset,
				Type: "Dataset",
				Name: "TODO",
				URL:  "TODO",
			})
			logger.Debug("Added DatasetDetails for dataset", dataset)
		}
	}

	return nil
}

// getAllDTs returns a list of all dataset IDs found in the CWL.
func getAllDTs(cwlObj cwl.Cwl) []string {
	dts := make(map[string]bool)
	// Global inputs.
	for dt := range cwlObj.Inputs {
		dts[dt] = true
	}
	// Global outputs.
	for dt := range cwlObj.Outputs {
		dts[dt] = true
	}
	datasets := make([]string, 0, len(dts))
	for k := range dts {
		datasets = append(datasets, k)
	}
	logger.Debug("Collected", len(datasets), "datasets from CWL")
	return datasets
}

// getAllSTs returns a list of all step IDs in the CWL.
func getAllSTs(cwlObj cwl.Cwl) []string {
	sts := make([]string, 0, len(cwlObj.Steps))
	for k := range cwlObj.Steps {
		sts = append(sts, k)
	}
	logger.Debug("Collected", len(sts), "steps from CWL")
	return sts
}

func parameterExists(graph []any, param string) bool {
	for _, item := range graph {
		if p, ok := item.(FormalParameter); ok {
			if param == p.ID {
				return true
			}
		}
	}
	return false
}

func datasetExists(graph []any, dataset string) bool {
	for _, item := range graph {
		if p, ok := item.(DatasetDetails); ok {
			if dataset == p.ID {
				return true
			}
		}
	}
	return false
}

func sscExists(graph []any, ssc string) bool {
	for _, item := range graph {
		if p, ok := item.(SoftwareSourceCode); ok {
			if ssc == p.ID {
				return true
			}
		}
	}
	return false
}
