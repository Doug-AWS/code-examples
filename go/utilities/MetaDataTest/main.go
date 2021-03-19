package main

import (
	"flag"
	"fmt"

	"gopkg.in/yaml.v3"
)

/*
   The existing structure of metaday.yaml:

   files:
     - path: CopyObject/CopyObjectv2.go
       services:
         - s3
     - path: CopyObject/CopyObjectv2_test.go
       services:
         - s3

	Which gives us a struct like:

	type Metadata struct {
       Files []struct {
          Path     string   `yaml:"path"`
	      Services []string `yaml:"services"`
	   } `yaml:"files"`
    }

	But if we have a code example with multiple services,
	how would that look?

	Here's the struct that handles the yaml this puts out

type Metadata struct {
	Files []struct {
		Path     string `yaml:"path"`
		Services []struct {
			Service string   `yaml:"service"`
			Actions []string `yaml:"actions"`
		} `yaml:"services"`
	} `yaml:"files"`
}
*/

type Service struct {
	Service string
	Actions []string
}

type File struct {
	Path     string
	Services []Service
}

type Metadata struct {
	Files []File
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

func main() {
	debug := flag.Bool("d", false, "Whether to print additional info")
	flag.Parse()

	debugPrint(*debug, "Debugging enabled")

	s1 := Service{
		Service: "s3",
		Actions: []string{"DoThis", "DoThat"},
	}

	s2 := Service{
		Service: "sns",
		Actions: []string{"DidIt"},
	}

	file1 := File{
		Path:     "one/two",
		Services: []Service{s1, s2},
	}

	s3 := Service{
		Service: "sqs",
		Actions: []string{"DoSqsThis", "DoSqsThat"},
	}

	s4 := Service{
		Service: "sns",
		Actions: []string{"SnsDidIt"},
	}

	file2 := File{
		Path:     "one/two",
		Services: []Service{s3, s4},
	}

	myData := Metadata{
		Files: []File{file1, file2},
	}

	// Display yaml
	output, err := yaml.Marshal(myData)
	if err != nil {
		fmt.Println("Got an error marshalling struct:")
		fmt.Println(err.Error())
		return
	}

	fmt.Println(string(output))
}
