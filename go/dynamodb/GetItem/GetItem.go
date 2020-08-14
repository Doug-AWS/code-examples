package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// Item defines the item for the table
type Item struct {
	Year   int
	Title  string
	Plot   string
	Rating float64
}

// GetTableItem retrieves the item with the year and title from the table
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     table is the name of the table
//     title is the movie title
//     year is when the movie was released
// Output:
//     If success, the information about the table item and nil
//     Otherwise, nil and an error from the call to GetItem or UnmarshalMap
func GetTableItem(debug bool, svc dynamodbiface.DynamoDBAPI, table, title *string, year *int) (*Item, error) {
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: table,
		Key: map[string]*dynamodb.AttributeValue{
			"Year": {
				N: aws.String(strconv.Itoa(*year)),
			},
			"Title": {
				S: title,
			},
		},
	})
	if err != nil {
		return nil, err
	}

    item := Item{}
    
    if  

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func main() {
	table := flag.String("t", "", "The table to retrieve item from")
	title := flag.String("n", "", "The name of the movie")
	year := flag.Int("y", -1, "The year the movie was released")
	debug := flag.Bool("d", false, "Whether to barf out more info")
	flag.Parse()

	if *table == "" || *title == "" || *year == -1 {
		fmt.Println("You must supply a table name, movie title, and valid year")
		fmt.Println("(-t TABLE -n NAME -y YEAR")
		return
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	item, err := GetTableItem(svc, table, title, year)
	if err != nil {
		fmt.Println("Got an error retrieving the item:")
		fmt.Println(err)
		return
	}

	fmt.Println("Found item:")
	fmt.Println("Year:  ", item.Year)
	fmt.Println("Title: ", item.Title)
	fmt.Println("Plot:  ", item.Plot)
	fmt.Println("Rating:", item.Rating)
}
