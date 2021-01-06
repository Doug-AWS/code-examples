package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func verbosePrint(verbose bool, s string) {
	if verbose {
		fmt.Println(s)
	}
}

// If text is:
//   <info><title>CopyObject.go</title></info>
//   lang == "go"
//   sectionID == "copyobject"
// Return:
//   <info><title="copyobject.title">CopyObject</title></info>
// If lang == "", return (as sectionID is likely copyobject.go):
//   <info><title="copyobject.go.title">CopyObject.go</title></info>
func patchTitle(verbose bool, text string, lang string, sectionID string) string {
	// Split by '<'
	parts := strings.Split(text, "<")

	// That gives us:
	//   info>
	//   title>CopyObject.go
	//   /title>
	//   /info>
	if len(parts) != 4 {
		verbosePrint(verbose, text+" did not have four parts")
		return ""
	}

	// Get title tag and text
	bits := strings.Split(parts[1], ">")
	// This should give us:
	//   title
	//   CopyObject.go
	if len(bits) != 2 {
		verbosePrint(verbose, parts[1]+" did not have two parts")
		return ""
	}

	if bits[0] != "title" {
		return ""
	}

	if lang != "" {
		title := strings.Replace(bits[1], "."+lang, "", 1) // CopyObject
		return "<" + parts[0] + "<" + bits[0] + " id=\"" + sectionID + "\">" + title + "<" + parts[2] + "<" + parts[3]
	}

	return "<" + parts[0] + "<" + bits[0] + " id=\"" + sectionID + "\">" + bits[1] + "<" + parts[2] + "<" + parts[3]
}

// If text is a section tag, such as:
//   <section id="copyobject.go">
// and lang is "go"
// return "copyobject"
func getSection(verbose bool, text string, lang string) string {
	// Trap strings too short to a section tag
	s := strings.TrimLeft(text, " \t")

	// Is it long enough to be a sectiont tag?
	minLength := len("<section id=\" \">")

	if len(s) < minLength {
		verbosePrint(verbose, "Not a section tag")
		return ""
	}

	// Get first 8 chars
	first8 := s[0:8]

	if first8 != "<section" {
		verbosePrint(verbose, "Not a section tag")
		return ""
	}

	// Split s by spaces, which should give us:
	//   <section
	// and
	//   id="copyobject.go">
	parts := strings.Split(s, " ")

	// parts should be:
	//   id=
	// and
	//   copyobject.go
	// and
	//   >
	// So return "" if not three parts
	if len(parts) != 3 {
		fmt.Println("section tag did not have an ID attribute")
		return ""
	}

	// Make sure the first part is "id="
	if parts[0] != "id=" {
		fmt.Println("section tag did not have an ID attribute")
		return ""
	}

	if lang == "" {
		return parts[1]
	}

	newS := strings.Replace(parts[1], "."+lang, "", 1)

	return newS
}

// Patches the section ID and creates the associated title ID in an XML (Zonbook) file
// Name is going to be something like:
//    CopyObject.xml

// Returns true if we created patch file name.lmx
func patchFile(name string, lang string, verbose bool) bool {
	parts := strings.Split(name, ".") // gives us filename and xml
	// make sure it has two parts
	if len(parts) != 2 {
		fmt.Println("The filename " + name + " did not match NAME.EXTENSION")
		return false
	}

	saveName := parts[0] + ".lmx" // filename.lmx

	// Get contents of file
	inFile, err := os.Open(name)
	if err != nil {
		fmt.Println("Got error opening " + name)
		return false
	}

	// Put output in filename.lmx
	outFile, err := os.Create(saveName)
	if err != nil {
		fmt.Println("Got error creating " + saveName)
		return false
	}

	defer inFile.Close()
	defer outFile.Close()

	// True if we find a <section> tag like:
	//   <section id="copyobject.go">
	inSection := false
	lineNumber := 0

	scanner := bufio.NewScanner(inFile)

	for scanner.Scan() {
		text := scanner.Text()
		lineNumber++

		if verbose {
			msg := fmt.Sprintf("%s %d %s", "Parsing line", lineNumber, ":")
			fmt.Println(msg)
			fmt.Println(text)
		}

		// If line is empty, print it out as-is
		ll := len(text)

		if ll == 0 {
			outFile.WriteString("\n")
			continue
		}

		sectionID := getSection(verbose, text, lang)

		if sectionID != "" {
			inSection = true
		}

		// If we don't have a section tag, we can pass eveything through for now
		if !inSection {
			outFile.WriteString(text + "\n")
			continue
		}

		patchedTitle := patchTitle(verbose, text, lang, sectionID)

		if patchedTitle != "" {
			outFile.WriteString(patchedTitle + "\n")
		} else {
			outFile.WriteString(text + "\n")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Got error scanning " + saveName)
		return false
	}

	outFile.Close()

	return true
}

// Updates a Zonbook XML file so the title has an ID (section ID + ".title")
// If a -l flag is supplied, lops off the .lang from the section ID and title value
func main() {
	// Process command line args
	filePtr := flag.String("f", "", "Fully-qualified path to XML file")
	// If no language, we don't touch the section ID or title value
	langPtr := flag.String("l", "", "The programming language (so we know what to lop off any section ID or title value")
	verbosePtr := flag.Bool("v", false, "Whether to show verbose output")

	flag.Parse()

	// Validate args
	filePath := *filePtr
	lang := *langPtr
	verbose := *verbosePtr

	// Make sure we have a file
	if filePath == "" {
		fmt.Println("You must supply a fully-qualified path to XML file (-f file-path")
		return
	}

	verbosePrint(verbose, "Determining whether "+filePath+" exists.")

	// Confirm file exists
	_, err := os.Stat(filePath)
	if err != nil {
		// if os.IsNotExist(err) {
		fmt.Println(filePath + " does not exist.")
		os.Exit(1)
		// }
	} else {
		verbosePrint(verbose, filePath+" exists")
	}

	// Make sure we are only parsing an XML file
	// First get filename from path
	fileParts := strings.Split(filePath, "\\") // From D:\src\aws-cdk-examples\file.xml
	// we get: {D:, src, aws-cdk-examples, file.xml}
	length := len(fileParts)

	fileName := fileParts[length-1] // Now we have file.xml, so is it an xml file?

	parts := strings.Split(fileName, ".") // gives us file and xml
	if len(parts) < 2 || parts[1] != "xml" {
		fmt.Println(fileName + "is not an XML file")
		return
	}

	verbosePrint(verbose, filePath)

	patched := patchFile(filePath, lang, verbose)

	if patched {
		verbosePrint(verbose, "Successfully patched"+filePath)
	}
}
