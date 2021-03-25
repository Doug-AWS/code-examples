// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

func main() {
	email := flag.String("e", "", "The email address of the user")
	userPoolId := flag.String("p", "", "The ID of the user pool")
	userName := flag.String("n", "", "The name of the user (optional)")
	confirm := flag.Bool("c", false, "Confirm user's email address")

	flag.Parse()

	if *email == "" || *userPoolId == "" {
		fmt.Println("You must supply an email address and user pool ID")
		fmt.Println("Usage: go run CreateUser.go -e EMAIL-ADDRESS -p USER-POOL-ID")
		return
	}

	// If userName is empty, just get it from the first part of the email address
	parts := strings.Split(*email, "@")

	if len(parts) != 2 {
		fmt.Println(*email + " is not a valid email address format")
		return
	}

	if *userName == "" {
		*userName = parts[0]
		fmt.Println("User name is " + *userName)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got an error creating default configuration")
		return
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	if *confirm {
		newUserData := &cognitoidentityprovider.AdminCreateUserInput{
			MessageAction: types.MessageActionTypeResend,
			UserPoolId:    userPoolId,
			Username:      userName,
			DesiredDeliveryMediums: []types.DeliveryMediumType{
				"EMAIL",
			},
			UserAttributes: []types.AttributeType{
				{
					Name:  aws.String("email"),
					Value: email,
				},
			},
		}

		_, err = client.AdminCreateUser(context.TODO(), newUserData)
		if err != nil {
			fmt.Println("Got error confirming user's email address:")
			fmt.Println(err.Error())
			return
		}

		fmt.Println("Confirmed user's email address.")
	} else {
		newUserData := &cognitoidentityprovider.AdminCreateUserInput{
			UserPoolId: userPoolId,
			Username:   userName,
			DesiredDeliveryMediums: []types.DeliveryMediumType{
				"EMAIL",
			},
			UserAttributes: []types.AttributeType{
				{
					Name:  aws.String("email"),
					Value: email,
				},
			},
		}

		_, err = client.AdminCreateUser(context.TODO(), newUserData)
		if err != nil {
			fmt.Println("Got error creating user:")
			fmt.Println(err.Error())
			return
		}

		fmt.Println("Created user.")
	}
}
