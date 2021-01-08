// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
// snippet-start:[dynamodb.gov2.scan_table_items]
package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	// github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
)

// Item holds info about each item that scan returns
type Item struct {
	Year   int
	Title  string
	Rating float64
}

// DynamodbScanAPI defines the interface for the Scan function.
// We use this interface to test the function using a mocked service.
type DynamodbScanAPI interface {
	Scan(ctx context.Context,
		params *dynamodb.ScanInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// ScanTableItems retrieves the Amazon Dynamodb table items that match the input parameters.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If successful, a ScanOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to Scan.
func ScanTableItems(c context.Context, api DynamodbScanAPI, input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	result, err := api.Scan(c, input)

	return result, err
}

// Get the movies with a minimum rating of minRating in year
func main() {
	table := flag.String("t", "", "The name of the table")
	minRating := flag.Float64("r", -1.0, "The minimum rating of the movies to retrieve")
	year := flag.Int("y", -1, "The year the movies to retrieve were released")
	flag.Parse()

	if *table == "" || *minRating < 0.0 || *year < 0 {
		fmt.Println("You must supply a table name, minimum rating of 0.0, and year > 0 but < 2020")
		fmt.Println("(-t TABLE -r RATING -y YEAR)")
		return
	}
	// Create the expression to fill the input struct.
	// Get all movies in that year; we'll pull out those with a higher rating later
	filt := "Year = :val"

	// Or we could get the movies by ratings and pull out those with the right year later
	//    "info.rating > :val"

	// Get back the title, year, and rating
	proj := "Title, Year, Rating"

	// Value for expression
	av := types.AttributeValue{
		N: year,
	}

	valueMap := make(map[string]types.AttributeValue)
	valueMap[":val"] = av

	input := &dynamodb.ScanInput{
		ExpressionAttributeValues: valueMap,
		FilterExpression:          &filt,
		ProjectionExpression:      &proj,
		TableName:                 table,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Using the Config value, create the DynamoDB client
	// Create a new DynamoDB Service Client
	client := dynamodb.NewFromConfig(cfg)

	// Dictionary<string, AttributeValue> lastKeyEvaluated = null
	lastKeyEvaluated := make(map[string]types.AttributeValue)
	done := false
	var items []Item

	for !done {
		result, err := ScanTableItems(context.Background(), client, input)
		if err != nil {
			fmt.Println("Got an error scanning table:")
			fmt.Println(err)
			return
		}

		lastKeyEvaluated = result.LastEvaluatedKey

		if len(lastKeyEvaluated) == 0 {
			done = true
			break
		}

		// Items is an array of maps of [string]types.AttributeValue
		gotYear := false
		gotTitle := false
		gotRating := false

		for _, item := range result.Items {
			var i Item

			for key, value := range item {
				var union types.AttributeValue
				// type switches can be used to check the union value
				switch v := union.(type) {
				case *types.AttributeValueMemberN:
					if key == "year" {
						i.Year, err = strconv.Atoi(v.Value)
						if err != nil {
							fmt.Println("Got an error converting year " + v.Value + " to an int")
							return
						}

						gotYear = true
					}

					if key == "rating" {
						f, err := strconv.ParseFloat(v.Value, 64)
						if err != nil {
							fmt.Println("Got an error converting rating " + v.Value + " to a float")
							return
						}

						i.Rating = f

						gotRating = true
					}

				case *types.AttributeValueMemberS:
					if key == "title" {
						i.Title = v.Value
						gotTitle = true
					}
				}
			}

			if gotYear && gotTitle && gotRating {
				if i.Rating > *minRating {
					items = append(items, i)
				}
			}
		}
	}

	fmt.Println("Found", strconv.Itoa(len(items)), "movie(s) with a rating above", *minRating, "in", *year)
}

// snippet-end:[dynamodb.go.scan_table_items]
