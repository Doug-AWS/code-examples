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

func verbosePrint(v bool, s string) {
	if v {
		fmt.Println(s)
	}
}

func main() {
	sourceNumber := flag.String("s", "", "The phone number that appears on recipients' devices when they receive the message")
	destinationNumber := flag.String("d", "", "The phone number that you want to send the voice message to")
	// originNumber := flag.String("o", "", "The phone number that Amazon Pinpoint should use to send the voice message") // Should this be the same as sourceNumber???
	// configName := flag.String("c", "", "The name of the configuration set that you want to use to send the message")
	// text := flag.String("t", "", "The text of the message to send")

	verbose := flag.Bool("v", false, "Whether to barf out more info")
	flag.Parse()

	if *sourceNumber == "" || *destinationNumber == "" {
		fmt.Println("You must supply a source and destination telephone number")
		fmt.Println("(-s SOURCE -d DESTINATION")
		return
	}

	verbosePrint(*verbose, "Verbose enabled")

	// (import "github.com/aws/aws-sdk-go-v2/config")
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := pinpointsmsvoice.NewFromConfig(cfg)

	/*
		callInstructionsMessage := types.CallInstructionsMessageType{
			Text: aws.String("en-US"), // From Polly language topic https://docs.aws.amazon.com/polly/latest/dg/voicelist.html
		}

		plainTextMessage := types.PlainTextMessageType{
			LanguageCode: aws.String("en-US"), // Can this be different from the previous value?
			Text:         text,
			VoiceId:      aws.String("Joey"),
		}
	*/

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

	result, err := client.SendVoiceMessage(context.Background(), input)
	if err != nil {
		fmt.Println("Got an error calling SendVoiceMessage:")
		fmt.Println(err)
		return
	}

	fmt.Println("Message ID: " + *result.MessageId)

	fmt.Println("")
}
