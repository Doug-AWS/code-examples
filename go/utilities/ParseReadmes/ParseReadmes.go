package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Element defines the different tags
type Element int

const (
	// H1 tag
	H1 Element = iota
	// H2 tag
	H2
	// Para tag
	Para
	// Bullet tag
	Bullet
	// Step tag
	Step
	// Other tags, ignored
	Other
)

// Usage displays the command-line options
func Usage() {
	fmt.Println("USAGE:")
	fmt.Println("")
	fmt.Println("go run ParseReadmes.go -d DIRECTORY [-c CONFIG] [-v VERBOSE]")
	fmt.Println(" where:")
	fmt.Println("  DIRECTORY is the directory in which the readme.html files are found")
	fmt.Println("  CONFIG is the filename of your config file.")
	fmt.Println("  if omitted, defaults to config.json")
	fmt.Println("  -v gives you verbose/debugging info")
	fmt.Println("")
}

// Config represents the resources we need to configure before we can run the example
type Config struct {
	Abstract  string `json:"Abstract"`
	ChapAbbv  string `json:"ChapAbbv"`
	ChapID    string `json:"ChapID"`
	ChapTitle string `json:"ChapTitle"`
	Entity    string `json:"Entity"`
	Format    string `json:"Format"`
	InfoText  string `json:"InfoText"`
	SDK       string `json:"SDK"`
}

// GetConfigInfo creates a Config object from a JSON file
func GetConfigInfo(verbose *bool, configFileName *string) (Config, error) {
	var globalConfig Config
	// Open config file
	content, err := ioutil.ReadFile(*configFileName)
	if err != nil {
		return globalConfig, err
	}

	text := string(content)

	err = json.Unmarshal([]byte(text), &globalConfig)
	if err != nil {
		return globalConfig, err
	}

	// Only entry required for RST conversion
	if globalConfig.Format == "" || globalConfig.SDK == "" {
		msg := "You must supply a Format entry (rst or xml) and SDK entry in " + *configFileName
		return globalConfig, errors.New(msg)
	}

	if globalConfig.Format == "rst" {
		verbosePrint(verbose, "Format: "+globalConfig.Format)
		verbosePrint(verbose, "SDK:    "+globalConfig.SDK)
	}

	// Make sure the remaining entries are there for xml
	if globalConfig.Abstract == "" ||
		globalConfig.ChapAbbv == "" ||
		globalConfig.ChapID == "" ||
		globalConfig.ChapTitle == "" ||
		globalConfig.Entity == "" ||
		globalConfig.InfoText == "" {
		msg := "You must supply an Abstract, ChapAbbv, ChapID, ChapTitle, Entity, and InfoText entry in " + *configFileName + " for XML format"
		return globalConfig, errors.New(msg)
	}

	if globalConfig.Format == "xml" {
		verbosePrint(verbose, "Abstract:  "+globalConfig.Abstract)
		verbosePrint(verbose, "ChapAbbv:  "+globalConfig.ChapAbbv)
		verbosePrint(verbose, "ChapID:    "+globalConfig.ChapID)
		verbosePrint(verbose, "ChapTitle: "+globalConfig.ChapTitle)
		verbosePrint(verbose, "Entity:    "+globalConfig.Entity)
		verbosePrint(verbose, "InfoText:  "+globalConfig.InfoText)
		verbosePrint(verbose, "SDK:       "+globalConfig.SDK)
	}

	return globalConfig, nil
}

func verbosePrint(verbose *bool, s string) {
	if *verbose {
		fmt.Println(s)
	}
}

// Character used for MD headings, e.g., # Heading1, ## Heading 2, etc.
var mdH = "#"

// Character used for RST H1 headings
var rstH1 = "="

// Character used for RST H2 headings
var rstH2 = "-"

// Return the string of mdH, rstH1, or rstH2 chars the length of the heading
func getRstUnderscores(level int, heading *string) string {
	ch := ""

	switch level {
	case 0:
		ch = mdH
	case 1:
		ch = rstH1
		break
	case 2:
		ch = rstH2
		break
	default:
		return ""
	}

	i := 0

	for i < len(*heading) {
		ch += ch
		i++
	}

	return ch
}

