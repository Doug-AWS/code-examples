package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

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

// Create a Zonbook XML file containing a link to a folder, such as javav2/example_code/sns.
// (as that is where the user can find a README.md that describes all of the code examples for that language/service
//  AND a list of files).
// The format of the filename is SERVICE_LANGUAGE_link.xml
func createLinkFile(debug bool, service, language, path, outDir string) (string, error) {
	outFileName := service + "_" + language + "_link.xml"

	debugPrint(debug, "Creating link file "+outFileName)

	file, err := os.Create(outDir + "/" + outFileName)
	if err != nil {
		msg := "Could not create file " + outFileName
		return "", errors.New(msg)
	}

	// Get service entity from map file
	svc := serviceMap[service+"-entity"]
	sdk := serviceMap[language+"-sdk"]

	outString := "<ulink url=\"" + path + "\">Code examples for " + svc + "in the " + sdk + "SDK</ulink>"

	debugPrint(debug, "Creating link:")
	debugPrint(debug, outString)

	_, err = file.WriteString(outString)
	if err != nil {
		msg := "Could not write to file " + outFileName
		return "", errors.New(msg)
	}

	file.Close()

	return outString, nil
}

// Takes a language directory, like gov2, and returns the language extension, like go
func mapLanguageToExtension(debug bool, language string) string {
	lang := ""

	switch language {
	case "dotnetv3":
		lang = "cs"
	case "gov2":
		lang = "go"
	}

	return lang
}

// Takes a language directory, like gov2, and returns the tab name for it, like Go
func mapLanguage(debug bool, language string) string {
	lang := ""

	switch language {
	case "dotnetv3":
		lang = ".NET/C#"
	case "gov2":
		lang = "Go"
	}

	return lang
}

func LcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// Creates a section, where all of the code examples for a given service.api are presented in a tablist.

