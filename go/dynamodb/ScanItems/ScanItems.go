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
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	// github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue

	// "github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

// Item holds info about the new item
type Item struct {
	Year   int
	Title  string
	Plot   string
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
	filt := expression.Name("Year").Equal(expression.Value(year))

	// Or we could get the movies by ratings and pull out those with the right year later
	//    filt := expression.Name("info.rating").GreaterThan(expression.Value(min_rating))

	// Get back the title, year, and rating
	proj := expression.NamesList(expression.Name("Title"), expression.Name("Year"), expression.Name("Rating"))

	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		fmt.Println("Got error building expression:")
		fmt.Println(err)
		return
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 table,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Using the Config value, create the DynamoDB client
	// Create a new DynamoDB Service Client
	client := dynamodb.NewFromConfig(cfg)

	result, err := ScanTableItems(context.Background(), client, input)
	if err != nil {
		fmt.Println("Got an error scanning table:")
		fmt.Println(err)
		return
	}

	var items []Item

	for _, i := range result.Items {
		item := Item{}

		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err)
			return
		}

		// Which ones had a higher rating than the minimum value?
		if item.Rating > *minRating {
			// Or it we had filtered by rating previously:
			//   if item.Year == year {
			items = append(items, item)
		}
	}

	fmt.Println("Found", strconv.Itoa(len(items)), "movie(s) with a rating above", *minRating, "in", *year)
}

// snippet-end:[dynamodb.go.scan_table_items]
