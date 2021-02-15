package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
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

/* KeyInfo defines the key of the item to update
type KeyInfo struct {
	Path string
}
*/

/* RekognitionDetectLabelsAPI defines the interface for the DetectLabels function
type RekognitionDetectLabelsAPI interface {
	DetectLabels(ctx context.Context,
		params *rekognition.DetectLabelsInput,
		optFns ...func(*rekognition.Options)) (*rekognition.DetectLabelsOutput, error)
}
*/

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

/* GetLabelInfo retrieves information about the labels
func GetLabelInfo(c context.Context, api RekognitionDetectLabelsAPI, input *rekognition.DetectLabelsInput) (*rekognition.DetectLabelsOutput, error) {
	resp, err := api.DetectLabels(c, input)

	return resp, err
}
*/

/* UpdateTableItem appends the data in entries to the table item with path key
func UpdateTableItem(debug bool, table string, key string, entries []Entry) error {
	keyAttr := make(map[string]dbTypes.AttributeValue, 1)
	keyAttr["path"] = &dbTypes.AttributeValueMemberS{
		Value: key,
	}

	numItems := len(entries)
	attrs := make(map[string]dbTypes.AttributeValue, numItems)

	for _, e := range entries {
		if e.Label != "" {
			attrs["Label"] = &dbTypes.AttributeValueMemberS{
				Value: e.Label,
			}
			attrs["Confidence"] = &dbTypes.AttributeValueMemberS{
				Value: e.Confidence,
			}
		}
	}

	// (import "github.com/aws/aws-sdk-go-v2/config")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)

	dynamodbInput := &dynamodb.UpdateItemInput{
		Key:       keyAttr,
		TableName: aws.String(table),
	}

	_, err = dynamodbClient.UpdateItem(context.TODO(), dynamodbInput)
	if err != nil {
		msg := "Got error calling PutItem: " + err.Error()
		return errors.New(msg)
	}

	return nil
}
*/

// UpdateTableItem appends the data in entries to the table item with path == key
func UpdateTableItem(debug bool, table string, key string, entries []Entry) error {
	l := len(entries)

	if debug {
		fmt.Println("Got " + strconv.Itoa(l) + " entries:")
		for _, e := range entries {
			if e.Label == "" {
				continue
			}

			fmt.Println("Label:      " + e.Label)
			fmt.Println("Confidence: " + e.Confidence)
		}

		fmt.Println()
	}

	if l < 2 {
		msg := "Got only " + strconv.Itoa(l) + " entries"
		return errors.New(msg)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got an error loading configuration")
		return err
	}

	client := dynamodb.NewFromConfig(cfg)

	// We're adding these to the same table item,
	// so specify the key, based on the original filename
	// If key doesn't start with uploads/, prefix it with that
	parts := strings.Split(key, "/")

	if len(parts) != 2 {
		key = "uploads/" + key
	}

	keyAttr := make(map[string]dbTypes.AttributeValue, 1)
	keyAttr["path"] = &dbTypes.AttributeValueMemberS{
		Value: key,
	}

	/*
		tableKey := KeyInfo{
			Path: path,
		}

		key, err := attributevalue.MarshalMap(tableKey)
		if err != nil {
			fmt.Println("Got error marshalling path")
			return err
		}
	*/

	// ExpressionAttributeValues map[string]types.AttributeValue
	exprAttrs := make(map[string]dbTypes.AttributeValue, 2)

	for _, e := range entries {
		if e.Label == "" {
			continue
		}

		/*
					"UpdateExpression": "set Replies = Replies + :num",
			        "ExpressionAttributeValues": {
			            ":num": {"N": "1"}
		*/

		//exprAttrs[":label"] = &dbTypes.AttributeValueMemberS{
		//	Value: e.Label,
		//}

		exprAttrs[":confidence"] = &dbTypes.AttributeValueMemberS{
			Value: e.Confidence,
		}

		expr := "set " + e.Label + "-Confidence = :confidence"

		debugPrint(debug, "Setting Label: "+e.Label+" and Confidence == "+e.Confidence+" to table "+table+" for item: "+key)

		input := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(table),
			Key:                       keyAttr,
			UpdateExpression:          aws.String(expr),
			ExpressionAttributeValues: exprAttrs,
			ReturnValues:              "ALL_NEW",
		}

		_, err = client.UpdateItem(context.TODO(), input)
		if err != nil {
			fmt.Println("Got error calling UpdateItem: ")
			return err
		}
	}

	return nil
}

func main() {
	bucket := flag.String("b", "doug-test-imagerecog", "The name of the bucket containing the photo") // doug-test-imagerecog
	key := flag.String("k", "uploads/DonaldTrumpTheLemmingLeader.jpg", "The name of the photo")       //
	table := flag.String("t", "", "The table to add the info to")                                     // ImageRecognition
	debug := flag.Bool("d", false, "Whether to barf out more info")
	//region := flag.String("r", "us-west-2", "The region")
	flag.Parse()

	if *bucket == "" || *key == "" {
		fmt.Println("You must specify a bucket and key (JPG or PNG); table to store info (-b BUCKET -k KEY -t table)")
		return
	}

	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Using the Config value, create the Rekognition client
	client := rekognition.NewFromConfig(cfg)

	input := &rekognition.DetectLabelsInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: bucket,
				Name:   key,
			},
		},
	}

	// Build the request with its input parameters
	resp, err := client.DetectLabels(context.TODO(), input)

	//resp, err := GetLabelInfo(context.TODO(), client, input)
	if err != nil {
		panic("failed to get labels, " + err.Error())
	}

	fmt.Println("Info about " + *key + ":")

	var entries = make([]Entry, 1)

	for _, label := range resp.Labels {
		c := fmt.Sprintf("%f", *label.Confidence)
		e := Entry{
			Label:      *label.Name,
			Confidence: c,
		}

		entries = append(entries, e)

		/*
			fmt.Println("Label: " + *label.Name)
			s := fmt.Sprintf("%f", *label.Confidence)
			if err == nil {
				fmt.Println("  Confidence: " + s)
			}

			fmt.Println()
		*/
	}

	if *table != "" {
		err := UpdateTableItem(*debug, *table, *key, entries)
		if err != nil {
			fmt.Println("Got an error saving info to table " + *table + ":")
			fmt.Println(err)
		}
	}
}
