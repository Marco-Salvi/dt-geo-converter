package cwl

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Cwl represents the top-level CWL workflow.
type Cwl struct {
	CWLVersion   string                       `yaml:"cwlVersion"`
	Class        string                       `yaml:"class"`
	Inputs       map[string]any               `yaml:"inputs"` // TODO find a better way to represent both IOType and Dataset instead of any
	Outputs      map[string]Output            `yaml:"outputs"`
	Requirements map[string]map[string]string `yaml:"requirements,omitempty"`
	Steps        map[string]Step              `yaml:"steps"`
}

// Output represents an output in the CWL workflow.
type Output struct {
	Type         IOType `yaml:"type"`
	OutputSource string `yaml:"outputSource"`
}

// Step represents a step in the CWL workflow.
type Step struct {
	Run any               `yaml:"run,omitempty"` // TODO find a better way to represent both Run and string objects instead of any
	In  map[string]string `yaml:"in"`
	Out []string          `yaml:"out"`
}

// Run represents the run section of a step.
type Run struct {
	Class   string            `yaml:"class"`
	Inputs  map[string]IOType `yaml:"inputs"`
	Outputs map[string]IOType `yaml:"outputs"`
}

// IOType represents the type of an input/output (e.g., Directory, File).
type IOType string

type Input struct {
	Type IOType `yaml:"type"`
	Docs string `yaml:"docs,omitempty"`
}

const (
	Directory IOType = "Directory"
	File      IOType = "File"
)

func (c Cwl) SaveToFile(name string) error {
	v, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(name, v, 0622)
	if err != nil {
		return err
	}

	return nil
}

func ImportCWL(filePath string) (Cwl, error) {
	f, err := os.ReadFile(filePath)
	if err != nil {
		return Cwl{}, err
	}

	var cwl Cwl
	err = yaml.Unmarshal(f, &cwl)
	if err != nil {
		return Cwl{}, fmt.Errorf("cwl to unmarshal: %s. err: %v", filePath, err)
	}
	return cwl, nil
}
