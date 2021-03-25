// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// Creates Cognito user pool POOL
//
// Usage:
//    go run CreateUserPool.go -p POOL [-s SUBJECT]
func main() {
	poolName := flag.String("p", "", "The name of the pool")
	subject := flag.String("s", "Join my user pool", "The subject in email to users")
	flag.Parse()

	if *poolName == "" || *subject == "" {
		fmt.Println("You must supply a user pool name and email subject (-p POOL -s \"SUBJECT\")")
		return
	}

	emailMsg := "{username} {####}"
	emailSubject := *subject

	waitDays := int32(1)

	emailVerifyMsg := "{####}"
	emailVerifySub := *subject

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got an error creating default configuration")
		return
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	params := &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: poolName,
		AdminCreateUserConfig: &types.AdminCreateUserConfigType{
			AllowAdminCreateUserOnly: false, // false == users can sign themselves up
			InviteMessageTemplate: &types.MessageTemplateType{
				EmailMessage: &emailMsg,     // Welcome message to new users
				EmailSubject: &emailSubject, // Welcome subject to new users
			},
			UnusedAccountValidityDays: waitDays, // How many days to wait before rescinding offer
		},
		AutoVerifiedAttributes: []types.VerifiedAttributeType{ // Auto-verified means the user confirmed the SNS message
			"email", // Required; either email or phone_number
		},
		EmailVerificationMessage: &emailVerifyMsg,
		EmailVerificationSubject: &emailVerifySub,
		Policies: &types.UserPoolPolicyType{
			PasswordPolicy: &types.PasswordPolicyType{
				MinimumLength:    6, // Require a password of at least 6 chars
				RequireLowercase: false,
				RequireNumbers:   false,
				RequireSymbols:   false,
				RequireUppercase: false,
			},
		},
	}

	fmt.Println("")

	cgResp, cgErr := client.CreateUserPool(context.TODO(), params)
	if cgErr != nil {
		fmt.Println("Got error attempting to create user pool:")
		fmt.Println(err.Error())
		return
	}

	fmt.Println("")
	fmt.Println("ARN: " + *cgResp.UserPool.Arn)
}
