package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rTypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nfnt/resize"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

// ExifEntry defines an exif name/value pair
type ExifEntry struct {
	entryName string
	entryTag  string
}

var exifEntries []ExifEntry

// DataEntry defines a rekognition name/value pair
type DataEntry struct {
	Label      string
	Confidence string
}

//var dataEntries []DataEntry

// Printer defines a struct
type Printer struct{}

// Walk traverses the image metadata
func (p Printer) Walk(name exif.FieldName, tag *tiff.Tag) error {
	e := ExifEntry{
		entryName: string(name),
		entryTag:  fmt.Sprintf("%s", tag),
	}

	exifEntries = append(exifEntries, e)

	return nil
}

func fileIsValid(fileName string) error {
	// Make sure file extension is jpg or png
	parts := strings.Split(fileName, ".")

	if len(parts) < 2 {
		msg := fileName + " has no file extension"
		return errors.New(msg)
	}

	if parts[1] != "jpg" && parts[1] != "png" {
		msg := fileName + " has neither a jpg nor png file extension"
		return errors.New(msg)
	}

	return nil
}

func addMetaDataToTable(tableName string, key string, entries []ExifEntry) error {
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

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)

	dynamodbInput := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      attrs,
	}

	_, err = dynamodbClient.PutItem(context.TODO(), dynamodbInput)
	if err != nil {
		msg := "Got error calling PutItem: " + err.Error()
		return errors.New(msg)
	}

	return nil
}

// Appends the data in entries to the table item with path == key
func addObjectDataToTable(tableName string, key string, entries []DataEntry) error {
	l := len(entries)

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

	keyAttr := make(map[string]types.AttributeValue, 1)
	keyAttr["path"] = &types.AttributeValueMemberS{
		Value: key,
	}

	// ExpressionAttributeValues map[string]types.AttributeValue
	exprAttrs := make(map[string]types.AttributeValue, 2)

	for _, e := range entries {
		if e.Label == "" {
			continue
		}

		exprAttrs[":confidence"] = &types.AttributeValueMemberS{
			Value: e.Confidence,
		}

		expr := "set " + e.Label + "Confidence = :confidence"

		input := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tableName),
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

func saveMetadata(bucketName string, fileName string, tableName string) error {
	file, err := os.Open(fileName)

	if err != nil {
		msg := "Unable to open file " + fileName
		return errors.New(msg)
	}

	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		return err
	}

	var p Printer
	err = x.Walk(p)
	if err != nil {
		return err
	}

	err = addMetaDataToTable(tableName, fileName, exifEntries)

	return err
}

// RekognitionDetectLabelsAPI defines the interface for the DetectLabels function.
// We use this interface to test the function using a mocked service.
type RekognitionDetectLabelsAPI interface {
	DetectLabels(ctx context.Context,
		params *rekognition.DetectLabelsInput,
		optFns ...func(*rekognition.Options)) (*rekognition.DetectLabelsOutput, error)
}

// GetLabels retrieves the lables in a jpg or png image.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If successful, a DetectLablesOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to DetectLabels.
func GetLabels(c context.Context, api RekognitionDetectLabelsAPI, input *rekognition.DetectLabelsInput) (*rekognition.DetectLabelsOutput, error) {
	resp, err := api.DetectLabels(c, input)

	return resp, err
}

// saveObjectData saves the labels found by Rekognition in the table
func saveObjectData(labels []rTypes.Label, tableName string, fileName string) error {
	var entries = make([]DataEntry, 1)

	for _, label := range labels {
		if *label.Name == "" {
			continue
		}

		c := fmt.Sprintf("%f", *label.Confidence)
		e := DataEntry{
			Label:      *label.Name,
			Confidence: c,
		}

		entries = append(entries, e)
	}

	err := addObjectDataToTable(tableName, fileName, entries)
	if err != nil {
		return err
	}

	return nil
}

func saveFile(bucketName string, fileName string) error {
	file, err := os.Open(fileName)

	if err != nil {
		msg := "Unable to open file " + fileName
		return errors.New(msg)
	}

	defer file.Close()

	// Add "uploads/ prefix to filename"
	uploadedFileName := "uploads/" + fileName

	input := &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &uploadedFileName,
		Body:   file,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		msg := "Got configuration error loading context: " + err.Error()
		return errors.New(msg)
	}

	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		msg := "Got error uploading file:"
		return errors.New(msg)
	}

	return nil
}

