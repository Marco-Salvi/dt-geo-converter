package cwl

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Cwl represents the top-level CWL workflow.
type Cwl struct {
	CWLVersion string            `yaml:"cwlVersion"`
	Class      string            `yaml:"class"`
	Inputs     map[string]IOType `yaml:"inputs"`
	Outputs    map[string]Output `yaml:"outputs"`
	Steps      map[string]Step   `yaml:"steps"`
}

// Output represents an output in the CWL workflow.
type Output struct {
	Type         IOType `yaml:"type"`
	OutputSource string `yaml:"outputSource"`
}

// Step represents a step in the CWL workflow.
type Step struct {
	Run Run               `yaml:"run"`
	In  map[string]string `yaml:"in"`  // Corrected to map[string]string
	Out []string          `yaml:"out"` // Corrected to []string
}

// Run represents the run section of a step.
type Run struct {
	Class   string            `yaml:"class"`
	Inputs  map[string]IOType `yaml:"inputs"`
	Outputs map[string]IOType `yaml:"outputs"`
}

// IOType represents the type of an input/output (e.g., Directory, File).
type IOType string

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
