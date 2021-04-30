package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"

	"gopkg.in/yaml.v2"
)

// Metadata defines the elements in metatdata.yaml
type Metadata struct {
	Files []struct {
		Path        string `yaml:"path"`
		Description string `yaml:"description"`
		Services    []struct {
			Service string   `yaml:"service"`
			Actions []string `yaml:"actions"`
		} `yaml:"services"`
	} `yaml:"files"`
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

func main() {
	debugPtr := flag.Bool("d", false, "Show extra info")
	flag.Parse()
	debug := *debugPtr

	path := "https://raw.githubusercontent.com/awsdocs/aws-doc-sdk-examples/doug-test-go-metadata/dotnetv3/ACM/metadata.yaml"

	fmt.Println("Parsing " + path)

	var metadata Metadata

	results, err := http.Get(path)
	if err != nil {
		fmt.Println("Got an error getting the file:")
		fmt.Println(err)
		return
	}
	defer results.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(results.Body)
	bytes := buf.Bytes()

	err = yaml.Unmarshal(bytes, &metadata)
	if err != nil {
		fmt.Println("Got an error unmarshalling:")
		fmt.Println(err)
		return
	}

	debugPrint(debug, "Unmarshalled data for "+path)
	if debug {
		fmt.Println("Data:")
		fmt.Println(metadata)
	}

	// Iterate through files.
	// Create a link and tab for services/operations
	for _, data := range metadata.Files {
		debugPrint(debug, "")
		debugPrint(debug, "Path:        "+data.Path)
		debugPrint(debug, "Description: "+data.Description)
		debugPrint(debug, "Services")

		for _, s := range data.Services {
			debugPrint(debug, "  "+s.Service)
			debugPrint(debug, "    Actions:")

			for _, a := range s.Actions {
				debugPrint(debug, "      "+a)

			}
		}
	}

}
