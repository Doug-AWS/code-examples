package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Entry stores the values from the DynamoDB table
type Entry struct {
	Path        string `json:"path"`
	Action      string `json:"action"`
	SDK         string `json:"sdk"`
	Service     string `json:"service"`
	Target      string `json:"target"`
	Description string `json:"description"`
}

// Config stores the values from config.json
type Config struct {
	Table    string   `json:"TableName"`    // The name of the DynamoDB table
	Services []string `json:"ServiceNames"` // The list of services
	Sdks     []string `json:"SdkNames"`     // The list of SDKs, by language extension
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

	if globalConfig.Table == "" {
		msg := "You musts supply a value for TableName in " + configFileName
		return errors.New(msg)
	}

	return nil
}

func debugPrint(debug bool, s string) {
	if debug {
		fmt.Println(s)
	}
}

func getSdkEntityName(debug bool, sdk string) (string, error) {
	retVal := ""
	var err error

	switch sdk {
	case "java":
		retVal = "JavaV2long"
		break
	case "js":
		retVal = "JSBlong"
		break
	case "go":
		retVal = "Golong"
		break
	case "php":
		retVal = "PHPlong"
		break
	case "rb":
		retVal = "Rubylong"
		break

	default:
		msg := "Unidentified SDK: " + sdk
		err = errors.New(msg)
		break
	}

	return retVal, err
}

func getServiceEntityName(debug bool, service string) (string, error) {
	retVal := ""
	var err error

	switch service {
	case "cloudwatch":
		retVal = "CloudWatch"
		break
	case "dynamodb":
		retVal = "DynamoDB"
		break
	case "ec2":
		retVal = "EC2"
		break
	case "iam":
		retVal = "IAM"
		break
	case "kinesis":
		retVal = "Kinesis"
		break
	case "kms":
		retVal = "KMS"
		break
	case "rekognition":
		retVal = "Rekognition"
		break
	case "s3":
		retVal = "S3"
		break
	case "sqs":
		retVal = "SQS"
		break
	case "sns":
		retVal = "SNS"
		break
	case "ssm":
		retVal = "SSM"
		break
	case "sts":
		retVal = "STS"
		break

	default:
		msg := "Unidentified service: " + service
		err = errors.New(msg)

	}

	return retVal, err
}

/*
   The entity for sns-code-examples (action == section);
   I'm only showing go, but you get the drift:

   <!ENTITY sns-code-examples '<table class="table">
      <title>Amazon SNS code examples in AWS SDK developer guides</title>
      <tgroup cols="1">
        <tbody>
          <row>
            <entry>
              <para><ulink url="https://aws.github.io/aws-sdk-go-v2/docs/code-examples/sns/">&Golong;</ulink></para>
            </entry>
          </row>
        </tbody>
      </tgroup>
    </table>'>
*/

func createEntity(debug bool, service string, action string, actionEntries []Entry, f *os.File) error {
	if action == "" {
		return nil
	}

	debugPrint(debug, "Got "+strconv.Itoa(len(actionEntries))+" entry/entries for action "+action)

	// Sort entries by sdk
	sort.Slice(actionEntries[:], func(i, j int) bool {
		return actionEntries[i].SDK < actionEntries[j].SDK
	})

	entity := ""
	serviceEntity, err := getServiceEntityName(debug, service) // SNS for SNS
	if err != nil {
		fmt.Println("Got an error retrieving the entity name for the " + service + " service")
	}

	if action == "section" {
		debugPrint(debug, "Creating section entity for "+service)
		entity += "<!ENTITY " + service + "-code-examples '<table class=\"table\">\n"
		entity += "   <title>&" + serviceEntity + "; code examples in AWS SDK developer guide</title>\n"
	} else {
		debugPrint(debug, "Creating section/action entity for "+service+"/"+action)
		entity += "<!ENTITY " + service + "-" + action + "-code-examples '<table class=\"table\">\n"
		entity += "   <title>&" + serviceEntity + "; " + action + " code examples in AWS SDK developer guides</title>\n"
	}

	entity += "   <tgroup cols=\"1\">\n"
	entity += "     <tbody>\n"

	for _, a := range actionEntries {
		debugPrint(debug, "Path:        "+a.Path)
		debugPrint(debug, "Description: "+a.Description)
		// sdkEntity, err := getSdkEntityName(debug, a.SDK) // golang for go
		if err != nil {
			fmt.Println("Got an error retrieving the entity name for the " + a.SDK + " SDK")
		}

		entity += "       <row>\n"
		entity += "         <entry>\n"
		entity += "           <para><ulink url=\"" + a.Path + "\">" + a.Description + "</ulink></para>\n"
		entity += "         </entry>\n"
		entity += "       </row>\n"
	}

	entity += "     </tbody>\n"
	entity += "   </tgroup>\n"
	entity += " </table>'>\n"

	_, err = f.WriteString(entity + "\n")
	if err != nil {
		fmt.Println("Got an error creating entity for " + service + " action " + action)
		return err
	}

	return nil
}

