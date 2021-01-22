package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, s3Event events.S3Event) {
	fmt.Println("Hanlder function called")

		for _, record := range s3Event.Records {
			s3 := record.S3
			fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3.Bucket.Name, s3.Object.Key)
		}
}

func main() {
	lambda.Start(handler)
}
