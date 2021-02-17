package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

// Calculate the size of the image after scaling
func calculateRatioFit(srcWidth, srcHeight int, maxWidth, maxHeight float64) (int, int) {
	ratio := math.Min(maxWidth/float64(srcWidth), maxHeight/float64(srcHeight))
	return int(math.Ceil(float64(srcWidth) * ratio)), int(math.Ceil(float64(srcHeight) * ratio))
}

// generate a thumbnail
func makeThumbnail(debug bool, imagePath, savePath string, maxWidth, maxHeight float64) error {
	parts := strings.Split(imagePath, ".")

	file, _ := os.Open(imagePath)
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	b := img.Bounds()
	width := b.Max.X
	height := b.Max.Y

	debugPrint(debug, "Original width:  "+strconv.Itoa(width))
	debugPrint(debug, "Original height: "+strconv.Itoa(height))

	// Keep width/height ratio
	w, h := calculateRatioFit(width, height, maxWidth, maxHeight)

	debugPrint(debug, "Thumbnail width:  "+strconv.Itoa(w))
	debugPrint(debug, "Thumbnail height: "+strconv.Itoa(h))

	// Call the resize library for image scaling
	m := resize.Resize(uint(w), uint(h), img, resize.Lanczos3)

	// files that need to be saved
	imgfile, _ := os.Create(savePath)
	defer imgfile.Close()

	// save the file in JPG or PNG format
	switch parts[1] {
	case "jpg":
		err := jpeg.Encode(imgfile, m, nil)
		if err != nil {
			return err
		}

		break
	case "png":
		err = png.Encode(imgfile, m)
		if err != nil {
			return err
		}

		break

	default:
		msg := "Unsupported format: " + parts[1]
		return errors.New(msg)
	}

	return nil
}

var configFileName = "config.json"

// Config keeps track of our requested thumbnail size
type Config struct {
	MaxWidthString  string `json:"MaxWidth"`
	MaxHeightString string `json:"MaxHeight"`
	MaxWidth        float64
	MaxHeigth       float64
}

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

	globalConfig.MaxWidth, err = strconv.ParseFloat(globalConfig.MaxWidthString, 64)
	if err != nil {
		globalConfig.MaxWidth = 320
	}

	globalConfig.MaxHeigth, err = strconv.ParseFloat(globalConfig.MaxHeightString, 64)
	if err != nil {
		globalConfig.MaxHeigth = 240
	}

	fmt.Println("Max width: " + globalConfig.MaxWidthString)
	fmt.Println("Max height: " + globalConfig.MaxHeightString)

	return nil
}

func main() {
	imageFile := flag.String("f", "", "The JPG or PNG image to resize")
	debug := flag.Bool("d", false, "Whether to barf out additional info")
	flag.Parse()

	if *imageFile == "" {
		fmt.Println("You must supply the name of an image (-f FILENAME)")
		return
	}

	err := populateConfiguration()
	if err != nil {
		fmt.Println("Got an error loading configuration from " + configFileName + ":")
		fmt.Println(err)
		return
	}

	parts := strings.Split(*imageFile, ".")

	saveFile := parts[0] + "thumb." + parts[1]

	err = makeThumbnail(*debug, *imageFile, saveFile, globalConfig.MaxWidth, globalConfig.MaxHeigth)
	if err != nil {
		fmt.Println("Could not create thumbnail of " + *imageFile)
	} else {
		fmt.Println("Created thumbnail in " + saveFile)
	}
}
