package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	//	"github.com/aws/aws-lambda-go/events"
	//	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

//type item map[string]types.AttributeValue

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

// AddDataToTable adds the name/tag values in entries to table
func AddDataToTable(table string, key string, entries []Entry) error {
	numItems := len(entries) + 1
	attrs := make(map[string]types.AttributeValue, numItems)

	attrs["path"] = &types.AttributeValueMemberS{
		Value: key,
	}

	for _, e := range entries {
		if e.entryName != "" {
			attrs[e.entryName] = &types.AttributeValueMemberS{
				Value: e.entryTag,
			}
		}
	}

	// (import "github.com/aws/aws-sdk-go-v2/config")
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

// SaveMetadata gets the ELIF info from key "uploads/*.[jpg | png] and stores in table
func SaveMetadata(test bool, bucket string, key string, table string) error {
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

	if test {
		for _, e := range entries {
			if e.entryName != "" {
				fmt.Println(e.entryName + " == " + e.entryTag)
			}
		}

		return nil
	}

	err = AddDataToTable(table, key, entries)

	return err
}

func main() {
	bucket := flag.String("b", "", "The name of the bucket containing the photo") // doug-test-imagerecog
	key := flag.String("k", "", "The name of the photo")                          //
	table := flag.String("t", "", "The table to add the info to")                 // ImageRecognition
	test := flag.Bool("x", false, "Whether to just barf out the name/value pairs")

	flag.Parse()

	if *bucket == "" || *key == "" || *table == "" {
		fmt.Println("You must specify a bucket, key (JPG or PNG with 'uploads' prefix), and table (-b BUCKET -k KEY -t table)")
		return
	}

	err := SaveMetadata(*test, *bucket, *key, *table)
	if err != nil {
		msg := "Got error saving metadata from key '" + *key + "' in bucket '" + *bucket + "' to table ' " + *table + "':"
		fmt.Println(msg)
		fmt.Println(err)

		return
	}

	if !*test {
		msg := "Saved metadata from key '" + *key + "' in bucket '" + *bucket + "' to table ' " + *table + "':"
		fmt.Println(msg)
	}
}
