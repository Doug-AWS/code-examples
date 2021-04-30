package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// RepoTree represents the files in a repo
// See https://docs.github.com/en/free-pro-team@latest/rest/reference/git#trees
type RepoTree struct {
	Sha  string `json:"sha"` // Like a checksum for the request
	URL  string `json:"url"` // The API URL for the request
	Tree []struct {
		Path string `json:"path"` // The path to the directory/file after https://github.com/awsdocs/aws-doc-sdk-examples/blob/master/
		Mode string `json:"mode"` // The file permissions. Most folders are ; files 100644
		Type string `json:"type"` // "blob" for files; "tree" for directories
		Sha  string `json:"sha"`  // Like a checksum for the directory/file
		Size int    `json:"size"` // The size, in bytes, of files
		URL  string `json:"url"`  // The API URL for the file
	} `json:"tree"`
	Truncated bool `json:"truncated"` // Whether the request was truncated (more to come)
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

// If output directory does not exist, create it.
func createOutdir(debug bool, outDir string) error {
	src, err := os.Stat(outDir)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(outDir, 0755)
		if errDir != nil {
			panic(err)
		}
		return nil
	}

	if src.Mode().IsRegular() {

		return errors.New("Exists as a file!")
	}

	return nil
}

/* Eventually we'll create a complete tablist for all applicable code examples like:

    <tablist>
        <tablistentry>
            <tabname>LANGUAGE</tabname>
            <tabcontent>
			    <para>See the
                    <ulink url="LINK TO CODE EXAMPLE">SOURCE FILE NAME</ulink> code example in the GitHub repository.
                </para>
            </tabcontent>
		</tablistentry>
		...
	</tablist>
*/

// Create a Zonbook XML file containing link to a code example that can be XINCLUDEd in a tab.
// The format of the filename is SERVICE-OPERATION-LANGUAGE-link.xml
func createLinkFile(debug bool, service, operation, language, sourceFileName, sourceFileURL, outDir string) error {
	outFileName := service + "-" + operation + "-" + language + "-link.xml"

	debugPrint(debug, "Creating link file "+outFileName)

	file, err := os.Create(outDir + "/" + outFileName)
	if err != nil {
		return err
	}

	_, err = file.WriteString("<ulink url=\"" + sourceFileURL + "\">" + sourceFileName + "</ulink>")
	if err != nil {
		return err
	}

	file.Close()

	return nil
}

func mapLanguage(debug bool, language string) string {
	// Takes a language directory, like gov2, and returns the tab name for it, like Go
	lang := ""
	switch language {
	case "dotnetv3":
		lang = ".NET/C#"
	case "gov2":
		lang = "Go"
	}

	return lang
}

// Create a Zonbook XML file that can be include as a tab in a table.
// The format of the filename is SERVICE-OPERATION-LANGUAGE-tab.xml
/* <tablistentry>
       <tabname>LANGUAGE</tabname>
           <tabcontent>
               <para>See the
                   <ulink url="LINK TO CODE EXAMPLE">SOURCE FILE NAME</ulink> code example in the GitHub repository.
                </para>
            </tabcontent>
		</tablistentry>
*/
func createTabFile(debug bool, service, operation, language, sourceFileName, sourceFileURL, outDir string) error {
	linkFileName := service + "-" + operation + "-" + language + "-link.xml"
	outFileName := service + "-" + operation + "-" + language + "-tab.xml"

	debugPrint(debug, "Creating link file "+outFileName)

	file, err := os.Create(outDir + "/" + outFileName)
	if err != nil {
		return err
	}

	lang := mapLanguage(debug, language)

	_, err = file.WriteString("<tablistentry>")
	if err != nil {
		return err
	}
	_, err = file.WriteString("<tabname>" + lang + "</tabname>")
	if err != nil {
		return err
	}
	_, err = file.WriteString("<tabcontent>")
	if err != nil {
		return err
	}
	_, err = file.WriteString("<para>See the")
	if err != nil {
		return err
	}
	_, err = file.WriteString("<xi:include href=\"" + linkFileName + "\"/>")
	if err != nil {
		return err
	}
	_, err = file.WriteString("</para>")
	if err != nil {
		return err
	}
	_, err = file.WriteString("<tabcontent>")
	if err != nil {
		return err
	}
	_, err = file.WriteString("</tablistentry>")
	if err != nil {
		return err
	}

	file.Close()

	return nil
}

func isValidLanguage(debug bool, lang string) bool {
	for _, l := range globalConfig.LanguageDirs {
		if lang == l {
			return true
		}
	}

	return false
}

// Metadata defines the elements in metatdata.yaml
type Metadata struct {
	Files []struct {
		Path        string `yaml:"path"`
		Description string `yaml:"description"`
		Services    []struct {
			Service string   `yaml:"service"`
			Actions []string `yaml:"actions"`
		} `yaml:"services"`
	} `yaml:"files"`
}

