package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type Service struct {
	Service string   `yaml:"service"`
	Actions []string `yaml:"actions"`
}

type File struct {
	Description string    `yaml:"description"`
	Path        string    `yaml:"path"`
	Services    []Service `yaml:"services"`
}

// Metadata caches the info in metadata.yaml
type Metadata struct {
	Files []File `yaml:"files"`
}

func main() {
	// Create some metadata
	svc := Service{
		Service: "service1",
		Actions: []string{
			"action1",
			"action2",
		},
	}

	file := File{
		Description: "Blah, blah, blah",
		Path:        "c:/",
		Services: []Service{
			svc,
		},
	}

	// Now marshall it into YAML
	metadata := Metadata{
		Files: []File{
			file,
		},
	}

	result, err := yaml.Marshal(metadata)
	if err != nil {
		fmt.Println("Got error unmarshalling YAML: " + err.Error())
	} else {
		fmt.Println(string(result))
	}
}
