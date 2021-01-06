// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

func init() {
	_, _ = callLambda()
}

func callLambda() (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	client := lambda.NewFromConfig(cfg)

	input := &lambda.GetAccountSettingsInput{}
	resp, err := client.GetAccountSettings(context.Background(), input)
	if err != nil {
		return "", err
	}

	output, _ := json.Marshal(resp.AccountUsage)

	return string(output), err
}

func handleRequest(ctx context.Context, event events.SQSEvent) (string, error) {
	// event
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	log.Printf("EVENT: %s", eventJSON)

	// environment variables
	log.Printf("REGION: %s", os.Getenv("AWS_REGION"))
	log.Println("ALL ENV VARS:")

	for _, element := range os.Environ() {
		log.Println(element)
	}

	// request context
	lc, _ := lambdacontext.FromContext(ctx)
	log.Printf("REQUEST ID: %s", lc.AwsRequestID)

	// global variable
	log.Printf("FUNCTION NAME: %s", lambdacontext.FunctionName)

	// context method
	deadline, _ := ctx.Deadline()
	log.Printf("DEADLINE: %s", deadline)

	// AWS SDK call
	usage, err := callLambda()
	if err != nil {
		return "ERROR", err
	}

	return usage, nil
}

func main() {
	runtime.Start(handleRequest)
}
