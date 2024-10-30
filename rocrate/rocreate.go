package rocrate

import (
	"encoding/json"
	"os"
)

type RoCrate struct {
	Context string `json:"@context"`
	Graph   []any  `json:"@graph"`
}

type Metadata struct {
	ID         string  `json:"@id"`
	Type       string  `json:"@type"`
	ConformsTo []IDRef `json:"conformsTo"`
	About      IDRef   `json:"about"`
}

type Workflow struct {
	ID          string  `json:"@id"`
	Type        string  `json:"@type"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	License     string  `json:"license"`
	Author      IDRef   `json:"author"`
	ConformsTo  []IDRef `json:"conformsTo"`
	HasPart     []IDRef `json:"hasPart"`
	MainEntity  IDRef   `json:"mainEntity"`
}

type CreativeWork struct {
	ID      string `json:"@id"`
	Type    string `json:"@type"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ComputationalWorkflowFile struct {
	ID                  string   `json:"@id"`
	Type                []string `json:"@type"`
	Name                string   `json:"name"`
	Author              IDRef    `json:"author"`
	Creator             IDRef    `json:"creator"`
	ProgrammingLanguage IDRef    `json:"programmingLanguage"`
	Input               []IDRef  `json:"input"`
	Output              []IDRef  `json:"output"`
}

type FormalParameter struct {
	ID             string `json:"@id"`
	Type           string `json:"@type"`
	AdditionalType string `json:"additionalType"`
	ConformsTo     IDRef  `json:"conformsTo"`
	Description    string `json:"description"`
	WorkExample    IDRef  `json:"workExample"`
	Name           string `json:"name"`
	ValueRequired  bool   `json:"valueRequired"`
}

type ComputerLanguage struct {
	ID         string `json:"@id"`
	Type       string `json:"@type"`
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	URL        string `json:"url"`
}

type SoftwareSourceCode struct {
	ID                  string `json:"@id"`
	Type                string `json:"@type"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Author              IDRef  `json:"author"`
	License             string `json:"license"`
	ProgrammingLanguage string `json:"programmingLanguage"`
	Url                 string `json:"url,omitempty"`
	// SoftwareVersion string `json:"softwareVersion"`
}

type DatasetDetails struct {
	ID       string `json:"@id"`
	Type     string `json:"@type"`
	Name     string `json:"name"`
	Abstract string `json:"abstract,omitempty"`
	URL      string `json:"url"`
	// need to be a pointer to make omitempty work correctly
	Author        *IDRef `json:"author,omitempty"`
	ExampleOfWork *IDRef `json:"exampleOfWork,omitempty"`
}

type Person struct {
	ID          string `json:"@id"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	Affiliation IDRef  `json:"affiliation"`
}

type Organization struct {
	ID   string `json:"@id"`
	Type string `json:"@type"`
	Name string `json:"name"`
}

type IDRef struct {
	ID string `json:"@id,omitempty"`
}

func (r RoCrate) SaveToFile(name string) error {
	v, err := json.Marshal(r)
	if err != nil {
		return err
	}

	err = os.WriteFile(name, v, 0622)
	if err != nil {
		return err
	}

	return nil
}
