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

func main() {
	config := flag.String("c", "abc", "Set to any value")
	flag.Parse()

	c := GetConfig(*config)

	fmt.Println("Config: " + c.config)
}