// Creates a Zonbook section with a tablist where each tab entry is the combination of service, operation, and language.
// Every entry in files should be the same service and operation,
// So we use the values of the first.
// Note that there might be more than one tablistentry for a specific language.
/*
    <?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE chapter PUBLIC "-//OASIS//DTD DocBook XML V4.5//EN" "file://zonbook/docbookx.dtd"[
    <!ENTITY % xinclude SYSTEM "file://AWSShared/common/xinclude.mod">
    %xinclude;
    <!ENTITY % phrases-shared SYSTEM "file://AWSShared/common/phrases-shared.ent">
    %phrases-shared;
]>

<section id="example-SERVICE-ACTION">
    <info>
        <title id="example-SERVICE-ACTION.title">
            ACTION-ENTRY-FROM-MAP-FILE-TITLE</title>
        <titleabbrev>ACTION-TITLE-ENTRY-FROM-MAP-FILE using an &AWS; SDK</titleabbrev>
    </info>
    <para>The following examples show how to write code to ACTION-ENTRY-TITLE-ENTRY-FROM-MAP-FILE-WITH-FIRST-LETTER-LOWERCASE using  &AWS; SDKs</para>
    <tablist>
        <tablistentry region="SEA;IAD;-BJS;-DCA;-LCK;"> FOR-CPP-OR-GO
            <tabname>LANGUAGE</tabname>
            <tabcontent><xi:include href="file://AWSShared/code-samples/docs/sns_Publish_php.xml"/></tabcontent>
        </tablistentry>
    </tablist>
    <para>For a complete list of &AWS; SDK developer guides and code examples, including help getting started and information about previous versions,
        see <xref linkend="sdk-general-information-section" endterm="sdk-general-information-section.title"></xref>.</para>
</section>
*/
// The format of the name is service_Action_section.xml.
// So for S3 CreateBucket: s3_CreateBucket_section.xml.
func createOperationChapter(debug bool, files []XFile, outDir string) error {
	// Since every files entry is the same service and operation, use the first on
	service := files[0].service
	operation := files[0].operation

	// Get service-action entry from map file
	/*
		// Get service entity from map file
		svc := serviceMap[service+"-entity"]
		sdk := serviceMap[language+"-sdk"]
	*/
	desc := serviceMap[service+"-"+operation]
	desc = strings.TrimRight(desc, ".")
	title := serviceMap[service+"-"+operation+"-title"]
	titleLower := LcFirst(title)

	outFileName := service + "_" + operation + "_chapter.xml"

	debugPrint(debug, "Creating section file "+outFileName)

	file, err := os.Create(outDir + "/" + outFileName)
	if err != nil {
		return err
	}

	_, err = file.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("<!DOCTYPE chapter PUBLIC \"-//OASIS//DTD DocBook XML V4.5//EN\" \"file://zonbook/docbookx.dtd\"[\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    <!ENTITY % xinclude SYSTEM \"file://AWSShared/common/xinclude.mod\">\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    %xinclude;\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    <!ENTITY % phrases-shared SYSTEM \"file://AWSShared/common/phrases-shared.ent\">\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    %phrases-shared;\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("]>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("<section id=\"example-" + service + "-" + operation + "\">\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    <info>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        <title id=\"example" + service + "-" + operation + ".title\">\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("            " + desc + " using an &AWS; SDK</title>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        <titleabbrev>" + title + "</titleabbrev>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    </info>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    <para>The following examples show how to write code that " + titleLower + " using  &AWS; SDKs.</para>\n")
	if err != nil {
		return err
	}

	// Create the tablist
	language := files[0].language
	lang := mapLanguage(debug, language)

	if language == "cpp" || language == "go" {
		_, err = file.WriteString("    <tablist region=\"SEA;IAD;-BJS;-DCA;-LCK;\">\n")
		if err != nil {
			return err
		}
	} else {
		_, err = file.WriteString("    <tablist>\n")
		if err != nil {
			return err
		}
	}

	_, err = file.WriteString("      <tablistentry>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        <tabname>" + lang + "</tabname>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        <tabcontent>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("          <para>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("            See the \n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("            " + files[0].link + "\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("            code example.\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("          </para>\n")
	if err != nil {
		return err
	}

	for _, f := range files {
		if language != f.language {
			language = f.language
			lang = mapLanguage(debug, language)

			// Close initial tablistentry tag
			_, err = file.WriteString("        </tabcontent>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("      </tablistentry>\n")
			if err != nil {
				return err
			}

			// Start new one
			_, err = file.WriteString("      <tablistentry>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("        <tabname>" + lang + "</tabname>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("        <tabcontent>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("          <para>\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("            See the \n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("            " + f.link + "\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("            code example.\n")
			if err != nil {
				return err
			}

			_, err = file.WriteString("          </para>\n")
			if err != nil {
				return err
			}
		}
	}

	// Close final tablistentry tag and taglist tag
	_, err = file.WriteString("        </tabcontent>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("     </tablistentry>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("    </tablist>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("	    <para>For a complete list of &AWS; SDK developer guides and code examples, including help getting started and information about previous versions,\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("        see <xref linkend=\"sdk-general-information-section\" endterm=\"sdk-general-information-section.title\"></xref>.</para>\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("</section>\n")
	if err != nil {
		return err
	}

	return nil
}

func createTabFile(debug bool, service, operation, language, link, outDir string) error {
	outFileName := service + "_" + operation + "_" + language + "_tab.xml"

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

func isValidLanguage(debug bool, languageDirs []string, lang string) bool {
	for _, l := range languageDirs {
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
		SdkLink     string `yaml:"sdklink"`
		Services    []struct {
			Service string   `yaml:"service"`
			Actions []string `yaml:"actions"`
		} `yaml:"services"`
	} `yaml:"files"`
}

func processFiles(debug bool, metadata Metadata, local bool, branch, path, language, outDir string) error {
	filePrefix := "https://github.com/awsdocs/aws-doc-sdk-examples/tree/" + branch + "/"

	if local {
		filePrefix = ""
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
			// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			// Skip everything but SNS for now
			if s.Service != "sns" {
				return nil
			}

			// Since path contains metadata.yaml as the last element,
			// chop /metadata.yaml off.
			p := strings.TrimRight(filePrefix+path, "metadata.yaml")
			p = strings.TrimRight(p, ".") // In case it was .metadata.yaml

			link, err := createLinkFile(debug, s.Service, language, p, outDir)
			if err != nil {
				fmt.Println("Got an error creating link file:")
				fmt.Println(err.Error())
				return err
			}

			for _, a := range s.Actions {
				// If somehow we made it this far, but the action is "test", skip it
				if a == "test" {
					continue
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

func processLocalMetadata(debug bool, branch, path, outDir string) error {
	// The path should look something like:
	//  C:\GitHub\aws-doc-sdk-examples\dotnetv3\SNS\metadata.yaml
	// So split it by '\' and make sure the last part is metadata.yaml or .metadata.yaml
	// and the next-to-last part is a legit service name.
	parts := strings.Split(path, "\\")

	if parts[len(parts)-1] != "metadata.yaml" && parts[len(parts)-1] != ".metadata.yaml" {
		msg := "The path does not end in a metadata filename"
		return errors.New(msg)
	}

	if parts[len(parts)-2] != "sns" && parts[len(parts)-2] != "SNS" {
		return nil
	}

	// Get contents of file and stuff it into a Metadata struct
	var metadata Metadata

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Got an error opening " + path)
		return err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&metadata); err != nil {
		// fmt.Println("Skipping metadata file due to YAML parsing errors: " + path)
		return err
	}

	debugPrint(debug, "Found SNS metadata file "+path)
	language := mapLanguageToExtension(debug, parts[3])

	// Wade through metadata file and create entries
	err = processFiles(debug, metadata, false, branch, path, language, outDir)
	if err != nil {
		fmt.Println("Got an error processing local files:")
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func processRemoteMetadata(debug bool, branch, path, outDir string) error {
	// debugPrint(debug, "")
	// debugPrint(debug, "Processing file: "+path)

	metafilePrefix := "https://raw.githubusercontent.com/awsdocs/aws-doc-sdk-examples/" + branch + "/"
	// The path should look something like:
	//    https://raw.githubusercontent.com/awsdocs/aws-doc-sdk-examples/doug-test-go-metadata/dotnetv3/ACM/metadata.yaml

	// So if we split by '/',
	parts := strings.Split(metafilePrefix+path, "/")
	length := len(parts)
	// filename := parts[length-1]
	// service := parts[length-2]
	language := parts[length-3]

	language = mapLanguageToExtension(debug, language)

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

	err = processFiles(debug, metadata, false, branch, path, language, outDir)
	if err != nil {
		fmt.Println("Got an error processing remote files:")
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func processRemoteFiles(debug bool, config Config, input string) error {
	debugPrint(debug, "Input to processRemoteFiles:")
	debugPrint(debug, input)

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

		valid := isValidLanguage(debug, config.LanguageDirs, lang)

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

		err := processRemoteMetadata(debug, config.Branch, leaf.Path, config.Outdir)
		if err != nil {
			return err
		}
	}

	return nil
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("    go run CreateZonbookIncludes.go [-u NAME] [-d] [-h]")
	fmt.Println(" where:")
	fmt.Println("    NAME      is the name of the GitHub user used for basic authentication for the GitHub API (when running remotely)")
	fmt.Println("              the default is the value of UserName in config.json")
	fmt.Println("    -d        displays additional debugging information")
	fmt.Println("    -h        displays this error message and quits")
}

func addEntryToServiceMap(debug bool, s string) error {
	// Skip comments
	if strings.HasPrefix(s, "#") {
		return nil
	}

	// Split string by ":"
	parts := strings.Split(s, ":")

	if len(parts) != 2 {
		msg := "Incorrect format: " + s
		return errors.New(msg)
	}

	// Make sure both parts don't have any spaces around them.
	p0 := strings.TrimSpace(parts[0])
	p1 := strings.TrimSpace(parts[1])

	//debugPrint(debug, p0+": "+p1)

	serviceMap[p0] = p1

	return nil
}

func fillServiceMap(debug bool) error {
	serviceMap = make(map[string]string)

	// Open mapFile, read every line, split each by ":", add to map
	inFile, err := os.Open(mapFile)
	if err != nil {
		fmt.Println("Got an error reading " + mapFile)
		return err
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		err := addEntryToServiceMap(debug, scanner.Text())
		if err != nil {
			fmt.Println("Got an error adding entry to service map:")
			fmt.Println(err)
			return err
		}
	}

	debugPrint(debug, "Added "+strconv.Itoa(len(serviceMap))+" entries to the service map")

	return nil
}

// Config defines the configuration values from config.json
type Config struct {
	UserName     string   `json:"UserName"`
	Branch       string   `json:"Branch"`
	LanguageDirs []string `json:"LanguageDirs"`
	LocalRoot    string   `json:"LocalRoot"`
	Mode         string   `json:"Mode"`
	Outdir       string   `json:"Outdir"`
	Services     []string `json:"Services"`
	MapFile      string   `json:"MapFIle"`
}

var configFileName = "config.json"

// var globalConfig Config

func populateConfiguration() (Config, error) {
	var globalConfig Config

	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return globalConfig, err
	}

	text := string(content)

	err = json.Unmarshal([]byte(text), &globalConfig)
	if err != nil {
		return globalConfig, err
	}

	return globalConfig, nil
}

func isCorrectGitubBranch(debug bool, branch, dir string) error {
	// If running in local mode, ensure we are scanning the correct branch.
	cmd := exec.Command("git", "branch", "--show")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	s := string(out)
	// Strip off \n
	s = strings.TrimRight(s, "\n")

	debugPrint(debug, "Found local branch "+s)

	if s != branch {
		msg := "'" + s + "' is not the branch '" + branch + "'"
		return errors.New(msg)
	}

	// Make sure the branch is up-to-date
	cmd = exec.Command("git", "pull")
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

type RateLimit struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}

var rateLimit RateLimit

func checkForRateLimit(s string) error {
	//text := string(content)

	err := json.Unmarshal([]byte(s), &rateLimit)
	if err != nil {
		return err
	}

	if rateLimit.DocumentationURL == "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting" {
		msg := "Rate limit exceeded"
		return errors.New(msg)
	}

	return nil
}

// Get info from local GitHub clone
func runLocally(debug bool, config Config) error {
	/*
	   1. Iterate through local clone
	   2. Find metadata.yaml (or .metadata.yaml) files
	   3. Send them off to processLocalMetadata
	*/

	// Print out current directory so we can figure out path
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not retrieve current directory")
	} else {
		fmt.Println("Current directory: " + dir)
	}

	// Recursively look for metadata.yaml or .metadata.yaml files
	debugPrint(debug, "Searching local folder for metadata files:")
	debugPrint(debug, config.LocalRoot)

	searchDir := config.LocalRoot

	fileList := []string{}
	err = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})

	for _, file := range fileList {
		// Split file by '\' and see if the last part is metadata.yaml or .metadata.yaml
		parts := strings.Split(file, "\\")
		if parts[len(parts)-1] == "metadata.yaml" || parts[len(parts)-1] == ".metadata.yaml" {
			err = processLocalMetadata(debug, config.Branch, file, config.Outdir)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

// Query GitHub repo using REST API
func runRemote(debug bool, config Config) error {
	gitHubURL := "https://api.github.com"
	query := gitHubURL + "/repos/awsdocs/aws-doc-sdk-examples/git/trees/" + config.Branch + "?recursive=1"

	// debugPrint(*debug, "Querying: ")
	// debugPrint(*debug, query)

	jsonData := ""
	jsonValue, _ := json.Marshal(jsonData)

	request, err := http.NewRequest("GET", query, bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("Got an error creating HTTP request:")
		fmt.Println(err)
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/vnd.github.v3+json")

	request.SetBasicAuth(config.UserName, "")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return err
	}

	data, _ := ioutil.ReadAll(response.Body)

	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, data, "", "\t")
	if error != nil {
		fmt.Println("Got an error indenting JSON bytes:")
		fmt.Println(err)
		return err
	}

	s := prettyJSON.String()

	// Check for rate limit:
	err = checkForRateLimit(s)
	if err != nil {
		fmt.Println("Hit the GitHub API rate limit")
		return err
	}

	err = processRemoteFiles(debug, config, s)
	if err != nil {
		fmt.Println("Got an error processing files:")
		fmt.Println(err)
	}

	return nil
}

var mapFile string
var serviceMap map[string]string

func main() {
	config, err := populateConfiguration()
	if err != nil {
		fmt.Println("Got an error parsing " + configFileName + ":")
		fmt.Println(err)
		return
	}

	userName := flag.String("u", config.UserName, "Your GitHub user name, for basic authentication")
	debug := flag.Bool("d", false, "Whether to barf out more info. False by default.")
	help := flag.Bool("h", false, "Displays usage and quits")
	flag.Parse()

	if *help {
		usage()
		return
	}

	if config.Mode == "remote" && *userName == "" {
		usage()
		return
	}

	if config.Branch == "" {
		config.Branch = "master"
	}

	if config.Mode == "" || config.LocalRoot == "" {
		config.Mode = "Remote"
	}

	mapFile = config.MapFile

	if config.Mode == "local" {
		err := isCorrectGitubBranch(*debug, config.Branch, config.LocalRoot)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	err = createOutdir(config.Outdir)
	if err != nil {
		fmt.Println("Could not create output directory " + config.Outdir)
		return
	}

	err = fillServiceMap(*debug)
	if err != nil {
		fmt.Println("Got an error parsing " + mapFile)
		fmt.Println(err)
		return
	}

	debugPrint(*debug, "User: "+*userName)

	// Get metadata.yaml (or .metadata.yaml) files locally or via GitHub REST API:
	if config.Mode == "local" {
		debugPrint(*debug, "Running in local mode")
		err := runLocally(*debug, config)
		if err != nil {
			fmt.Println("Got error running in local mode:")
			fmt.Println(err.Error())
			return
		}
	} else {
		debugPrint(*debug, "Running in remote mode")
		err := runRemote(*debug, config)
		if err != nil {
			fmt.Println("Got error running in local mode:")
			fmt.Println(err.Error())
			return
		}
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

	fmt.Println("Found", len(XFiles), "operations")

	service := ""
	operation := ""

	var OFiles []XFile

	for _, f := range XFiles {
		if f.service != service {
			// If OFiles isn't empty, create a section from the entries
			if OFiles != nil {
				err := createOperationChapter(*debug, OFiles, config.Outdir)
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
				err := createOperationChapter(*debug, OFiles, config.Outdir)
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
