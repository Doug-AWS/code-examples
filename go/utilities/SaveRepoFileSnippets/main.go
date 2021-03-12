package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func doesLineStartWithTag(debug bool, ext string, tag string, line string) (bool, error) {
	retVal := false
	// A snippet-start looks something like:
	// "  // snippet-start:[s3.go.create_bucket.imports] "

	/* Process line:
	   1. Strip off whitespace, comment char(s) (language dependent), whitespace
	   2. Return whether the result starts with
	        snippet-start
	*/

	if line == "" {
		return false, nil
	}

	part := strings.TrimSpace(line)

	switch ext {
	case "go":
		part = strings.TrimLeft(part, "//")
		part = strings.TrimSpace(part)
		// We should now have something like:
		//   "snippet-start:[s3.go.create_bucket.imports]",
		// so see if the first N characters (len(tag)) match tag
		// debugPrint(debug, "Line after stripping whitespace and comment: "+part)

		// If part is not at least as long as tag, return false
		if len(part) < len(tag) {
			return false, nil
		}

		firstN := part[:len(tag)]
		// debugPrint(debug, "First "+strconv.Itoa(len(tag))+" chars: "+firstN)
		retVal = firstN == tag

	default:
		msg := "Unrecognized file extension: " + ext
		return false, errors.New(msg)
	}

	return retVal, nil
}

func saveSnippet(debug bool, ext string, filename string, content []string) error {
	// We get a name like:
	//   s3.go.create_bucket.imports
	// So we save content as s3.go.create_bucket.imports.txt
	// And wrap it in a programlisting tag for s3.go.create_bucket.imports.xml

	// Since content lines might include an embedded snippet tag, igore those as we write out the file.

	numLines := 0

	for _, line := range content {
		start, err := doesLineStartWithTag(debug, ext, snippetStart, line)
		if err != nil {
			fmt.Println("Got an error determining whether line starts with snippet start tag:")
			fmt.Println(err)
			return err
		}
		end, err := doesLineStartWithTag(debug, ext, snippetEnd, line)
		if err != nil {
			fmt.Println("Got an error determining whether line starts with snippet end tag:")
			fmt.Println(err)
			return err
		}

		if !start && !end {
			numLines++
		}
	}

	debugPrint(debug, "Input had "+strconv.Itoa(len(content))+" lines")
	debugPrint(debug, "Saving text snippet of "+strconv.Itoa(numLines)+" lines in:    "+filename+".txt")
	debugPrint(debug, "Saving Zonbook snippet in: "+filename+".xml")
	debugPrint(debug, "")

	return nil
}

func getSnippetName(debug bool, line string) string {
	// Line must look something like:
	//   // snippet-start:[s3.go.create_bucket.imports]
	// Split line by '[' and then ']'
	snippetParts := strings.Split(line, "[")
	// If there aren't two parts, return an error
	if len(snippetParts) != 2 {
		return ""
	}

	// snippetParts[0] should look like:
	//   // snippet-start:
	// snippetParts[1] should look like:
	//  s3.go.create_bucket.imports]
	// so split snippetParts[1] by ']':
	endParts := strings.Split(snippetParts[1], "]")

	// Again, if there aren't two parts, return an error
	if len(endParts) != 2 {
		return ""
	}

	// endParts[0] should look like:
	//   s3.go.create_bucket.imports
	// so return it

	return endParts[0]

}

var snippetStart string = "snippet-start"
var snippetEnd string = "snippet-end"

func findSnippets(debug bool, ext string, content []string) error {
	var snippetStrings []string
	inSnippet := false
	snippetName := ""

	for i, line := range content {
		start, err := doesLineStartWithTag(debug, ext, snippetStart, line)
		if err != nil {
			fmt.Println("Got an error determining whether line starts with snippet start tag:")
			fmt.Println(err)
			return err
		}
		end, err := doesLineStartWithTag(debug, ext, snippetEnd, line)
		if err != nil {
			fmt.Println("Got an error determining whether line starts with snippet end tag:")
			fmt.Println(err)
			return err
		}

		if !inSnippet {
			if start {
				inSnippet = true
				snippetName = getSnippetName(debug, line)

				if snippetName == "" {
					msg := "Got blank snippet name"
					return errors.New(msg)
				}

				// debugPrint(debug, "Got snippet start name: "+snippetName)
				// debugPrint(debug, "")
			}
		} else {
			if start {
				// We've encountered a new snippet
				err := findSnippets(debug, ext, content[i:])
				if err != nil {
					fmt.Println("Got an error calling findSnippets with " + strconv.Itoa(len(content[:i])) + " strings ")
					return err
				}
			} else {
				if end {
					name := getSnippetName(debug, line)
					if name == "" {
						msg := "Got an empty snippet name"
						return errors.New(msg)
					}

					// debugPrint(debug, "Got snippet end name: "+name)
					// debugPrint(debug, "")

					if name == snippetName {
						// debugPrint(debug, "End name '"+name+"' matches start '"+snippetName+"'")
						// debugPrint(debug, "")

						saveSnippet(debug, ext, snippetName, snippetStrings)
						return nil
					} else {
						// debugPrint(debug, "End name '"+name+"' does NOT match start '"+snippetName+"'")
						// debugPrint(debug, "")

						// It's not a match, so add line to snippetStrings
						snippetStrings = append(snippetStrings, line)
					}

				} else {
					// It's not an end tag, so just add the line to snippetStrings
					snippetStrings = append(snippetStrings, line)
				}
			}
		}
	}

	return nil
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

func main() {
	path := flag.String("p", "", "The path to the file to parse for snippets")
	debug := flag.Bool("d", false, "Whether to print additional info")
	flag.Parse()

	if *path == "" {
		fmt.Println("You must specify a path to the file to search for snippets")
		fmt.Println("-p PATH")
		return
	}

	debugPrint(*debug, "Searching "+*path+" for snippet tags")
	debugPrint(*debug, "")

	req, err := http.NewRequest("GET", *path, nil)
	if err != nil {
		fmt.Println("Got an error creating HTTP request:")
		fmt.Println(err.Error())
		return
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Got and error making HTTP request")
		return
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Got an error reading response body")
		return
	}

	myStrings := strings.Split(string(bytes), "\n")

	/*
		err = dumpStrings(*debug, myStrings)
		if err != nil {
			fmt.Println("Got an error dumping strings:")
			fmt.Println(err.Error())
			return
		}
	*/

	// Get extension from path
	parts := strings.Split(*path, ".")
	ext := parts[len(parts)-1]

	debugPrint(*debug, "Found extension: "+ext)
	debugPrint(*debug, "")

	err = findSnippets(*debug, ext, myStrings)
	if err != nil {
		fmt.Println("Got an error searching for snippets:")
		fmt.Println(err.Error())
	}
}
