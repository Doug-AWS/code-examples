package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Config stores the values from config.json
type Config struct {
	Table string `json:"TableName"`
	Key   string `json:"KeyName"`
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

func main() {
	err := populateConfiguration()
	if err != nil {
		fmt.Println("Could not get values from " + configFileName)
		return
	}
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Using the Config value, create the DynamoDB client
	// Create a new DynamoDB Service Client
	client := dynamodb.NewFromConfig(cfg)

	// Create attribute definitions
	var attrs []types.AttributeDefinition

	keyAttr := types.AttributeDefinition{
		AttributeName: &globalConfig.Key,
		AttributeType: types.ScalarAttributeTypeS,
	}

	attrs = append(attrs, keyAttr)

	// Create key schema elements
	var keySchemaElements []types.KeySchemaElement

	keyElement := types.KeySchemaElement{
		AttributeName: &globalConfig.Key,
		KeyType:       "HASH",
	}

	keySchemaElements = append(keySchemaElements, keyElement)

	// Create table
	input := &dynamodb.CreateTableInput{
		TableName:            &globalConfig.Table,
		AttributeDefinitions: attrs,
		KeySchema:            keySchemaElements,
		BillingMode:          types.BillingModePayPerRequest,
	}

	_, err = client.CreateTable(context.TODO(), input)
	if err != nil {
		panic("failed to describe table, " + err.Error())
	}

	fmt.Println("Created " + globalConfig.Table + " with partition key " + globalConfig.Key)
}
