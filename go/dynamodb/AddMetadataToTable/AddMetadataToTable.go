package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"gopkg.in/yaml.v2"
)

// Config holds the info in config.json
type Config struct {
	Table string `json:"TableName"`
}

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

var configFileName = "config.json"

var globalConfig Config

func populateConfiguration() error {
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	text := string(content)

	err = json.Unmarshal([]byte(text), &globalConfig)
	if err != nil {
		return err
	}

	if globalConfig.Table == "" {
		msg := "You musts supply a value for TableName " + configFileName
		return errors.New(msg)
	}

	return nil
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

func transmogrifyPath(debug bool, service, path string) string {
	// Change path like:
	//   CopyObject/CopyObjectv2.go
	// To:
	//   https://aws.github.io/aws-sdk-go-v2/docs/code-examples/s3/copyobject/

	// First split the path by '/'
	parts := strings.Split(path, "/")
	dir := strings.ToLower(parts[0])

	return "https://aws.github.io/aws-sdk-go-v2/docs/code-examples/" + service + "/" + dir
}

func addMetadataToTable(debug bool, filename string, tablename string, ext string, target string) (int, error) {
	debugPrint(debug, "Parsing "+filename)

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		msg := "Got error reading YAML file: " + filename + "\n" + err.Error()
		return 0, errors.New(msg)
	}

	var metadata Metadata
	err = yaml.Unmarshal(yamlFile, &metadata)
	if err != nil {
		msg := "Got error unmarshalling YAML: for " + filename + "\n" + err.Error()
		return 0, errors.New(msg)
	}

	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := dynamodb.NewFromConfig(cfg)

	itemsAdded := 0

	for _, f := range metadata.Files {
		for _, s := range f.Services {
			path := transmogrifyPath(debug, s.Service, f.Path)
			for _, a := range s.Actions {
				// Don't add actions that are "test"
				if a != "test" {
					debugPrint(debug, "Path:        "+path)
					debugPrint(debug, "Action:      "+a)
					debugPrint(debug, "SDK:         "+ext)
					debugPrint(debug, "Service:     "+s.Service)
					debugPrint(debug, "Target:      "+target)
					debugPrint(debug, "Description: "+f.Description)
					debugPrint(debug, "")

					// Create attributes for new table item
					attrs := make(map[string]types.AttributeValue, 6)

					attrs["path"] = &types.AttributeValueMemberS{
						Value: path,
					}

					attrs["action"] = &types.AttributeValueMemberS{
						Value: a,
					}

					attrs["sdk"] = &types.AttributeValueMemberS{
						Value: ext,
					}

					attrs["service"] = &types.AttributeValueMemberS{
						Value: s.Service,
					}

					attrs["target"] = &types.AttributeValueMemberS{
						Value: target,
					}

					attrs["description"] = &types.AttributeValueMemberS{
						Value: f.Description,
					}

					dynamodbInput := &dynamodb.PutItemInput{
						TableName: &globalConfig.Table,
						Item:      attrs,
					}

					_, err = client.PutItem(context.TODO(), dynamodbInput)
					if err != nil {
						fmt.Println("Got error calling PutItem: ")
						fmt.Println(err)
						return 0, err
					}

					itemsAdded++
				}
			}
		}

		debugPrint(debug, "")
	}

	return itemsAdded, nil
}

func main() {
	root := flag.String("r", "", "The root of the Go v2 directory on this computer")
	ext := flag.String("e", "", "The file extension of the code examples")
	target := flag.String("t", "guide", "guide or catalog")
	debug := flag.Bool("d", false, "Whether to barf out more info")

	flag.Parse()

	debugPrint(*debug, "Debugging enabled")

	if *root == "" || *ext == "" || *target == "" {
		fmt.Println("You must supply the path to the SDK files (-r ROOT),")
		fmt.Println("extension of the files in that directory (-e EXTENSION, such as 'go'),")
		fmt.Println("and target (-t guide | -t catalog)")
		return
	}

	dir := *root

	// If root ends with a '/', remove it as we always add it later
	/*
			   firstN := s[0:N]
		       lastN  := s[len(s)-N:]
	*/
	if dir[len(dir)-1:] == "/" {
		debugPrint(*debug, "Root before: "+*root)
		*root = dir[:len(dir)-1]
		debugPrint(*debug, "Root after: "+*root)
	}

	err := populateConfiguration()
	if err != nil {
		fmt.Println("Could not parse " + configFileName)
		return
	}

	// Navigate into sub-directories,
	// if metatdata.yaml found,
	// call addMetadataToTable with full path.
	files, err := ioutil.ReadDir(*root)
	if err != nil {
		fmt.Println("Could not read contents of directory " + *root)
		return
	}

	numEntries := 0

	for _, f := range files {
		if f.IsDir() {
			debugPrint(*debug, " Navigating into "+f.Name())

			dFiles, err := ioutil.ReadDir(*root + "/" + f.Name())
			if err != nil {
				fmt.Println("Could not read contents of directory " + *root + "/" + f.Name())
				return
			}

			debugPrint(*debug, "Read contents of "+*root+"/"+f.Name())

			for _, m := range dFiles {
				if m.Name() == "metadata.yaml" {
					num, err := addMetadataToTable(*debug, *root+"/"+f.Name()+"/"+m.Name(), globalConfig.Table, *ext, *target)
					if err != nil {
						fmt.Println("Got an error adding items to table:")
						fmt.Println(err.Error())
						return
					}

					debugPrint(*debug, "Read contents of "+*root+"/"+f.Name()+"/"+m.Name())
					numEntries += num

				}
			}
		}
	}

	fmt.Println("Added " + strconv.Itoa(numEntries) + " items to the table")
}
