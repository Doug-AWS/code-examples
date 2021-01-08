package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	// "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	//"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

// Item holds info about the items returned by Scan
type Item struct {
	Year   int
	Title  string
	Rating float64
}

// Get the movies with a minimum rating of 8.0 in 2011
func main() {
	tableName := "Movies"
	minRating := 4.0
	year := 2013

	// Create the Expression to fill the input struct with.
	// Get all movies in that year; we'll pull out those with a higher rating later
	filt := expression.Name("Year").Equal(expression.Value(year))

	// Or we could get by ratings and pull out those with the right year later
	//    filt := expression.Name("info.rating").GreaterThan(expression.Value(min_rating))

	// Get back the title, year, and rating
	proj := expression.NamesList(expression.Name("Title"), expression.Name("Year"), expression.Name("Rating"))

	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		fmt.Println("Got error building expression:")
		fmt.Println(err.Error())
		return
	}

	// Build the query input parameters
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Using the Config value, create the DynamoDB client
	// Create a new DynamoDB Service Client
	client := dynamodb.NewFromConfig(cfg)

	var items []Item

	resp, err := client.Scan(context.Background(), input)
	if err != nil {
		fmt.Println("Got an error scanning the table:")
		fmt.Println(err.Error())
		return
	}

	itms := []Item{}

	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &itms)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
	}

	items = append(items, itms...)

	var goodItems []Item

	for _, item := range items {
		// Which ones had a higher rating than minimum?
		if item.Rating > minRating {
			// Or it we had filtered by rating previously:
			//   if item.Year == year {
			goodItems = append(goodItems, item)

			fmt.Println("Title: ", item.Title)
			fmt.Println("Rating:", item.Rating)
			fmt.Println()
		}
	}

	numItems := strconv.Itoa(len(goodItems))

	fmt.Println("Found", numItems, "movie(s) with a rating above", minRating, "in", year)
}
