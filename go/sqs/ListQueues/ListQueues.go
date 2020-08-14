package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// Client is the client
type Client struct {
	client sqsiface.SQSAPI
	config string
}

// GetConfig gets the configuration
func GetConfig(s string) Client {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)

	return Client{
		client: svc,
		config: s,
	}
}

// GetQueues returns a list of queue names
func GetQueues(svc sqsiface.SQSAPI) (*sqs.ListQueuesOutput, error) {
	// snippet-start:[sqs.go.list_queues.call]
	result, err := svc.ListQueues(nil)
	// snippet-end:[sqs.go.list_queues.call]
	if err != nil {
		return nil, err
	}

	return result, nil
}

func main() {
	config := flag.String("c", "", "Set to \"none\" to just exit")
	flag.Parse()

	c := GetConfig(*config)

	if c.config == "none" {
		fmt.Println("Config: " + c.config)
		return
	}

	// Create a session that gets credential values from ~/.aws/credentials
	// and the default region from ~/.aws/config
	// and a service client using that session
	// snippet-start:[sqs.go.list_queues.sess]
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)
	// snippet-end:[sqs.go.list_queues.sess]

	result, err := GetQueues(svc)
	if err != nil {
		fmt.Println("Got an error retrieving queue URLs:")
		fmt.Println(err)
		return
	}

	// snippet-start:[sqs.go.list_queues.display]
	for i, url := range result.QueueUrls {
		fmt.Printf("%d: %s\n", i, *url)
	}
	// snippet-end:[sqs.go.list_queues.display]
}

// snippet-end:[sqs.go.list_queues]
