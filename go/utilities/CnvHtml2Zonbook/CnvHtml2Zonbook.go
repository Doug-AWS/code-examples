package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func verbosePrint(verbose bool, s string) {
	if verbose {
		fmt.Println(s)
	}
}

// If it's a heading, e.g., <h2 id="abc">HEADING</h2>,
// return 2, "abc", "HEADING"; otherwise, return 0, "", ""
func getHeading(text string, verbose bool) (int, string, string) {
	// Trap strings too short to be headings
	// Minimal heading: <h1>a</h1> is 10 chars
	l := len(text)

	if l < 10 {
		verbosePrint(verbose, "The line is too short to be a heading")
		return 0, "", ""
	}

	// Get first 2 chars
	first2 := text[0:2]

	if first2 != "<h" {
		verbosePrint(verbose, "The first two characters, "+first2+" are not <h")
		return 0, "", ""
	}

	level := string(text[2])

	verbosePrint(verbose, "Got level "+level)

	// For: <h2 id="aws-s3-construct-library">AWS S3 Construct Library</h2>
	// we return 2, "aws-s3-construct-library", "AWS S3 Construct Library"
	parts := strings.Split(text, "\"") // gives us "<h2 id=", "aws-s3-construct-library", ">AWS S3 Construct Library</h2>"
	// make sure we get three chunks
	if len(parts) < 3 {
		verbosePrint(verbose, "Could not get all of the parts of the string:")
		verbosePrint(verbose, text)
		return 0, "", ""
	}

	id := parts[1]
	title := parts[2]

	verbosePrint(verbose, "Got id: "+id)

	// title looks like:
	//    >AWS S3 Construct Library</h2>
	// so lop off the first char and the last 5 chars
	// Get first 2 chars
	title = title[1 : len(title)-5]
	verbosePrint(verbose, "Got title: "+title)

	i, err := strconv.Atoi(level)
	if err != nil {
		// It's not an int
		return 0, "", ""
	}

	return i, id, title
}

func writeStart(file os.File) {
	file.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	file.WriteString("<!DOCTYPE section PUBLIC \"-//OASIS//DTD DocBook XML V4.5//EN\" \"file://zonbook/docbookx.dtd\"\n")
	file.WriteString(" [\n")
	file.WriteString("  <!ENTITY % xinclude SYSTEM \"file://AWSShared/common/xinclude.mod\">\n")
	file.WriteString("  %xinclude;\n")
	file.WriteString("  <!ENTITY % phrases-shared SYSTEM \"file://AWSShared/common/phrases-shared.ent\">\n")
	file.WriteString("  %phrases-shared;\n")
	file.WriteString("  <!ENTITY % cdk_entities SYSTEM \"../shared/cdk.ent\">\n")
	file.WriteString("  %cdk_entities;\n")
	file.WriteString(" ]>\n")
	file.WriteString("<section role=\"topic\" id=\"about_examples\">\n")
	file.WriteString("  <info>\n")
	file.WriteString("    <title id=\"about_examples.title\">&CDK; Examples</title>\n")
	file.WriteString("<abstract>\n")
	file.WriteString("  <para>Examples for the &cdk;.</para>\n")
	file.WriteString("</abstract>\n")
	file.WriteString("</info>\n")
	file.WriteString("\n")
	file.WriteString("<para>The ")
	file.WriteString("<ulink url=\"https://github.com/aws-samples/aws-cdk-examples\">CDK Examples</ulink> ")
	file.WriteString("repo on GitHub includes the following examples.</para>\n")
	file.WriteString("\n")
}

func startNewSection(file os.File, id string, title string) {
	file.WriteString("<section id=\"" + id + "\">\n")
	file.WriteString("  <info>\n")
	file.WriteString("    <title id=\"" + id + ".title\">" + title + "</title>\n")
	file.WriteString("  </info>\n")
}

func closeSection(file os.File) {
	file.WriteString("</section>\n")
}

func writeEnd(file os.File) {
	file.WriteString("</section>\n")
}

