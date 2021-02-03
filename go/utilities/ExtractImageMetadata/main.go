package main

import (
	"fmt"
	"log"
	"os"

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

func main() {
	if len(os.Args) < 2 {
		log.Fatal("please give filename as argument")
	}
	fname := os.Args[1]

	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	x, err := exif.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	entries = make([]Entry, 1)

	var p Printer
	x.Walk(p)

	// Sort entries:
	quickSort(&entries, 0, len(entries)-1)

	// Barf out entries:
	for _, e := range entries {
		if e.entryName != "" {
			fmt.Printf("%40s: %s\n", e.entryName, e.entryTag)
		}
	}
}
