// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
package main

import (
        "errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// Define a mock struct to use in unit tests
type mockSqsClient struct {
	sqsiface.SQSAPI
}

func (m *mockSqsClient) ListQueues(input *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
//    resp := sqs.ListQueuesOutput{}
    _ = sqs.ListQueuesOutput()
//    return &resp, nil
    return nil, errors.New("error")
}

func TestListQueues60(t *testing.T) {
	thisTime := time.Now()
	nowString := thisTime.Format("20060102150405")
	t.Log("Starting unit test at " + nowString)

	mockSvc := &mockSqsClient{}

	_, err := GetQueues(mockSvc)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Retrieved queues")
}
