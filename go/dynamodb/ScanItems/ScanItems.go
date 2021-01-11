package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Item holds info about the items returned by Scan
type Item struct {
	Title string
	Info  struct {
		Rating float64
	}
}

// Get the movies above a minimum rating in a specific year.
func main() {
	table := flag.String("t", "", "The name of the table to scan.")
	rating := flag.Float64("r", -1.0, "The minimum rating for a movie to retrieve.")
	year := flag.Int("y", 1899, "The year when the movie was released.")
	verbose := flag.Bool("v", false, "Whether to show info about the movie.")

	flag.Parse()

	if *table == "" || *rating < 0.0 || *year < 1900 {
		fmt.Println("You must supply the name of the table, a rating above zero, and a year after 1900:")
		fmt.Println("-t TABLE -r RATING -y YEAR")
		return
	}

	// Get all movies in that year.
	filt1 := expression.Name("year").Equal(expression.Value(*year))
	// Get movies with the rating above the minimum.
	filt2 := expression.Name("info.rating").GreaterThan(expression.Value(*rating))

	// Get back the title and rating (we know the year)
	proj := expression.NamesList(expression.Name("title"), expression.Name("info.rating"))

	expr, err := expression.NewBuilder().WithFilter(filt1).WithFilter(filt2).WithProjection(proj).Build()
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
		TableName:                 table,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Using the Config value, create the DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	var items []Item

	resp, err := client.Scan(context.Background(), input)
	if err != nil {
		fmt.Println("Got an error scanning the table:")
		fmt.Println(err.Error())
		return
	}

	itms := []Item{}

	err = attributevalue.UnmarshalListOfMaps(resp.Items, &itms)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
	}

	items = append(items, itms...)

	for _, item := range items {
		if *verbose {
			fmt.Println("Title: ", item.Title)
			fmt.Println("Rating:", item.Info.Rating)
			fmt.Println()
		}
	}

	numItems := strconv.Itoa(len(items))

	fmt.Println("Found", numItems, "movie(s) with a rating above", *rating, "in", *year)
}
