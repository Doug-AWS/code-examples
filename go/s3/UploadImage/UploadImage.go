package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config contains our configuration info
type Config struct {
	BucketName string `json:"BucketName"`
	MaxWait    int    `json:"MaxWait"`
}

var configFileName = "config.json"

var globalConfig Config

func populateConfiguration() error {
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	text := string(content)

	err = json.Unmarshal([]byte(text), &globalConfig)
	if err != nil {
		return err
	}

	return nil
}

func multiplyDuration(factor int64, d time.Duration) time.Duration {
	return time.Duration(factor) * d
}

func main() {
	filename := flag.String("f", "", "The file to upload")
	flag.Parse()

	err := populateConfiguration()
	if err != nil {
		return
	}

	if *filename == "" {
		fmt.Println("You must supply a file to upload (-f FILE)")
		return
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := s3.NewFromConfig(cfg)

	file, err := os.Open(*filename)

	if err != nil {
		fmt.Println("Unable to open file " + *filename)
		return
	}

	defer file.Close()

	// Add uploads/ prefix to trigger notification
	*filename = "uploads/" + *filename

	input := &s3.PutObjectInput{
		Bucket: &globalConfig.BucketName,
		Key:    filename,
		Body:   file,
	}

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		fmt.Println("Got error uploading file:")
		fmt.Println(err)
		return
	}

	// Create thumbnail name from original name
	// So myFile.jpg -> thumbs/myFilethumb.jpg
	parts := strings.Split(*filename, ".")
	thumbName := parts[0] + "thumb." + parts[1]
	thumbPath := "thumbs/" + thumbName

	// Wait for thumbnail to appear
	getInput := &s3.GetObjectInput{
		Bucket: &globalConfig.BucketName,
		Key:    &thumbPath,
	}

	wait := 1

	downLoader := manager.NewDownloader(s3.NewFromConfig(cfg))
	getBuf := manager.NewWriteAtBuffer([]byte{})

	var fileMode os.FileMode
	const OsRead = 04
	const OsWrite = 02
	const OsEx = 01
	const PERM = OsRead | OsWrite | OsEx
	fileMode = os.ModeDir | PERM

	foundIt := false
	totalTime := 0

	for wait < globalConfig.MaxWait {
		// Download thumbName
		fmt.Println("Waiting " + strconv.Itoa(wait) + " seconds to download thumbnail")
		ts := multiplyDuration(int64(wait), time.Second)
		time.Sleep(ts)
		totalTime += wait

		_, err = downLoader.Download(context.TODO(), getBuf, getInput)

		if err != nil {
			wait = wait * 2
		} else {
			// Save thumbnail
			err := ioutil.WriteFile(thumbName, getBuf.Bytes(), fileMode)
			if err != nil {
				// handle error
			}

			foundIt = true

			fmt.Println("Saved " + thumbName)
		}
	}

	if !foundIt {
		fmt.Println("Waited " + strconv.Itoa(totalTime) + " seconds total, but did not download " + thumbName)
	}
}
