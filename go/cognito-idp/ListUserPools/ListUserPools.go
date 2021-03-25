// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

func main() {
	verbose := flag.Bool("v", false, "Whether to show creation date and users")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Got an error creating default configuration")
		return
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	result, err := client.ListUserPools(
		context.TODO(),
		&cognitoidentityprovider.ListUserPoolsInput{
			MaxResults: 10,
		})
	if err != nil {
		fmt.Println("Got error listing user pools:")
		fmt.Println(err)
		return
	}

	fmt.Println("User pools:")
	fmt.Println("")

	for _, pool := range result.UserPools {
		fmt.Println("Name:    " + *pool.Name)
		fmt.Println("ID:      " + *pool.Id)
		desc, err := client.DescribeUserPool(
			context.TODO(),
			&cognitoidentityprovider.DescribeUserPoolInput{
				UserPoolId: pool.Id})
		if err != nil {
			fmt.Println("Got error describing pool:")
			fmt.Println(err)
			return
		}

		fmt.Println("ARN:     " + *desc.UserPool.Arn)

		if *verbose {
			fmt.Println("Created:", *pool.CreationDate)
			fmt.Println("Users:")

			rsp, err := client.ListUsers(context.TODO(),
				&cognitoidentityprovider.ListUsersInput{
					UserPoolId: *&pool.Id,
				})

			if err != nil {
				fmt.Println("Got an error listing users")
			} else {
				for _, user := range rsp.Users {
					fmt.Print("  " + *user.Username)
					for _, a := range user.Attributes {
						if *a.Name == "email" {
							fmt.Println(" (" + *a.Value + ")")
						}
					}
				}

				fmt.Println("")
			}
		}

		fmt.Println("")
	}

	fmt.Println("Found", len(result.UserPools), "pool(s)")
}
