package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

func isNameValid(key string) error {
	var err error
	// Ignore anything that doesn't have upload prefix or end with jpg or png
	// Make sure key ends in JPG or PNG
	parts := strings.Split(key, ".")

	if len(parts) < 2 {
		msg := "Could not split '" + key + "' into name/extension"
		err = errors.New(msg)
	} else {
		if parts[1] != "jpg" && parts[1] != "png" {
			msg := "'" + key + "' is not jpg or png"
			err = errors.New(msg)
		} else {
			// Trap anything without upload/ prefix
			pieces := strings.Split(parts[0], "/")

			if pieces[0] != "uploads" {
				msg := key + " does not have uploads/ prefix, it has " + pieces[0] + " prefix"
				err = errors.New(msg)
			}
		}
	}

	return err
}

func main() {
	key := flag.String("k", "", "The name of the photo") //
	flag.Parse()

	if *key == "" {
		fmt.Println("You must specify a key (JPG or PNG with 'uploads' prefix) -k KEY)")
		return
	}

	fmt.Println("Testing whether " + *key + " is a valid photo file")

	err := isNameValid(*key)
	if err != nil {
		fmt.Println(*key + " is not valid because:")
		fmt.Println(err.Error())
	}
}
