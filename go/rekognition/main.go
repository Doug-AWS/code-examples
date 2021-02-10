package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

// RekognitionDetectLabelsAPI defines the interface for the DetectLabels function
type RekognitionDetectLabelsAPI interface {
	DetectLabels(ctx context.Context,
		params *rekognition.DetectLabelsInput,
		optFns ...func(*rekognition.Options)) (*rekognition.DetectLabelsOutput, error)
}

// GetLabelInfo retrieves information about the labels
func GetLabelInfo(c context.Context, api RekognitionDetectLabelsAPI, input *rekognition.DetectLabelsInput) (*rekognition.DetectLabelsOutput, error) {
	resp, err := api.DetectLabels(c, input)

	return resp, err
}

func main() {
	bucket := flag.String("b", "", "The name of the bucket containing the photo")
	key := flag.String("k", "", "The name of the photo")
	//region := flag.String("r", "us-west-2", "The region")
	flag.Parse()

	if *bucket == "" || *key == "" {
		fmt.Println("You must specify a bucket and key name (-b BUCKET -k KEY)")
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
	resp, err := GetLabelInfo(context.Background(), client, input)
	if err != nil {
		panic("failed to get labels, " + err.Error())
	}

	fmt.Println("Info about " + *key + ":")

	for _, label := range resp.Labels {
		fmt.Println("Label: " + *label.Name)
		s := fmt.Sprintf("%f", *label.Confidence)
		if err == nil {
			fmt.Println("  Confidence: " + s)
		}

		fmt.Println()
	}
}
