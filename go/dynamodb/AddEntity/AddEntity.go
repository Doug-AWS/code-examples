package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Config stores the values from config.json
type Config struct {
	Table string `json:"TableName"`
	Key   string `json:"KeyName"`
	//Service   string
	Services []string `json:"ServiceNames"` // To validate the -s option
	//Sdk       string
	Sdks []string `json:"SdkNames"` // To validate the -k option
	//Target    string
	Targets []string `json:"TargetNames"` // To validate the -t option
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

	if globalConfig.Table == "" || globalConfig.Key == "" {
		msg := "You musts supply a value for TableName and KeyName in " + configFileName
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

func main() {
	path := flag.String("p", "", "The fully-qualified path to the link target")
	service := flag.String("s", "", "The service for which the entity is created")
	sdk := flag.String("k", "", "The SDK's programming language extension (cpp, py, etc.)")
	target := flag.String("t", "", "The target of the entity: a service (guide) or code (catalog)")
	action := flag.String("a", "", "Whether the link is to a (section) such as 'sns', or Action, such as 'CreateTopic'")
	debug := flag.Bool("d", false, "Whether to barf out more info")

	flag.Parse()

	debugPrint(*debug, "Debugging enabled")

	if *path == "" || *service == "" || *sdk == "" || *target == "" || *action == "" {
		fmt.Println("You must supply a path, service, sdk, target, and action value")
		fmt.Println("-p PATH -s SERVICE -k SDK -t TARGET -a ACTION")
		fmt.Println("See config.json for valid values for service, sdk, target, and action")
		return
	}

	debugPrint(*debug, "")
	debugPrint(*debug, "Path:    "+*path)
	debugPrint(*debug, "Service: "+*service)
	debugPrint(*debug, "SDK:     "+*sdk)
	debugPrint(*debug, "Target:  "+*target)
	debugPrint(*debug, "Action:  "+*action)
	debugPrint(*debug, "")

	err := populateConfiguration()
	if err != nil {
		fmt.Println("Could not parse " + configFileName)
		return
	}

	isValid := isValidItem(*service, globalConfig.Services)
	if !isValid {
		fmt.Println(*service + " is not in the list of services:")
		fmt.Println(globalConfig.Services)
		return
	}

	isValid = isValidItem(*sdk, globalConfig.Sdks)
	if !isValid {
		fmt.Println(*sdk + " is not in the list of SDKs:")
		fmt.Println(globalConfig.Sdks)
		return
	}

	isValid = isValidItem(*target, globalConfig.Targets)
	if !isValid {
		fmt.Println(*target + " is not in the list of targets:")
		fmt.Println(globalConfig.Targets)
		return
	}

	/*
		We overload action.
		If it's "section", we have a link to
		something like the SNS code examples section in the Java Dev guide.
		Otherwise, it's an individual topic for the *action operation.
	*/

	err = isPathValid(*path, *sdk, *target)
	if err != nil {
		fmt.Println(*path + " is not a valid path for target " + *target)
		return
	}

	attrs := make(map[string]types.AttributeValue, 5)

	attrs["path"] = &types.AttributeValueMemberS{
		Value: *path,
	}

	attrs["service"] = &types.AttributeValueMemberS{
		Value: *service,
	}

	attrs["sdk"] = &types.AttributeValueMemberS{
		Value: *sdk,
	}

	attrs["target"] = &types.AttributeValueMemberS{
		Value: *target,
	}

	attrs["action"] = &types.AttributeValueMemberS{
		Value: *action,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)

	dynamodbInput := &dynamodb.PutItemInput{
		TableName: &globalConfig.Table,
		Item:      attrs,
	}

	_, err = dynamodbClient.PutItem(context.TODO(), dynamodbInput)
	if err != nil {
		fmt.Println("Got error calling PutItem: ")
		fmt.Println(err)
		return
	}
}