func writePreRSTAb(file os.File, configFile *Config) {
	us := getRstUnderscores(0, &configFile.ChapTitle)

	file.WriteString(".. Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.\n")
	file.WriteString("   SPDX-License-Identifier: CC-BY-SA-4.0\n")
	file.WriteString("\n")
	file.WriteString(".. _advanced-examples\n")
	file.WriteString(us + "\n")
	file.WriteString(configFile.ChapTitle + "\n")
	file.WriteString(us + "\n")
	file.WriteString("\n")
	file.WriteString(". meta::\n")
	file.WriteString("   :description: About the advanced examples that use the " + configFile.SDK + "\n")
	file.WriteString("   :keywords: " + configFile.SDK + " code examples\n")
	file.WriteString("")
	file.WriteString("This SDK provides the following advanced code examples.\n")
	file.WriteString("\n")
	file.WriteString(".. toctree::\n")
	file.WriteString("    :titlesonly:\n")
	file.WriteString("   :maxdepth: 1\n")
	file.WriteString("\n")
}

func writePreXMLAb(file os.File, configFile *Config) {
	file.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	file.WriteString("<!DOCTYPE chapter PUBLIC \"-//OASIS//DTD DocBook XML V4.5//EN\" \"file://zonbook/docbookx.dtd\"\n")
	file.WriteString(" [\n")
	file.WriteString("  <!ENTITY % xinclude SYSTEM \"file://AWSShared/common/xinclude.mod\">\n")
	file.WriteString("  %xinclude;\n")
	file.WriteString("  <!ENTITY % phrases-shared SYSTEM \"file://AWSShared/common/phrases-shared.ent\">\n")
	file.WriteString("  %phrases-shared;\n")
	file.WriteString("  <!ENTITY % cdk_entities SYSTEM \"../shared/cdk.ent\">\n")
	file.WriteString("  %cdk_entities;\n")
	file.WriteString(" ]>\n")
	file.WriteString("<chapter role=\"topic\" id=\"" + configFile.ChapID + "\">\n")
	file.WriteString("<info>\n")
	file.WriteString("<title id=\"about-libraries.title\">" + configFile.ChapTitle + "\"</title>\n")
	file.WriteString("<titleabbrev>" + configFile.ChapAbbv + "</titleabbrev>\n")
	file.WriteString("<abstract>\n")
	file.WriteString("<para>" + configFile.Abstract + "</para>\n")
	file.WriteString("</abstract>\n")
	file.WriteString("</info>\n")
	file.WriteString("\n")
	file.WriteString("<para>" + configFile.InfoText + "</para>\n")
	file.WriteString("\n")
}

// Concatenate parts into one, space-separated string, starting at parts[index]
func pasteSplit(index int, parts []string) string {
	length := len(parts)
	result := ""

	// say parts[0] == "hello" and parts[1] == "world" and index == 2 or more
	// return an empty string
	if length <= index {
		return ""
	}

	for i := index; i < length; i++ {
		result += " " + parts[index]
	}

	return result
}

func parseLine(verbose *bool, line *string) (Element, string) {
	element := Other
	text := *line
	// We look for and return:
	// # text ==> H1, text
	// ## text ==> H2, text
	// - text ==> Bullet, text
	// n. text ==> Step, text
	// text ==> Para, text
	// Get the chars BEFORE the first space
	parts := strings.Split(text, " ")

	// If we don't have a space
	if len(parts) < 2 {
		return element, text
	}

	switch parts[0] {
	case "#":
		element = H1
		text = pasteSplit(1, parts)
		break
	case "##":
		element = H2
		text = pasteSplit(1, parts)
		break
	case "-":
		element = Bullet
		text = pasteSplit(1, parts)
		break
	default:
		switch parts[0][1:2] {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			element = Step
			text = pasteSplit(1, parts)
			break
		default:
		}
		break
	}

	return element, text
}

// Parse a line of MD and barf out the RST tag and text to the file
func parseLineRST(verbose *bool, file os.File, line *string) error {
	verbosePrint(verbose, "Parsing MD line: "+*line)

	// Is this the right thing to do for a blank line?
	if *line == "" {
		file.WriteString("\n")
		return nil
	}

	element, text := parseLine(verbose, line)

	switch element {
	case H1:
		us := getRstUnderscores(1, &text)
		file.WriteString(us + "\n")
		file.WriteString(text + "\n")
		file.WriteString(us + "\n")

		break
	case H2:
		us := getRstUnderscores(2, &text)
		file.WriteString(us + "\n")
		file.WriteString(text + "\n")
		file.WriteString(us + "\n")

		break
	case Bullet, Para, Step:
		file.WriteString(*line + "\n")
		break
	default:
	}

	return nil
}

