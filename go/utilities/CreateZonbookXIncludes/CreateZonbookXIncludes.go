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
	"sort"
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
func createOutdir(outDir string) error {
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

// Create a Zonbook XML file containing link to a code example that can be XINCLUDEd in a tab.
// The format of the filename is SERVICE-OPERATION-LANGUAGE-link.xml
func createLinkFile(debug bool, service, operation, language, sourceFileName, sourceFileURL, outDir string) (string, error) {
	outFileName := service + "-" + operation + "-" + language + "-link.xml"

	// debugPrint(debug, "Creating link file "+outFileName)

	file, err := os.Create(outDir + "/" + outFileName)
	if err != nil {
		return "", err
	}

	outString := "<ulink url=\"" + sourceFileURL + sourceFileName + "\">" + operation + "</ulink>"

	// debugPrint(debug, "Creating link:")
	// debugPrint(debug, outString)

	_, err = file.WriteString(outString)
	if err != nil {
		return "", err
	}

	file.Close()

	return outString, nil
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

// Creates a Zonbook tablist where each tab entry is the combination of service, operation, and language.
// Every entry in files should be the same service and operation,
// So we use the values of the first.
// Note that there might be more than one tablistentry for a specific language.
/*
    <tablist>
      <tablistentry>
        <tabname>LANGUAGE</tabname>
        <tabcontent>
          <para>
            For more information, see
            <ulink url="LINK">DESCRIPTION</ulink>
          </para>
        </tabcontent>
        </tablistentry>
	</tablist>
*/
func createOperationTabList(debug bool, files []XFile, outDir string) error {
	outFileName := files[0].service + "-" + files[0].operation + "-tablist.xml"

	debugPrint(debug, "Creating tablist file "+outFileName)

	file, err := os.Create(outDir + "/" + outFileName)
	if err != nil {
		return err
	}

	_, err = file.WriteString("<tablist>\n")
	if err != nil {
		return err
	}

	language := files[0].language
	lang := mapLanguage(debug, language)

	_, err = file.WriteString("  <tablistentry>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    <tabname>" + lang + "</tabname>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    <tabcontent>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("      <para>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        See the \n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        " + files[0].link + "\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        code example.\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("      </para>\n")
	if err != nil {
		return err
	}

	for _, f := range files {
		if language != f.language {
			language = f.language
			lang = mapLanguage(debug, language)

			// Close initial tablistentry tag
			_, err = file.WriteString("    </tabcontent>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("  </tablistentry>\n")
			if err != nil {
				return err
			}

			// Start new one
			_, err = file.WriteString("  <tablistentry>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("    <tabname>" + lang + "</tabname>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("    <tabcontent>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("      <para>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("        See the \n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("        " + f.link + "\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("        code example.\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("      </para>\n")
			if err != nil {
				return err
			}
		}
	}

	// Close final tablistentry tag and taglist tag
	_, err = file.WriteString("    </tabcontent>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("  </tablistentry>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("</tablist>\n")
	if err != nil {
		return err
	}

	return nil
}

func createTabFile(debug bool, service, operation, language, link, outDir string) error {
	outFileName := service + "-" + operation + "-" + language + "-tab.xml"

	// debugPrint(debug, "Creating tab file "+outFileName)

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
	_, err = file.WriteString("<para>See the ")
	if err != nil {
		return err
	}
	_, err = file.WriteString(link)
	if err != nil {
		return err
	}
	_, err = file.WriteString(" code example.</para>")
	if err != nil {
		return err
	}
	_, err = file.WriteString("</tabcontent>")
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

type XFile struct {
	service   string
	operation string
	language  string
	link      string
}

var XFiles []XFile

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
	// debugPrint(debug, "")
	// debugPrint(debug, "Processing file: "+path)
	// The path should look something like:
	//    https://raw.githubusercontent.com/awsdocs/aws-doc-sdk-examples/doug-test-go-metadata/dotnetv3/ACM/metadata.yaml

	// So if we split by '/',
	parts := strings.Split(path, "/")
	length := len(parts)
	// filename := parts[length-1]
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
	_, err = buf.ReadFrom(results.Body)
	if err != nil {
		fmt.Println("Got an error getting bytes from result")
		return err
	}

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
			// debugPrint(debug, "Skipping test file "+data.Path)
			continue
		}

		for _, s := range data.Services {
			for _, a := range s.Actions {
				// If somehow we made it this far, but the action is "test", skip it
				if a == "test" {
					continue
				}

				// Since path contains metadata.yaml as the last element,
				// chop /metadata.yaml off.
				p := strings.TrimRight(path, "metadata.yaml")
				p = strings.TrimRight(p, ".") // In case it was .metadata.yaml

				link, err := createLinkFile(debug, s.Service, a, language, data.Path, p, outDir)
				if err != nil {
					return err
				}

				// Add entry to list of files
				xf := XFile{service: s.Service, operation: a, language: language, link: link}
				XFiles = append(XFiles, xf)

				err = createTabFile(debug, s.Service, a, language, link, outDir)

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

		// debugPrint(debug, "Calling processMetadata with path:")
		// debugPrint(debug, "  "+filePrefix+leaf.Path)

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
	fmt.Println("    NAME      is the name of the GitHub user used for basic authentication for the GitHub API")
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

	userName := flag.String("u", globalConfig.UserName, "Your GitHub user name, for basic authentication")
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

	err = createOutdir(outdir)
	if err != nil {
		fmt.Println("Could not create output directory " + outdir)
		return
	}

	if *debug {
		fmt.Println("User: " + *userName)
	}

	gitHubURL := "https://api.github.com"
	query := gitHubURL + "/repos/awsdocs/aws-doc-sdk-examples/git/trees/" + globalConfig.Branch + "?recursive=1"

	// debugPrint(*debug, "Querying: ")
	// debugPrint(*debug, query)

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

	sort.Slice(XFiles, func(i, j int) bool {
		if XFiles[i].service < XFiles[j].service {
			return true
		} else if XFiles[i].service > XFiles[j].service {
			return false
		} else {
			if XFiles[i].operation < XFiles[j].operation {
				return true
			} else {
				if XFiles[i].operation > XFiles[j].operation {
					return false
				}
			}
		}
		return XFiles[i].language < XFiles[j].language
	})

	fmt.Println("Found ", len(XFiles), "operations")

	service := ""
	operation := ""

	var OFiles []XFile

	for _, f := range XFiles {
		if f.service != service {
			// If OFiles isn't empty, create a tablist from the entries
			if OFiles != nil {
				err := createOperationTabList(*debug, OFiles, outdir)
				if err != nil {
					fmt.Println("Got an error creating tablist")
					fmt.Println(err.Error())
					return
				}

				// Reset OFiles
				OFiles = nil
			}
		}

		// We have the same service, but do we have the same operation?

		service = f.service
		if f.operation != operation {
			// If OFiles isn't empty, create a tablist from the entries
			if OFiles != nil {
				err := createOperationTabList(*debug, OFiles, outdir)
				if err != nil {
					fmt.Println("Got an error creating tablist")
					fmt.Println(err.Error())
					return
				}

				// Reset OFiles
				OFiles = nil
			}

			operation = f.operation
		}

		// We have the same service and operation, so append them to the list
		ofile := XFile{service: f.service, operation: f.operation, language: f.language, link: f.link}
		OFiles = append(OFiles, ofile)
	}
}
