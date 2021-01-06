// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func main() {
	zipFile := flag.String("z", "", "The name of the ZIP file (without the .zip extension)")
	bucket := flag.String("b", "", "the name of bucket to which the ZIP file is uploaded")
	function := flag.String("f", "", "The name of the Lambda function")
	packageClass := flag.String("p", "", "The name of the package.class handling the call")
	roleARN := flag.String("r", "", "The ARN of the role that calls the function")

	flag.Parse()

	if *zipFile == "" || *bucket == "" || *function == "" || *packageClass == "" || *roleARN == "" {
		fmt.Println("You must supply a zip file name, bucket name, function name, handler, and role ARN:")
		fmt.Println("-z ZIPFILE -b BUCKET -f FUNCTION -p PACKAGE.CLASS -r ROLE-ARN")
		return
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got an error loading the configuration")
		return
	}

	client := lambda.NewFromConfig(cfg)

	contents, err := ioutil.ReadFile(*zipFile + ".zip")
	if err != nil {
		fmt.Println("Could not read " + *zipFile + ".zip")
		return
	}

	createCode := &types.FunctionCode{
		S3Bucket:        bucket,
		S3Key:           zipFile,
		S3ObjectVersion: aws.String(""),
		ZipFile:         contents,
	}

	createArgs := &lambda.CreateFunctionInput{
		Code:         createCode,
		FunctionName: function,
		Handler:      packageClass,
		Role:         roleARN,
		Runtime:      types.RuntimeGo1x,
	}

	result, err := client.CreateFunction(context.Background(), createArgs)
	if err != nil {
		fmt.Println("Cannot create function: " + err.Error())
		return
	}

	fmt.Println(result)
}