// Parse a line of MD and barf out the apprpriate XML tag and text to the file
func parseLineXML(verbose *bool, file os.File, line *string) error {
	verbosePrint(verbose, "Parsing MD line: "+*line)

	// Is this the right thing to do for a blank line?
	if *line == "" {
		file.WriteString("\n")
		return nil
	}

	return nil
}

// Parse MD file and barf out chunks into RST or XML file
func parseReadme(verbose *bool, file os.File, format, name, dir *string) error {
	// Open file
	inFile, err := os.Open(*dir + "/" + *name) // parse serverless.md
	if err != nil {
		return err
	}

	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)

	lineNumber := 0

	for scanner.Scan() {
		text := scanner.Text()
		lineNumber++

		if *verbose {
			msg := fmt.Sprintf("%s %d %s", "Parsing line", lineNumber, ":")
			fmt.Println(msg)
			fmt.Println(text)
		}

		switch *format {
		case "rst":
			err = parseLineRST(verbose, file, &text)
			if err != nil {
				return err
			}

			break
		case "xml":
			err = parseLineXML(verbose, file, &text)
			if err != nil {
				return err
			}

			break
		default:
			msg := "The format is neither rst nor xml, which should never happen!"
			return errors.New(msg)
		}

	}

	return nil
}

func run(verbose *bool, configFile Config, dir *string) error {
	ext := ""

	// Prep about-examples topic
	switch configFile.Format {
	case "rst":
		ext = "rst"
		break
	case "xml":
		ext = "xml"
		break
	default:
		msg := "The format is neither rst nor xml, which should never happen!"
		return errors.New(msg)
	}

	aboutFileName := *dir + "/about-examples" + ext

	// Create empties the file if it exists
	abFile, err := os.Create(aboutFileName)
	if err != nil {
		return err
	}

	verbosePrint(verbose, "Created "+aboutFileName)

	defer abFile.Close()

	switch configFile.Format {
	case "rst":
		writePreRSTAb(*abFile, &configFile)
		break
	case "xml":
		writePreXMLAb(*abFile, &configFile)
		break
	default:
		msg := "The format is neither rst nor xml, which should never happen!"
		return errors.New(msg)
	}

	files, err := ioutil.ReadDir(*dir)
	if err != nil {
		return err
	}

	verbosePrint(verbose, "Reading contents of "+*dir)

	// f.Name is going to be something like:
	//    serverless.html
	for _, f := range files {
		name := f.Name()
		_, err = os.Stat(*dir + "/" + name)
		if err != nil {
			return err
		}

		// Make sure we are only parsing MD files
		parts := strings.Split(name, ".") // gives us serverless and md
		extension := parts[1]

		parsed := false

		if extension == "md" {
			verbosePrint(verbose, "Parsing MD file: "+name)

			err := parseReadme(verbose, *abFile, &configFile.Format, &name, dir)
			if err != nil {
				return err
			}

			if parsed {
				verbosePrint(verbose, "Adding package to list of libraries")
				abFile.WriteString("<xi:include href=\"" + parts[0] + ".xml\"/>\n")
			}
		}
	}

	// Close about-libraries.xml
	abFile.WriteString("</chapter>")

	return nil
}

func main() {
	// Process command line args
	dir := flag.String("d", "", "Directory containing packages in HTML files")
	config := flag.String("c", "config.json", "The name of your configuration file, default is config.json")
	verbose := flag.Bool("v", false, "Whether to show verbose output")
	help := flag.Bool("h", false, "Display usage and quit")

	flag.Parse()

	if *help {
		Usage()
		return
	}

	// Make sure we have a directory and config file
	if *dir == "" || *config == "" {
		Usage()
		return
	}

	verbosePrint(verbose, "Determining whether "+*dir+" exists.")

	// Confirm dir exists
	_, err := os.Stat(*dir)
	if err != nil {
		fmt.Println(*dir + " does not exist.")
		return
	}

	verbosePrint(verbose, *dir+" exists")

	configFile, err := GetConfigInfo(verbose, config)
	if err != nil {
		fmt.Println("Got an error reading config file " + *config + ":")
		fmt.Println(err)
		return
	}

	err = run(verbose, configFile, dir)
	if err != nil {
		fmt.Println("Got an error:")
		fmt.Println(err)
	} else {
		fmt.Println("Done")
	}
}
