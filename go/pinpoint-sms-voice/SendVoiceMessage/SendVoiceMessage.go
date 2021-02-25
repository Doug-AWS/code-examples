package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoice"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoice/types"
	"github.com/aws/aws-sdk-go/aws"
)

// PPSendVoiceMessageAPI defines the interface for the SendVoiceMessage function.
// We use this interface to test the function using a mocked service.
type PPSendVoiceMessageAPI interface {
	SendVoiceMessage(ctx context.Context,
		params *pinpointsmsvoice.SendVoiceMessageInput,
		optFns ...func(*pinpointsmsvoice.Options)) (*pinpointsmsvoice.SendVoiceMessageOutput, error)
}

// SendMsg sends an Amazon Pinpoint SMS message.
// Inputs:
//     c is the context of the method call, which includes the AWS Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a SendVoiceMessageOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to SendVoiceMessage.
func SendMsg(c context.Context, api PPSendVoiceMessageAPI, input *pinpointsmsvoice.SendVoiceMessageInput) (*pinpointsmsvoice.SendVoiceMessageOutput, error) {
	return api.SendVoiceMessage(c, input)
}

func main() {
	sourceNumber := flag.String("s", "", "The phone number that appears on recipients' devices when they receive the message")
	destinationNumber := flag.String("d", "", "The phone number that you want to send the voice message to")

	flag.Parse()

	if *sourceNumber == "" || *destinationNumber == "" {
		fmt.Println("You must supply a source and destination telephone number")
		fmt.Println("(-s SOURCE -d DESTINATION")
		return
	}

	// (import "github.com/aws/aws-sdk-go-v2/config")
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := pinpointsmsvoice.NewFromConfig(cfg)

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
		//CallInstructionsMessage = callInstructionsMessage,
		//PlainTextMessage = plainTextMessage,
		SSMLMessage: sSMLMessage,
	}

	input := &pinpointsmsvoice.SendVoiceMessageInput{
		// CallerId: sourceNumber, // optional
		// ConfigurationSetName:   aws.String("???"),
		Content:                content, // optional???
		DestinationPhoneNumber: destinationNumber,
		OriginationPhoneNumber: sourceNumber,
	}

	result, err := SendMsg(context.TODO(), client, input)

	fmt.Println("Message ID: " + *result.MessageId)

	fmt.Println("")
}
