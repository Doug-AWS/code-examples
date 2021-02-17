package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

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

func quickSort(arr *[]Entry, start, end int) []Entry {
	if start < end {
		partitionIndex := partition(*arr, start, end)
		quickSort(arr, start, partitionIndex-1)
		quickSort(arr, partitionIndex+1, end)
	}
	return *arr
}

func partition(arr []Entry, start, end int) int {
	pivot := arr[end].entryName
	pIndex := start
	for i := start; i < end; i++ {
		if arr[i].entryName <= pivot {
			//  swap
			arr[i], arr[pIndex] = arr[pIndex], arr[i]
			pIndex++
		}
	}
	arr[pIndex], arr[end] = arr[end], arr[pIndex]
	return pIndex
}

func addDataToTable(file string, table string, entries []Entry) error {
	numItems := len(entries) + 1
	attrs := make(map[string]types.AttributeValue, numItems)

	attrs["path"] = &types.AttributeValueMemberS{
		Value: "uploads/" + file,
	}

	for _, e := range entries {
		if e.entryName != "" {
			attrs[e.entryName] = &types.AttributeValueMemberS{
				Value: e.entryTag,
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

func main() {
	file := flag.String("f", "", "The name of the JPG or PNG file to get ELIF info from")
	table := flag.String("t", "", "The name of the table to store info into")
	flag.Parse()

	if *file == "" {
		fmt.Println("You must supply a filename (-f FILENAME)")
		return
	}

	f, err := os.Open(*file)
	if err != nil {
		fmt.Println("Got an error opening " + *file + ":")
		fmt.Println(err)
		return
	}

	x, err := exif.Decode(f)
	if err != nil {
		fmt.Println("Got an error decoding EXIF info:")
		fmt.Println(err)
		return
	}

	entries = make([]Entry, 1)

	var p Printer
	x.Walk(p)

	if *table != "" {
		err := addDataToTable(*file, *table, entries)
		if err != nil {
			fmt.Println("Got an error adding data to table " + *table + ":")
			fmt.Println(err)
		}
	}

	// Sort entries (does it make any difference when we add to table?)
	quickSort(&entries, 0, len(entries)-1)

	// Barf out entries:
	for _, e := range entries {
		if e.entryName != "" {
			fmt.Printf("%40s: %s\n", e.entryName, e.entryTag)
		}
	}
}
