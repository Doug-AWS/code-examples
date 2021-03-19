package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Config stores the values from config.json
type Config struct {
	Table        string   `json:"TableName"`
	PartitionKey string   `json:"PartitionKeyName"`
	SortKey      string   `json:"SortKeyName"`
	Services     []string `json:"ServiceNames"` // To validate the -s option
	Sdks         []string `json:"SdkNames"`     // To validate the -k option
	Targets      []string `json:"TargetNames"`  // To validate the -t option
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

// ValidAwsGuidePrefix defines the valid prefix for path when target is "guide" FOR MOST GUIDES
const ValidAwsGuidePrefix = "https://docs.aws.amazon.com"

// ValidGoGuidePrefix defines the valid prefix for the go dev guide
const ValidGoGuidePrefix = "https://aws.github.io/aws-sdk-go-v2/docs/code-examples/"

func isPathValid(path string, sdk string, target string) error {
	// If target is "guide"
	// path must start with https://docs.aws.amazon.com
	msg := ""

	switch sdk {
	case "go":
		switch target {
		case "guide":
			valid := strings.HasPrefix(path, ValidGoGuidePrefix)
			if valid {
				return nil
			}
		}
	default:
		switch target {
		case "guide":
			valid := strings.HasPrefix(path, ValidAwsGuidePrefix)
			if valid {
				return nil
			}

			msg = "Path does not start with " + ValidAwsGuidePrefix
			return errors.New(msg)

		default:
			msg = "Valid path prefix is not set for " + target + " target"
		}
	}

	return errors.New(msg)
}

func isValidItem(item string, list []string) bool {
	// Is service in list of services?
	for _, l := range list {
		if item == l {
			return true
		}
	}

	return false
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

func addItemsToTable(debug bool, filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		msg := "Got an error opening " + filename
		return 0, errors.New(msg)
	}

	// Create DynamoDB client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		msg := "Got the configuration error: " + err.Error()
		return 0, errors.New(msg)
	}

	client := dynamodb.NewFromConfig(cfg)

	scanner := bufio.NewScanner(file)

	foundFirstLine := false

	i := 1
	numItems := 0

	for scanner.Scan() {
		text := scanner.Text()
		debugPrint(debug, "Line "+strconv.Itoa(i)+": "+text)
		i++

		// text should contain CSV strings like:
		//   "https://aws.github.io/aws-sdk-go-v2/docs/code-examples/sns/","section","go","sns","guide"
		// so split it by commas, make sure there are 5 parts.
		parts := strings.Split(text, ",")

		if len(parts) != 5 {
			fmt.Println(text)
			fmt.Println("Did not have 5 components")
			return 0, errors.New("Invalid entry")
		}

		// Each part is enclosed by double-quote (") marks.

		path := parts[0][1 : len(parts[0])-1]
		action := parts[1][1 : len(parts[1])-1]
		sdk := parts[2][1 : len(parts[2])-1]
		service := parts[3][1 : len(parts[3])-1]
		target := parts[4][1 : len(parts[4])-1]

		if !foundFirstLine {
			if path == "path" {
				// We have the first item
				// so make sure the schema of the CSV file is correct
				if action != "action" || sdk != "sdk" || service != "service" || target != "target" {
					msg := "The CSV file does not have the correct schema: path, action, sdk, service, target"
					return 0, errors.New(msg)
				}

				debugPrint(debug, "CSV file has the correct schema")

				// Skip first line
				foundFirstLine = true
				continue
			}
		} else {
			debugPrint(debug, "")
			debugPrint(debug, "Path:    "+path)
			debugPrint(debug, "Service: "+service)
			debugPrint(debug, "SDK:     "+sdk)
			debugPrint(debug, "Target:  "+target)
			debugPrint(debug, "Action:  "+action)
			debugPrint(debug, "")

			isValid := isValidItem(service, globalConfig.Services)
			if !isValid {
				msg := service + " is not in the list of services:"
				return 0, errors.New(msg)
			}

			isValid = isValidItem(sdk, globalConfig.Sdks)
			if !isValid {
				msg := sdk + " is not in the list of SDKs:"
				return 0, errors.New(msg)
			}

			isValid = isValidItem(target, globalConfig.Targets)
			if !isValid {
				msg := target + " is not in the list of targets"
				return 0, errors.New(msg)
			}

			/*
				We overload action.
				If it's "section", we have a link to
				something like the SNS code examples section in the Java Dev guide.
				Otherwise, it's an individual topic for the *action operation.
			*/

			err = isPathValid(path, sdk, target)
			if err != nil {
				msg := path + " is not a valid path for target " + target
				return 0, errors.New(msg)
			}

			attrs := make(map[string]types.AttributeValue, 5)

			attrs["path"] = &types.AttributeValueMemberS{
				Value: path,
			}

			attrs["service"] = &types.AttributeValueMemberS{
				Value: service,
			}

			attrs["sdk"] = &types.AttributeValueMemberS{
				Value: sdk,
			}

			attrs["target"] = &types.AttributeValueMemberS{
				Value: target,
			}

			attrs["action"] = &types.AttributeValueMemberS{
				Value: action,
			}

			dynamodbInput := &dynamodb.PutItemInput{
				TableName: &globalConfig.Table,
				Item:      attrs,
			}

			_, err = client.PutItem(context.TODO(), dynamodbInput)
			if err != nil {
				msg := "Got error calling PutItem: " + err.Error()
				return 0, errors.New(msg)
			}

			debugPrint(debug, "Added item to table")

			numItems++
		}
	}

	file.Close()

	return numItems, nil
}

func main() {
	csvFile := flag.String("f", "", "The CSV file to get entries from")
	debug := flag.Bool("d", false, "Whether to barf out more info")

	flag.Parse()

	debugPrint(*debug, "Debugging enabled")

	if *csvFile == "" {
		fmt.Println("You must supply the name of a CSV file (-f FILENAME)")
		return
	}

	err := populateConfiguration()
	if err != nil {
		fmt.Println("Could not parse " + configFileName)
		return
	}

	numItems, err := addItemsToTable(*debug, *csvFile)
	if err != nil {
		fmt.Println("Got an error adding items to table:")
		fmt.Println(err.Error())
	}

	fmt.Println("Added " + strconv.Itoa(numItems) + " items to the table")
}
