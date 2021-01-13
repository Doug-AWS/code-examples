package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Metadata defines the elements in metatdata.yaml
type Metadata struct {
	Files []struct {
		Path        string `yaml:"path"`
		Description string `yaml:"description"`
		Services    []struct {
			Service    string   `yaml:"service"`
			Operations []string `yaml:"operations"`
		} `yaml:"services"`
	} `yaml:"files"`
}

func main() {
	filename := "metadata.yaml"

	fmt.Println("Parsing " + filename)

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return
	}

	var metadata Metadata
	err = yaml.Unmarshal(yamlFile, &metadata)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
	}

	fmt.Println("Found", len(metadata.Files), "files:")

	for _, data := range metadata.Files {
		fmt.Println("Path:        " + data.Path)
		fmt.Println("Description: " + data.Description)
		fmt.Println("Services")

		for _, s := range data.Services {
			fmt.Println("  " + s.Service)
			fmt.Println("    Operations:")

			for _, o := range s.Operations {
				fmt.Println("      " + o)
			}
		}
	}
}
