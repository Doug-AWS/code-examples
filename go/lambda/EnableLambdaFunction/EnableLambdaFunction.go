// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

func main() {
	// Args for AddPermissionInput
	service := flag.String("s", "", "The service that sends notifications to Lambda")
	function := flag.String("f", "", "The name of the Lambda function that's called")
	resource := flag.String("r", "", "The name of the resource that sends a notification to Lambda")
	// -s service, where s3 -> Principal: "s3.amazonaws.com"
	// -f function, -> FunctionName: "function"
	// -r resource, which for "-s s3 -r mybucket" -> SourceArn: "arn:aws:s3:::mybucket" AND

	flag.Parse()

	if *service == "" || *function == "" || *resource == "" {
		fmt.Println("You must supply the name of the service (-s SERVICE), function (-f FUNCTION), and resource (-r RESOURCE-NAME)")
		return
	}

	// Create Lambda service client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return
	}

	client := lambda.NewFromConfig(cfg)

	permArgs := &lambda.AddPermissionInput{}

	/* For the service, this app only supports:

	   DynamoDB (table)
	   S3 (bucket)
	   SNS (topic)
	   SQS (queue)
	*/

	// Get default region and account ID (to build ARNs):
	region := ""
	accountID := ""

	switch *service {
	case "dynamodb":
		// A DynamoDB table ARN looks like:
		//     arn:aws:dynamodb:REGION:ACCOUNT-ID:table/TABLE-NAME
		tableARN := "arn:aws:dynamodb:" + region + ":" + accountID + ":table/" + *resource

		permArgs = &lambda.AddPermissionInput{
			Action:       aws.String("lambda:InvokeFunction"),
			FunctionName: function,
			Principal:    aws.String(*service + ".amazonaws.com"),
			SourceArn:    aws.String(tableARN),
			StatementId:  aws.String("lambda_dynamodb_notification"),
		}

		break
	case "s3":
		// A bucket ARN looks like:
		//     arn:aws:s3:::BUCKET-NAME
		bucketARN := "arn:aws:s3:::" + *resource

		permArgs = &lambda.AddPermissionInput{
			Action:       aws.String("lambda:InvokeFunction"),
			FunctionName: function,
			Principal:    aws.String(*service + ".amazonaws.com"),
			SourceArn:    aws.String(bucketARN),
			StatementId:  aws.String("lambda_s3_notification"),
		}

		break
	case "sns":
		// An SNS topic ARN looks like:
		//     arn:aws:sns:REGION:ACCOUNT-ID:TOPIC-NAME
		topicARN := "arn:aws:sns:" + region + ":" + accountID + ":" + *resource

		permArgs = &lambda.AddPermissionInput{
			Action:       aws.String("lambda:InvokeFunction"),
			FunctionName: function,
			Principal:    aws.String(*service + ".amazonaws.com"),
			SourceArn:    aws.String(topicARN),
			StatementId:  aws.String("lambda_sns_notification"),
		}

		break

	case "sqs":
		// An SQS queue ARN looks like:
		//    arn:aws:sqs:REGION:ACCOUNT-ID:QUEUE-NAME
		queueARN := "arn:aws:sqs:" + region + ":" + accountID + ":" + *resource

		permArgs = &lambda.AddPermissionInput{
			Action:       aws.String("lambda:InvokeFunction"),
			FunctionName: function,
			Principal:    aws.String(*service + ".amazonaws.com"),
			SourceArn:    aws.String(queueARN),
			StatementId:  aws.String("lambda_sqs_notification"),
		}

		break
	default:
		fmt.Println("Cannot create permissions for service " + *service)
		return
	}

	result, err := client.AddPermission(context.Background(), permArgs)
	if err != nil {
		fmt.Println("Cannot configure function for notifications")
		return
	}

	fmt.Println(result)
}
