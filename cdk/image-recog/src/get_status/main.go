package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context) (string, error) {
	return "{ \"status\": \"SUCCEEDED\" }", nil
}

func main() {
	lambda.Start(handler)
}
