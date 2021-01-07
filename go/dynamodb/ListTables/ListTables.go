package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	// Load the SDK's default configuration and credentials values from the environment variables,
	// shared credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig()
	// To specify a region:
	//          config.LoadDefaultConfig(config.WithRegion("us-west-2"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	// Create a new DynamoDB Service Client
	client := dynamodb.NewFromConfig(cfg)

	resp, err := client.ListTables(context.Background(), &dynamodb.ListTablesInput{})
	if err != nil {
		panic("failed to list tables, " + err.Error())
	}

	fmt.Println("Tables in " + cfg.Region + " region:")

	for _, n := range resp.TableNames {
		fmt.Println(*n)
	}
}
