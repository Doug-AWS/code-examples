package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"

	// "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue" // ???
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go/aws"
)

// Entry defines an item we add to the db
type Entry struct {
	Label      string
	Confidence string
}

// MyEvent defines the event we get
type MyEvent struct {
	Bucket string
	Key    string
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

	/*
		tableKey := KeyInfo{
			Path: eventKey,
		}

		key, err := attributevalue.MarshalMap(tableKey)
		if err != nil {
			fmt.Println("Got error marshalling path")
			return err
		}
	*/

	attr := make(map[string]*dbTypes.AttributeValue, 1)

	for _, e := range entries {
		/*
			expr, err := attributevalue.MarshalMap(e)
			if err != nil {
				fmt.Println("Got error marshalling item")
				return err
			}
		*/
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

func handler(ctx context.Context, event MyEvent) (string, error) {
	fmt.Println("Got event in save Rekognition event handler:")
	fmt.Println(event)

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
				Bucket: &event.Bucket,
				Name:   &event.Key,
			},
		},
	}

	resp, err := client.DetectLabels(context.TODO(), input)

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

	err = addDataToTable(table, event.Key, entries)

	return "{ \"Bucket\": " + event.Bucket + ", \"Key\": " + event.Key + " }", err
}

func main() {
	lambda.Start(handler)
}
