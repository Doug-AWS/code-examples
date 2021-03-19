package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// Config stores the values from config.json
type Config struct {
	RoleName string   `json:"RoleName"`
	TableArn string   `json:"TableArn"`
	Accounts []string `json:"Accounts"`
}

var configFileName = "config.json"

var globalConfig Config

func populateConfiguration(debug bool) error {
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		msg := "Could not read " + configFileName
		return errors.New(msg)
	}

	text := string(content)

	err = json.Unmarshal([]byte(text), &globalConfig)
	if err != nil {
		msg := "Could not unmarshall bytes from " + configFileName
		return errors.New(msg)
	}

	if globalConfig.RoleName == "" || globalConfig.TableArn == "" || len(globalConfig.Accounts) == 0 {
		msg := "You musts supply a value for RoleName, TableArn, and values for Accounts in " + configFileName
		return errors.New(msg)
	}

	debugPrint(debug, "RoleName == "+globalConfig.RoleName)
	debugPrint(debug, "TableArn == "+globalConfig.TableArn)
	debugPrint(debug, "Accounts:")

	for _, a := range globalConfig.Accounts {
		debugPrint(debug, "  "+a)
	}

	return nil
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

func addAccountsToRole(debug bool, roleName string, tableArn string, accountList []string) error {
	dbPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid": "ListAndDescribe",
				"Action": []string{
					"dynamodb:List*",
					"dynamodb:DescribeReservedCapacity*",
					"dynamodb:DescribeLimits",
					"dynamodb:DescribeTimeToLive",
				},
				"Effect":   "Allow",
				"Resource": "*",
				"Principal": map[string]interface{}{
					"AWS": accountList,
				},
			},
			{
				"Sid": "SpecificTable",
				"Action": []string{
					"dynamodb:BatchGet*",
					"dynamodb:DescribeStream",
					"dynamodb:DescribeTable",
					"dynamodb:Get*",
					"dynamodb:Query",
					"dynamodb:Scan",
					"dynamodb:BatchWrite*",
					"dynamodb:CreateTable",
					"dynamodb:Delete*",
					"dynamodb:Update*",
					"dynamodb:PutItem",
				},
				"Effect":   "Allow",
				"Resource": tableArn,
				"Principal": map[string]interface{}{
					"AWS": accountList,
				},
			},
		},
	}

	policy, err := json.Marshal(dbPolicy)
	if err != nil {
		fmt.Println("Got error marshalling policy JSON:")
		fmt.Println(err.Error())
		return err
	}

	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Create a new IAM Service Client
	client := iam.NewFromConfig(cfg)

	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(string(policy)),
		RoleName:                 aws.String(roleName),
	}

	_, err = client.CreateRole(context.TODO(), input)
	if err != nil {
		fmt.Println("Error adding policy to bucket:")
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func main() {
	debug := flag.Bool("d", false, "Whether to print additional info.")
	flag.Parse()

	err := populateConfiguration(*debug)
	if err != nil {
		fmt.Println("Could not get values from " + configFileName)
		return
	}

	debugPrint(*debug, "Debugging enabled")

	err = addAccountsToRole(*debug, globalConfig.RoleName, globalConfig.TableArn, globalConfig.Accounts)
	if err != nil {
		fmt.Println("Got an error calling addAccountsToRole:")
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Added the following accounts to table with ARN " + globalConfig.TableArn)

	for _, a := range globalConfig.Accounts {
		fmt.Println("  " + a)
	}
}