func calculateRatioFit(srcWidth, srcHeight int, maxWidth, maxHeight float64) (int, int) {
	ratio := math.Min(maxWidth/float64(srcWidth), maxHeight/float64(srcHeight))
	return int(math.Ceil(float64(srcWidth) * ratio)), int(math.Ceil(float64(srcHeight) * ratio))
}

func putObject(bucketName, fileName string, body io.Reader) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)

	input := &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &fileName,
		Body:   body,
	}

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

// Returns the name of the thumbnail file in the S3 bucket
func createThumbnail(bucketName string, fileName string) (string, error) {
	file, err := os.Open(fileName)

	if err != nil {
		msg := "Unable to open file " + fileName
		return "", errors.New(msg)
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	b := img.Bounds()
	width := b.Max.X
	height := b.Max.Y

	fmt.Println("Original width:  " + strconv.Itoa(width))
	fmt.Println("Original height: " + strconv.Itoa(height))

	// Keep width/height ratio
	var maxWidth float64 = 80
	var maxHeight float64 = 80

	w, h := calculateRatioFit(width, height, maxWidth, maxHeight)

	fmt.Println("Thumbnail width:  " + strconv.Itoa(w))
	fmt.Println("Thumbnail height: " + strconv.Itoa(h))

	// Call the resize library for image scaling
	m := resize.Resize(uint(w), uint(h), img, resize.Lanczos3)

	// Create new object name from existing filename
	parts := strings.Split(fileName, ".")
	name := parts[0] + "thumb." + parts[1]

	// Body of S3 object
	var buf bytes.Buffer

	// save the file in JPG or PNG format
	switch parts[1] {
	case "jpg":
		err := jpeg.Encode(&buf, m, nil)
		if err != nil {
			fmt.Println("Got an error calling jpeg.Encode")
			return "", err
		}

		break
	case "png":
		err = png.Encode(&buf, m)
		if err != nil {
			fmt.Println("Got an error calling png.Encode")
			return "", err
		}

		break

	default:
		msg := "Unsupported format: " + parts[1]
		return "", errors.New(msg)
	}

	// Create S3 object with name == name and body == buf
	r := bytes.NewReader(buf.Bytes())

	// Add thumbs/ prefix so we can find all of them in thumbs/
	err = putObject(bucketName, "thumbs/"+name, io.Reader(r))
	return name, err
}

func main() {
	bucketName := flag.String("b", "", "The bucket to upload the file to")
	tableName := flag.String("t", "", "The table to store image data in")
	fileName := flag.String("f", "", "The file to upload")

	flag.Parse()

	if *bucketName == "" || *tableName == "" || *fileName == "" {
		fmt.Println("You must supply a bucket, table, and file to upload (-b BUCKET -t TABLE -f FILE)")
		return
	}

	// Make sure we have a jpg or png file
	err := fileIsValid(*fileName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Upload file to S3 bucket
	err = saveFile(*bucketName, *fileName)
	if err != nil {
		fmt.Println("Got an error saving " + *fileName + " to bucket " + *bucketName + ":")
		fmt.Println(err)
		return
	}

	fmt.Println("Saved '" + *fileName + "' in bucket " + *bucketName + " as 'uploads/" + *fileName + "'")

	// Save metadata in table
	err = saveMetadata(*bucketName, *fileName, *tableName)
	if err != nil {
		msg := "Got error saving metadata from '" + *fileName + "' in bucket '" + *bucketName + "':"
		fmt.Println(msg)
		fmt.Println(err)
		return
	}

	msg := "Saved metadata to table " + *tableName
	fmt.Println(msg)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got an error loading configuration")
		return
	}

	// Create the Rekognition client
	client := rekognition.NewFromConfig(cfg)

	// Filename has uploads/ prefix
	uploadedFilename := "uploads/" + *fileName

	input := &rekognition.DetectLabelsInput{
		Image: &rTypes.Image{
			S3Object: &rTypes.S3Object{
				Bucket: bucketName,
				Name:   &uploadedFilename,
			},
		},
	}

	// Get object data
	// GetLabels(c context.Context, api RekognitionDetectLabelsAPI, input *rekognition.DetectLabelsInput) (*rekognition.DetectLabelsOutput, error) {
	resp, err := GetLabels(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error calling GetLabels")
		return
	}

	// Save object data from Rekognition in table
	err = saveObjectData(resp.Labels, *tableName, *fileName)
	if err != nil {
		fmt.Println("Got an error saving object data:")
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Saved Rekognition data to table")

	thumbName, err := createThumbnail(*bucketName, *fileName)
	if err != nil {
		fmt.Println("Got an error creating thumbnail:")
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Saved thumbnail '" + thumbName + "' in bucket " + *bucketName)
}