func handleSectionTransition(currentSectionNumber int, newSectionNumber int, outFile *os.File, id string, title string) {
	// We have a new sub-section
	// e.g. were are in an h1 and find an h2
	if newSectionNumber > currentSectionNumber {
		startNewSection(*outFile, id, title)
		return
	}

	// We have a section at the same level
	// e.g., we are in an h2 and fnd an h2
	if newSectionNumber == currentSectionNumber {
		closeSection(*outFile)
		startNewSection(*outFile, id, title)
		return
	}

	// We have a higher-level section
	// e.g., we are in an h3 and find an h1
	// So we need to close current - new sections
	for currentSectionNumber >= newSectionNumber {
		closeSection(*outFile)
		currentSectionNumber--
	}

	startNewSection(*outFile, id, title)
}

// Opens an HTML file
// Name is going to be something like:
//    README.html

// Returns true if we created filename.xml from filename.html
func parseMd(name string, verbose bool) bool {
	parts := strings.Split(name, ".") // gives us filename and html
	xmlName := parts[0] + ".xml"      // filename.xml

	// Get input from filename.html
	inFile, err := os.Open(name)
	if err != nil {
		fmt.Println("Got error opening " + name)
		return false
	}

	outFileName := xmlName

	// Put output in filename.xml
	outFile, err := os.Create(outFileName)
	if err != nil {
		fmt.Println("Got error creating " + outFileName)
		return false
	}

	defer inFile.Close()
	defer outFile.Close()

	// Add initial lines to XML file
	writeStart(*outFile)

	lineNumber := 0
	//	inSection := false
	sectionNumber := 0
	inPara := false

	scanner := bufio.NewScanner(inFile)

	for scanner.Scan() {
		text := scanner.Text()
		lineNumber++

		if verbose {
			msg := fmt.Sprintf("%s %d %s", "Parsing line", lineNumber, ":")
			fmt.Println(msg)
			fmt.Println(text)
		}

		// If line is empty, skip it
		ll := len(text)

		if ll == 0 {
			continue
		}

		// Get the first three chars of the line
		firstThreeChars := text[:3]

		switch firstThreeChars {
		case "<h1", "<h2", "<h3":
			//			inSection = true

			hLevel, id, title := getHeading(text, verbose)

			// hLevel should be a 1
			if hLevel != 1 {
				fmt.Println("Error parsing heading!!!")
				fmt.Println("Should be an h1-3: " + text)
				os.Exit(1)
			}

			handleSectionTransition(sectionNumber, hLevel, outFile, id, title)
			sectionNumber = hLevel

			continue

		case "<p>":
			// Strip off initial '<p>',
			// create para tag,
			// and barf out everything until we find a '</p>':
			s := strings.TrimLeft(text, "<p>")
			outFile.WriteString(s + "\n")

			// If string ends with '</p>'
			// close para tag and move on.
			lastFourChars := text[4:]
			if lastFourChars == "</p>" {
				outFile.WriteString(text[:len(text)-4])
			} else {
				inPara = true
			}

			continue

		case "<ta", "<tb", "<td", "<th", "<tr", "</t":
			// Pass it through
			outFile.WriteString(text + "\n")
			continue

		default:
			if inPara {
				lastFourChars := text[4:]
				if lastFourChars == "</p>" {
					outFile.WriteString(text[:len(text)-4])
					inPara = false
				}
			}
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Got error scanning " + xmlName)
		return false
	}

	// We always have to close the last sub-section and the section
	closeSection(*outFile)
	writeEnd(*outFile)

	outFile.Close()

	return true
}

func main() {
	// Process command line args
	filePtr := flag.String("f", "", "Fully-qualified path to HTML file")
	verbosePtr := flag.Bool("v", false, "Whether to show verbose output")

	flag.Parse()

	// Validate args
	filePath := *filePtr
	verbose := *verbosePtr

	// Make sure we have a file
	if filePath == "" {
		fmt.Println("You must supply a fully-qualified path to HTML file (-f file-path")
		os.Exit(1)
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

	// Make sure we are only parsing HTML files
	// First get filename from path
	fileParts := strings.Split(filePath, "\\") // From D:\src\aws-cdk-examples\README.html
	// we get: {D:, src, aws-cdk-examples, README.html}
	length := len(fileParts)

	fileName := fileParts[length-1] // Now we have README.html, so is it an html file?

	parts := strings.Split(fileName, ".") // gives us README and html
	if len(parts) < 2 || parts[1] != "html" {
		fmt.Println(fileName + "is not an HTML file")
		return
	}

	verbosePrint(verbose, filePath)

	parsed := parseMd(filePath, verbose)

	if parsed {
		verbosePrint(verbose, "Successfully parsed"+filePath)
	}
}
