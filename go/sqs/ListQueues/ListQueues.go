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
    result, err := svc.ListQueues(nil)
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
    sess := session.Must(session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    }))

    svc := sqs.New(sess)

    result, err := GetQueues(svc)
    if err != nil {
        fmt.Println("Got an error retrieving queue URLs:")
        fmt.Println(err)
        return
    }

    for i, url := range result.QueueUrls {
        fmt.Printf("%d: %s\n", i, *url)
    }
}
