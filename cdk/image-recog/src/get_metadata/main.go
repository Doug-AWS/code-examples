package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

type item map[string]types.AttributeValue

// Entry defines an exif name/value pair
type Entry struct {
	entryName string
	entryTag  string
}

var entries []Entry

// Printer defines a struct
type Printer struct{}

// Walk traverses the image metadata
func (p Printer) Walk(name exif.FieldName, tag *tiff.Tag) error {
	e := Entry{
		entryName: string(name),
		entryTag:  fmt.Sprintf("%s", tag),
	}

	entries = append(entries, e)

	return nil
}

func getMetadata(bucket string, key string) error {
	// Make sure key ends in JPG or PNG
	parts := strings.Split(key, ".")

	if len(parts) < 2 {
		msg := "Could not split '" + key + "' into name/extension"
		return errors.New(msg)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		msg := "Got configuration error loading context: " + err.Error()
		return errors.New(msg)
	}

	s3Client := s3.NewFromConfig(cfg)

	s3Input := &s3.GetObjectInput{
		Bucket: &bucket, Key: &key,
	}

	s3Resp, err := s3Client.GetObject(context.TODO(), s3Input)
	if err != nil {
		msg := "Got error calling GetObject: " + err.Error()
		return errors.New(msg)
	}

	x, err := exif.Decode(s3Resp.Body)
	if err != nil {
		msg := "Got error decoding exif data: " + err.Error()
		return errors.New(msg)
	}

	entries = make([]Entry, 1)

	var p Printer
	x.Walk(p)

	attrs := make(map[string]*types.AttributeValue, len(entries))

	for _, e := range entries {
		if e.entryName != "" {
			attrs[e.entryName] = &types.AttributeValue{
				S: aws.String(e.entryTag),
			}
		}
	}

	// Add entries to DynamoDB table
	// Get table name from environment

	dynamodbClient := dynamodb.NewFromConfig(cfg)

	dynamodbInput := &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("tableName")),
		Item:      attrs,
	}

	_, err = dynamodbClient.PutItem(context.TODO(), dynamodbInput)
	if err != nil {
		msg := "Got error calling PutItem: " + err.Error()
		return errors.New(msg)
	}

	return nil
}

func handler(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3 := record.S3
		err := getMetadata(s3.Bucket.Name, s3.Object.Key)
		if err != nil {
			msg := "Did not get metadata from key '" + s3.Object.Key + "' in bucket '" + s3.Bucket.Name + "'"
			fmt.Println(msg)
		}
	}
}

func main() {
	lambda.Start(handler)
}
