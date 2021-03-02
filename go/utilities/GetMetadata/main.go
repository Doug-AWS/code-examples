package main

import (
	"fmt"
	"os"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

// Entry contains the key/value pairs for the
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

func main() {
	fileName := "DonaldTrumpTheLemmingLeader.jpg"
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Unable to open file " + fileName)
		return
	}

	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		fmt.Println("Got error decoding exif data: " + err.Error())
		return
	}

	entries = make([]Entry, 1)

	var p Printer
	err = x.Walk(p)
	if err != nil {
		fmt.Println("Got an error walking the entries: " + err.Error())
		return
	}

	for _, e := range entries {
		if e.entryName != "" {
			fmt.Println(e.entryName + " == " + e.entryTag)
		}
	}
}
