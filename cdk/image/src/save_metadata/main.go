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

	// "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
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

func isNameValid(key string) bool {
	// Ignore anything that doesn't have upload prefix or end with jpg or png
	// Make sure key ends in JPG or PNG
	parts := strings.Split(key, ".")

	if len(parts) < 2 {
		return false
	}

	if parts[1] != "jpg" && parts[1] != "png" {
		return false
	}

	// Trap anything without upload/ prefix
	pieces := strings.Split(parts[0], "/")

	if pieces[0] != "uploads" {
		return false
	}

	return true
}

func addDataToTable(table string, key string, entries []Entry) error {
	numItems := len(entries) + 1
	attrs := make(map[string]*types.AttributeValue, numItems)

	attrs["path"] = &types.AttributeValue{
		S: aws.String(key),
	}

	for _, e := range entries {
		if e.entryName != "" {
			attrs[e.entryName] = &types.AttributeValue{
				S: aws.String(e.entryTag),
			}
		}
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)

	dynamodbInput := &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      attrs,
	}

	_, err = dynamodbClient.PutItem(context.TODO(), dynamodbInput)
	if err != nil {
		msg := "Got error calling PutItem: " + err.Error()
		return errors.New(msg)
	}

	return nil
}

func saveMetadata(bucket string, key string, table string) error {
	// Ignore anything that doesn't have upload prefix or end with jpg or png
	// Make sure key ends in JPG or PNG
	// uploads/filename.jpg -> uploads/filename, jpg
	parts := strings.Split(key, ".")

	if len(parts) < 2 {
		msg := "Could not split '" + key + "' into name/extension"
		return errors.New(msg)
	}

	if parts[1] != "jpg" && parts[1] != "png" {
		msg := "Extension '" + parts[1] + "' is not jpg or png"
		return errors.New(msg)
	}

	// Trap anything without uploads/ prefix
	pieces := strings.Split(parts[0], "/")

	if pieces[0] != "uploads" {
		msg := key + " does not have uploads/ prefix"
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
	err = x.Walk(p)
	if err != nil {
		return err
	}

	err = addDataToTable(table, key, entries)

	return err
}
func handler(ctx context.Context, s3Event events.S3Event) (string, error) {
	numEvents := len(s3Event.Records)

	if numEvents < 1 {
		return "", errors.New("Save metadata function got an empty S3 event")
	}

	fmt.Println("Got event in save ELIF data handler:")
	fmt.Println(s3Event)

	s3 := s3Event.Records[0].S3

	// Get table name from environment
	table := os.Getenv("tableName")

	err := saveMetadata(s3.Bucket.Name, s3.Object.Key, table)
	if err != nil {
		msg := "Got error saving metadata from key '" + s3.Object.Key + "' in bucket '" + s3.Bucket.Name + "':"
		fmt.Println(msg)
		fmt.Println(err)

		return "", err
	}

	msg := "Saved metadata from key '" + s3.Object.Key + "' in bucket '" + s3.Bucket.Name + "'"
	fmt.Println(msg)

	output := "{ \"Bucket\": " + s3.Bucket.Name + ", \"Key\": " + s3.Object.Key + " }"
	fmt.Println("Returning: ")
	fmt.Println(output)

	return output, nil
}

func main() {
	lambda.Start(handler)
}