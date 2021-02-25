package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go/aws"
)

// MyEvent is the event we receive
type MyEvent struct {
	Bucket      string
	Key         string
	WaitTimeout int `json:"waitTimeout"`
}

/*
	ExecutedVersion string `json:"ExecutedVersion"`
	Payload         string `json:"Payload"`
	SdkHTTPMetadata struct {
		AllHTTPHeaders struct {
			XAmzExecutedVersion        []string `json:"X-Amz-Executed-Version"`
			XAmznRemappedContentLength []string `json:"x-amzn-Remapped-Content-Length"`
			Connection                 []string `json:"Connection"`
			XAmznRequestID             []string `json:"x-amzn-RequestId"`
			ContentLength              []string `json:"Content-Length"`
			Date                       []string `json:"Date"`
			XAmznTraceID               []string `json:"X-Amzn-Trace-Id"`
			ContentType                []string `json:"Content-Type"`
		} `json:"AllHttpHeaders"`
		HTTPHeaders struct {
			Connection                 string `json:"Connection"`
			ContentLength              string `json:"Content-Length"`
			ContentType                string `json:"Content-Type"`
			Date                       string `json:"Date"`
			XAmzExecutedVersion        string `json:"X-Amz-Executed-Version"`
			XAmznRemappedContentLength string `json:"x-amzn-Remapped-Content-Length"`
			XAmznRequestID             string `json:"x-amzn-RequestId"`
			XAmznTraceID               string `json:"X-Amzn-Trace-Id"`
		} `json:"HttpHeaders"`
		HTTPStatusCode int `json:"HttpStatusCode"`
	} `json:"SdkHttpMetadata"`
	SdkResponseMetadata struct {
		RequestID string `json:"RequestId"`
	} `json:"SdkResponseMetadata"`
	StatusCode int `json:"StatusCode"`

*/

// Entry defines an item we add to the db
type Entry struct {
	Label      string
	Confidence string
}

// KeyInfo defines the key of the item to update
type KeyInfo struct {
	Path string
}

func addDataToTable(table string, eventKey string, entries []Entry) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got an error loading configuration")
		return err
	}

	client := dynamodb.NewFromConfig(cfg)

	// We're adding these to the same table item,
	// so specify the key, based on the original filename
	// If eventKey doesn't start with uploads/, prefix it with that
	parts := strings.Split(eventKey, "/")

	if len(parts) != 2 {
		eventKey = "uploads/" + eventKey
	}

	key := make(map[string]*dbTypes.AttributeValue, 1)
	key["path"] = &dbTypes.AttributeValue{
		S: aws.String(eventKey),
	}

	attr := make(map[string]*dbTypes.AttributeValue, 1)

	for _, e := range entries {
		attr[e.Label] = &dbTypes.AttributeValue{
			S: aws.String(e.Confidence),
		}

		input := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(table),
			Key:                       key,
			UpdateExpression:          aws.String(""),
			ExpressionAttributeValues: attr,
		}

		_, err = client.UpdateItem(context.TODO(), input)
		if err != nil {
			fmt.Println("Got error calling UpdateItem: ")
			return err
		}
	}

	return nil
}

func handler(ctx context.Context, myEvent MyEvent) (string, error) {
	fmt.Println("Got event in save object data event handler:")
	fmt.Println(myEvent)

	/* Get bucket and key names from environment
	bucketName := os.Getenv("bucketName")
	keyName := os.Getenv("keyName")
	*/

	bucketName := myEvent.Bucket
	keyName := myEvent.Key

	if bucketName == "" || keyName == "" {
		msg := "Did not get bucket and key"
		return "", errors.New(msg)
	}

	fmt.Println("Got bucket name '" + bucketName + "' from environment variable")
	fmt.Println("Got key name    '" + keyName + "' from environment variable")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Unable to load SDK config")
		return "", err
	}

	// Using the Config value, create the Rekognition client
	client := rekognition.NewFromConfig(cfg)

	input := &rekognition.DetectLabelsInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: &bucketName,
				Name:   &keyName,
			},
		},
	}

	resp, err := client.DetectLabels(context.TODO(), input)
	if err != nil {
		return "", err
	}

	var entries = make([]Entry, 1)

	for _, label := range resp.Labels {
		c := fmt.Sprintf("%f", *label.Confidence)
		e := Entry{
			Label:      *label.Name,
			Confidence: c,
		}

		entries = append(entries, e)
	}

	table := os.Getenv("tableName")
	fmt.Println("Got table name  '" + table + "' from environment variable")

	err = addDataToTable(table, keyName, entries)
	if err != nil {
		return "", err
	}

	myEvent.WaitTimeout = 5
	fmt.Println("Returning: ")
	fmt.Println(myEvent)

	output, err := json.Marshal(&myEvent)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func main() {
	lambda.Start(handler)
}
