package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoice"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoice/types"
	"github.com/aws/aws-sdk-go/aws"
)

type PPSendVoiceMessageImpl struct{}

func (dt PPSendVoiceMessageImpl) SendVoiceMessage(ctx context.Context,
	params *pinpointsmsvoice.SendVoiceMessageInput,
	optFns ...func(*pinpointsmsvoice.Options)) (*pinpointsmsvoice.SendVoiceMessageOutput, error) {

	output := &pinpointsmsvoice.SendVoiceMessageOutput{
		MessageId: aws.String("1234567890"),
	}

	return output, nil
}

type Config struct {
	SourceNumber      string `json:"SourceNumber"`
	DestinationNumber string `json:"DestinationNumber"`
}

var configFileName = "config.json"

var globalConfig Config

func populateConfiguration(t *testing.T) error {
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	text := string(content)

	err = json.Unmarshal([]byte(text), &globalConfig)
	if err != nil {
		return err
	}

	if globalConfig.SourceNumber == "" {
		msg := "You must supply a SourceNumber value in " + configFileName
		return errors.New(msg)
	}

	if globalConfig.DestinationNumber == "" {
		msg := "You must supply a DestinationNumber value in " + configFileName
		return errors.New(msg)
	}

	t.Log("Source number:      " + globalConfig.SourceNumber)
	t.Log("Destination number: " + globalConfig.DestinationNumber)

	return nil
}

func TestCreateBucket(t *testing.T) {
	thisTime := time.Now()
	nowString := thisTime.Format("2006-01-02 15:04:05 Monday")
	t.Log("Starting unit test at " + nowString)

	err := populateConfiguration(t)
	if err != nil {
		t.Fatal(err)
	}

	api := &PPSendVoiceMessageImpl{}

	sSMLMessage := &types.SSMLMessageType{
		LanguageCode: aws.String("en-US"), // Can this be different from the previous value?
		Text: aws.String("<speak>This is a test message sent from " +
			"<emphasis>Amazon Pinpoint</emphasis> " +
			"using the " +
			"<break strength='weak'/>" +
			"AWS SDK for Go. " +
			"<amazon:effect phonation='soft'>Thank you for listening.</amazon:effect></speak>"),
		VoiceId: aws.String("Joey"),
	}

	content := &types.VoiceMessageContent{
		SSMLMessage: sSMLMessage,
	}

	input := &pinpointsmsvoice.SendVoiceMessageInput{
		Content:                content,
		DestinationPhoneNumber: &globalConfig.DestinationNumber,
		OriginationPhoneNumber: &globalConfig.SourceNumber,
	}

	result, err := SendMsg(context.Background(), *api, input)
	if err != nil {
		fmt.Println("Got an error calling SendVoiceMessage:")
		fmt.Println(err)
		return
	}

	fmt.Println("Message ID: " + *result.MessageId)

	fmt.Println("")
}
