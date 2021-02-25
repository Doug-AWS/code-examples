package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MyRequestParameters is the event we receive
type MyRequestParameters struct {
	BucketName string `json:"bucketName"`
	Host       string `json:"Host"`
	Key        string `json:"key"`
	XID        string `json:"x-id"`
}

/* MyEvent defines the event we get
type MyEvent struct {
	Bucket string `json:"Bucket"`
	Key    string `json:"Key"`
}
*/

func calculateRatioFit(srcWidth, srcHeight int, maxWidth, maxHeight float64) (int, int) {
	ratio := math.Min(maxWidth/float64(srcWidth), maxHeight/float64(srcHeight))
	return int(math.Ceil(float64(srcWidth) * ratio)), int(math.Ceil(float64(srcHeight) * ratio))
}

func getObject(bucket, key string) (io.ReadCloser, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		msg := "Got configuration error loading context: " + err.Error()
		return nil, errors.New(msg)
	}

	client := s3.NewFromConfig(cfg)

	input := &s3.GetObjectInput{
		Bucket: &bucket, Key: &key,
	}

	resp, err := client.GetObject(context.TODO(), input)
	if err != nil {
		msg := "Got error calling GetObject: " + err.Error()
		return nil, errors.New(msg)
	}

	return resp.Body, nil
}

func putObject(bucket, key string, body io.Reader) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)

	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   body,
	}

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

func makeThumbnail(bucket, key string) (string, error) {
	var maxWidth float64 = 80
	var maxHeight float64 = 80

	// Get key from bucket
	body, err := getObject(bucket, key)
	if err != nil {
		return "", err
	}

	img, _, err := image.Decode(body)
	if err != nil {
		return "", err
	}

	b := img.Bounds()
	width := b.Max.X
	height := b.Max.Y

	fmt.Println("Original width:  " + strconv.Itoa(width))
	fmt.Println("Original height: " + strconv.Itoa(height))

	// Keep width/height ratio
	w, h := calculateRatioFit(width, height, maxWidth, maxHeight)

	fmt.Println("Thumbnail width:  " + strconv.Itoa(w))
	fmt.Println("Thumbnail height: " + strconv.Itoa(h))

	// Call the resize library for image scaling
	m := resize.Resize(uint(w), uint(h), img, resize.Lanczos3)

	// Create new object name from existing key name
	parts := strings.Split(key, ".")

	// If it has a "uploads/" prefix, delete the prefix
	nameParts := strings.Split(parts[0], "/")
	name := parts[0] + "thumb." + parts[1]

	if nameParts[0] == "uploads" {
		name = nameParts[1] + "thumb." + parts[1]
	}

	// Body of S3 object
	var buf bytes.Buffer

	// save the file in JPG or PNG format
	switch parts[1] {
	case "jpg":
		err := jpeg.Encode(&buf, m, nil)
		if err != nil {
			return "", err
		}

		break
	case "png":
		err = png.Encode(&buf, m)
		if err != nil {
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
	err = putObject(bucket, "thumbs/"+name, io.Reader(r))
	return name, err
}

// func handler(ctx context.Context, event MyEvent) (string, error) {
// func handler(ctx context.Context, myEvent MyRequestParameters) (string, error) {
func handler(ctx context.Context, myEvent MyRequestParameters) error {
	// Get bucket and key names from environment
	bucketName := os.Getenv("bucketName")
	keyName := os.Getenv("keyName")

	if bucketName == "" || keyName == "" {
		msg := "Did not get bucket and key"
		return errors.New(msg)
	}

	fmt.Print("Got bucket name '" + bucketName + "' from environment variable")
	fmt.Print("Got key name    '" + bucketName + "' from environment variable")

	//	fmt.Println("Got event in create thumbmail handler:")
	//	fmt.Println(myEvent)

	//	savedObject, err := makeThumbnail(myEvent.BucketName, myEvent.Key)
	savedObject, err := makeThumbnail(bucketName, keyName)
	if err != nil {
		msg := "Got error creating thumbnail from key '" + myEvent.Key + "' in bucket '" + myEvent.BucketName + "':"
		fmt.Println(msg)
		fmt.Println(err)

		return /*"",*/ err
	}

	msg := "Created thumbnail '" + savedObject + "' in bucket '" + /*myEvent.BucketName*/ bucketName + "'"
	fmt.Println(msg)

	//return "{ \"bucket\": " + event.Bucket + ", \"key\": " + savedObject + " }", nil

	/* fmt.Println("Returning: ")
	   fmt.Println(myEvent)

	output, err := json.Marshal(&myEvent)
	if err != nil {
		return "", err
	}

	return string(output), nil
	*/
	return nil
}

func main() {
	lambda.Start(handler)
}