func processMetadata(debug bool, path, outDir string) error {
	// The path should look something like:
	//    https://raw.githubusercontent.com/awsdocs/aws-doc-sdk-examples/doug-test-go-metadata/dotnetv3/ACM/metadata.yaml

	// So if we split by '/',
	parts := strings.Split(path, "/")
	length := len(parts)
	filename := parts[length-1]
	// service := parts[length-2]
	language := parts[length-3]

	var metadata Metadata

	results, err := http.Get(path)
	if err != nil {
		fmt.Println("Got an error getting the file:")
		fmt.Println(err)
		return err
	}
	defer results.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(results.Body)
	bytes := buf.Bytes()

	err = yaml.Unmarshal(bytes, &metadata)
	if err != nil {
		fmt.Println("Got an error unmarshalling data for path " + path)
		fmt.Println(err)
		return err
	}

	// Iterate through files.
	// Create a link and tab for services/operations
	for _, data := range metadata.Files {
		// If Description is "test", skip it
		if data.Description == "test" {
			debugPrint(debug, "Skipping test file "+data.Path)
			continue
		}

		for _, s := range data.Services {
			for _, a := range s.Actions {
				// If somehow we made it this far, but the action is "test", skip it
				if a == "test" {
					continue
				}

				err := createLinkFile(debug, s.Service, a, language, filename, path, outDir)
				if err != nil {
					return err
				}
				err = createTabFile(debug, s.Service, a, language, filename, path, outDir)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func processFiles(debug bool, input string, outDir string) error {
	//debugPrint(debug, "Input to processFiles:")
	//debugPrint(debug, input)

	filePrefix := "https://raw.githubusercontent.com/awsdocs/aws-doc-sdk-examples/" + globalConfig.Branch + "/"

	// Unmarshal string into tree
	var repoTree RepoTree

	err := json.Unmarshal([]byte(input), &repoTree)
	if err != nil {
		fmt.Println("Got an error unmarshalling input")
		return err
	}

	for _, leaf := range repoTree.Tree {
		// leaf.Path is something like
		//   LANGUAGE/[example_code/]SERVICE/[more path crap/]filename
		// Split up the path '/'
		parts := strings.Split(leaf.Path, "/")
		lang := parts[0]

		valid := isValidLanguage(debug, lang)

		if !valid {
			continue
		}

		// We only care about metadata.yaml or .metadata.yaml files
		length := len(parts)
		filename := parts[length-1]

		if filename != "metadata.yaml" && filename != ".metadata.yaml" {
			continue
		}

		// Skip dotnetv3/dynamodb for now
		isDynamoDotnet := strings.Contains(leaf.Path, "dotnetv3/dynamodb")

		if isDynamoDotnet {
			continue
		}

		debugPrint(debug, "Calling processMetadata with path:")
		debugPrint(debug, "  "+filePrefix+leaf.Path)

		err := processMetadata(debug, filePrefix+leaf.Path, outDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("    go run CreateZonbookIncludes.go -u NAME [-d] [-h]")
	fmt.Println(" where:")
	fmt.Println("    NAME      is the name of the GitHub user used to the GitHub API")
	fmt.Println("              the default is the value of UserName in config.json")
	fmt.Println("    -d        displays additional debugging information")
	fmt.Println("    -h        displays this error message and quits")
}

// Config defines the configuration values from config.json
type Config struct {
	UserName     string   `json:"UserName"`
	Branch       string   `json:"Branch"`
	LanguageDirs []string `json:"LanguageDirs"`
	Services     []string `json:"Services"`
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

func main() {
	outdir := "build"

	err := populateConfiguration()
	if err != nil {
		fmt.Println("Got an error parsing " + configFileName + ":")
		fmt.Println(err)
		return
	}

	userName := flag.String("u", globalConfig.UserName, "Your GitHub user name")
	debug := flag.Bool("d", false, "Whether to barf out more info. False by default.")
	help := flag.Bool("h", false, "Displays usage and quits")
	flag.Parse()

	if *help {
		usage()
		return
	}

	if *userName == "" {
		usage()
		return
	}

	if globalConfig.Branch == "" {
		globalConfig.Branch = "master"
	}

	err = createOutdir(*debug, outdir)
	if err != nil {
		fmt.Println("Could not create output directory " + outdir)
		return
	}

	if *debug {
		fmt.Println("User: " + *userName)
	}

	gitHubURL := "https://api.github.com"
	query := gitHubURL + "/repos/awsdocs/aws-doc-sdk-examples/git/trees/" + globalConfig.Branch + "?recursive=1"

	debugPrint(*debug, "Querying: ")
	debugPrint(*debug, query)

	jsonData := ""
	jsonValue, _ := json.Marshal(jsonData)

	request, err := http.NewRequest("GET", query, bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("Got an error creating HTTP request:")
		fmt.Println(err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/vnd.github.v3+json")

	request.SetBasicAuth(*userName, "")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}

	data, _ := ioutil.ReadAll(response.Body)

	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, data, "", "\t")
	if error != nil {
		fmt.Println("Got an error indenting JSON bytes:")
		fmt.Println(err)
		return
	}

	err = processFiles(*debug, prettyJSON.String(), outdir)
	if err != nil {
		fmt.Println("Got an error processing files:")
		fmt.Println(err)
	}
}
