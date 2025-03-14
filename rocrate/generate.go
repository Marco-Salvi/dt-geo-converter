package rocrate

import (
	"dt-geo-db/cwl"
	"log"
)

func GenerateRoCrate(wf string, originalCwl cwl.Cwl) (RoCrate, error) {
	var graph []any

	// metadata object
	graph = append(graph, Metadata{
		ID:         "ro-crate-metadata.json",
		Type:       "CreativeWork",
		ConformsTo: []IDRef{{"https://w3id.org/ro/crate/1.1"}, {"https://w3id.org/workflowhub/workflow-ro-crate/1.0"}},
		About:      IDRef{"./"},
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

	graph = append(graph, ComputerLanguage{
		ID:         "https://w3id.org/workflowhub/workflow-ro-crate#cwl",
		Type:       "ComputerLanguage",
		Identifier: "https://www.commonwl.org/",
		Name:       "Common Workflow Language",
		URL:        "https://www.commonwl.org/",
	})

	// add an organization template
	graph = append(graph, Organization{
		ID:   "TODO: this is just a template. You can copy-paste this template to have more than one in the RO-Crate",
		Type: "Organization",
		Name: "TODO",
	})

	// add a person template
	graph = append(graph, Person{
		ID:          "TODO: this is just a template. You can copy-paste this template to have more than one in the RO-Crate",
		Type:        "Person",
		Name:        "TODO",
		Affiliation: IDRef{"TODO"},
	})

	err := addWorkflowToRoCrate(&graph, wf, originalCwl)
	if err != nil {
		return RoCrate{}, err
	}

	return RoCrate{
		Context: "https://w3id.org/ro/crate/1.1/context",
		Graph:   graph,
	}, nil
}

func addWorkflowToRoCrate(rocrate *[]any, wf string, originalCwl cwl.Cwl) error {
	datasets := getAllDTs(originalCwl)

	// check if the rocrate already has an item with the ./ id
	hasWorkflow := false
	for _, item := range *rocrate {
		if _, ok := item.(Workflow); ok {
			hasWorkflow = true
			break
		}
	}
	// if there is no item with the ./ id then add it, else skip it
	if !hasWorkflow {
		var workflowHasPart []IDRef
		workflowHasPart = append(workflowHasPart, IDRef{wf})

		for _, dataset := range datasets {
			workflowHasPart = append(workflowHasPart, IDRef{dataset})
		}
		for id, step := range originalCwl.Steps {
			if s, ok := step.Run.(string); ok {
				// the step is a sub-workflow
				workflowHasPart = append(workflowHasPart, IDRef{s})
			} else {
				workflowHasPart = append(workflowHasPart, IDRef{id})
			}
		}

		*rocrate = append(*rocrate, Workflow{
			ID:          "./",
			Type:        "Dataset",
			Name:        "TODO",
			Description: "TODO",
			License:     "TODO",
			Author:      IDRef{"TODO"},
			ConformsTo:  []IDRef{{"https://w3id.org/ro/wfrun/process/0.4"}, {"https://w3id.org/ro/wfrun/workflow/0.4"}, {"https://w3id.org/workflowhub/workflow-ro-crate/1.0"}},
			HasPart:     workflowHasPart,
			MainEntity:  IDRef{wf},
		})
	}

	// add the computationalworkflowfile item
	workflowInputsMap := make(map[string]string)
	var workflowInputs []IDRef
	workflowOutputsMap := make(map[string]string)
	var workflowOutputs []IDRef
	for s := range originalCwl.Inputs {
		workflowInputsMap[s] = "#" + s + "->" + wf
		workflowInputs = append(workflowInputs, IDRef{workflowInputsMap[s]})
	}
	for s := range originalCwl.Outputs {
		workflowOutputsMap[s] = "#" + wf + "->" + s
		workflowOutputs = append(workflowOutputs, IDRef{workflowOutputsMap[s]})
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

	// add formal parameters for each input and output
	for id, param := range workflowInputsMap {
		if parameterExists(*rocrate, param) {
			log.Printf("Parameter with id: %s, already exists", param)
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
	}
	for id, param := range workflowOutputsMap {
		if parameterExists(*rocrate, param) {
			log.Printf("Parameter with id: %s, already exists", param)
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
	}

	// add a softwaresourcecode item for each step of the workflow, if the step is a sub-workflow, call this function recursively
	for id, step := range originalCwl.Steps {
		if sw, ok := step.Run.(string); ok {
			subWorkflowCwl, err := cwl.ImportCWL(sw)
			if err != nil {
				return err
			}
			err = addWorkflowToRoCrate(rocrate, sw, subWorkflowCwl)
			if err != nil {
				return err
			}
		} else {
			if sscExists(*rocrate, id) {
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
		}
	}

	//Datasets
	for _, dataset := range datasets {
		// if the dataset already exist don't add it again
		if datasetExists(*rocrate, dataset) {
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
		} else {
			*rocrate = append(*rocrate, DatasetDetails{
				ID:   dataset,
				Type: "Dataset",
				Name: "TODO",
				URL:  "TODO",
			})
		}
	}

	return nil
}

// get all the dts for this workflow
func getAllDTs(cwl cwl.Cwl) []string {
	dts := make(map[string]bool)

	// add the global inputs
	for dt := range cwl.Inputs {
		dts[dt] = true
	}
	// add the global outputs
	for dt := range cwl.Outputs {
		dts[dt] = true
	}

	// for each step add all the input and output datasets used in the step
	// for _, st := range cwl.Steps {
	// 	for dt := range st.In {
	// 		dts[dt] = true
	// 	}
	// 	for _, dt := range st.Out {
	// 		dts[dt] = true
	// 	}
	// }

	datasets := make([]string, 0, len(dts))
	for k := range dts {
		datasets = append(datasets, k)
	}
	return datasets
}

// get all the steps for this workflow
func getAllSTs(cwl cwl.Cwl) []string {
	sts := make([]string, 0, len(cwl.Steps))
	for k := range cwl.Steps {
		sts = append(sts, k)
	}
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