func createServiceEntities(debug bool, table string, service string) error {
	outFileName := service + ".ent"
	debugPrint(debug, "Creating entities for "+service+" service in "+outFileName)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got a configuration error")
		return err
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)

	// Get items for that service
	filt := expression.Name("service").Equal(expression.Value(service))

	// Get back the title and rating (we know the year).
	proj := expression.NamesList(expression.Name("path"), expression.Name("action"), expression.Name("service"), expression.Name("sdk"), expression.Name("target"), expression.Name("description"))

	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		fmt.Println("Got error building expression:")
		return err
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(table),
	}

	resp, err := dynamodbClient.Scan(context.TODO(), input)
	if err != nil {
		fmt.Println("Got error calling Scan: ")
		return err
	}

	entries := []Entry{}

	err = attributevalue.UnmarshalListOfMaps(resp.Items, &entries)
	if err != nil {
		fmt.Println("Got an error unmarshalling table entries")
		return err
	}

	if debug {
		fmt.Println("path:        " + entries[0].Path)
		fmt.Println("Action:      " + entries[0].Action)
		fmt.Println("SDK:         " + entries[0].SDK)
		fmt.Println("Service:     " + entries[0].Service)
		fmt.Println("Target:      " + entries[0].Target)
		fmt.Println("Description: " + entries[0].Description)
	}

	debugPrint(debug, "Creating output file: "+outFileName)

	f, err := os.OpenFile(outFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Got an error opening " + outFileName)
		return err
	}

	defer f.Close()

	// Sort entries by Action
	sort.Slice(entries[:], func(i, j int) bool {
		return entries[i].Action < entries[j].Action
	})

	initAction := ""

	i := 0
	var actionEntries []Entry

	for i < len(entries) {
		if entries[i].Action != initAction {
			// Create entity from set of actions
			err := createEntity(debug, service, initAction, actionEntries, f)
			if err != nil {
				fmt.Println("Got an error creating entity")
				return err
			}

			// Reset actionEntries
			actionEntries = nil
			actionEntries = make([]Entry, 1)

			initAction = entries[i].Action

			debugPrint(debug, "Looking for entries with "+initAction+" action")

			actionEntries[0] = entries[i]
		} else {
			actionEntries = append(actionEntries, entries[i])
		}

		i++
	}

	// We have to create an entity for the last item
	err = createEntity(debug, service, initAction, actionEntries, f)
	if err != nil {
		fmt.Println("Got an error creating entity")
		return err
	}

	return nil
}

func main() {
	debug := flag.Bool("d", false, "Whether to barf out more info")

	flag.Parse()

	debugPrint(*debug, "Debugging enabled")

	err := populateConfiguration()
	if err != nil {
		fmt.Println("Could not parse " + configFileName)
		return
	}

	for _, e := range globalConfig.Services {
		err := createServiceEntities(*debug, globalConfig.Table, e)
		if err != nil {
			fmt.Println("Could not create entities for " + e + " service")
		}
	}
}
