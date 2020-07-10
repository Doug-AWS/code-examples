package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"    
)

func debugPrint(debug bool, s string) {
    if debug {
        fmt.Println(s)
    }
}

func testPrint(test bool, s string) {
    if test {
        fmt.Println(s)
    }
}

func addAccountToTable(region *string, table string, account string, policyID string) {
    s3Policy := map[string]interface{}{
        "Version": "2012-10-17",
        "Id":      policyID,
        "Statement": []map[string]interface{}{
            {
                "Sid":      "ListAndDescribe",
                "Action":   [
                    "dynamodb:List*",
                    "dynamodb:DescribeReservedCapacity*",
                    "dynamodb:DescribeLimits",
                    "dynamodb:DescribeTimeToLive"
                ],
                "Resource": "*",
            },
            {
                "Sid":      "SpecificTable",
                "Action":   [
                    "dynamodb:BatchGet*",
                    "dynamodb:DescribeStream",
                    "dynamodb:DescribeTable",
                    "dynamodb:Get*",
                    "dynamodb:Query",
                    "dynamodb:Scan",
                    "dynamodb:BatchWrite*",
                    "dynamodb:CreateTable",
                    "dynamodb:Delete*",
                    "dynamodb:Update*",
                    "dynamodb:PutItem"
                ],
                "Effect":   "Allow",
                "Resource": "arn:aws:dynamodb:*:*:table/" + table
                },
            },
        },
    }

    policy, err := json.Marshal(s3Policy)
    if err != nil {
        fmt.Println("Got error marshalling policy JSON:")
        fmt.Println(err.Error())
        os.Exit(1)
    }

    sess := session.Must(session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    }))

    svc := dynamodb.New(sess)

    // Add policy to table
    input := &dynamodb.Pu
    input := &s3.PutBucketPolicyInput{
        Bucket: aws.String(bucket),
        Policy: aws.String(string(policy)),
    }

    _, err = svc.PutBucketPolicy(input)
    if err != nil {
        fmt.Println("Error adding policy for account:", account)
        fmt.Println(err.Error())
    }
}

func addAccountsToBucket(region *string, bucket string, accountList []string, policyID string) {
    s3Policy := map[string]interface{}{
        "Version": "2012-10-17",
        "Id":      policyID,
        "Statement": []map[string]interface{}{
            {
                "Sid":      "AllowBucketAccess",
                "Action":   "s3:ListBucket",
                "Effect":   "Allow",
                "Resource": "arn:aws:s3:::" + bucket,
                "Principal": map[string]interface{}{
                    "AWS": accountList,
                },
            },
            {
                "Sid":      "AllowObjectAccess",
                "Action":   "s3:GetObject",
                "Effect":   "Allow",
                "Resource": "arn:aws:s3:::" + bucket + "/*",
                "Principal": map[string]interface{}{
                    "AWS": accountList,
                },
            },
        },
    }

    policy, err := json.Marshal(s3Policy)
    if err != nil {
        fmt.Println("Got error marshalling policy JSON:")
        fmt.Println(err.Error())
        os.Exit(1)
    }

    sess, err := session.NewSession(&aws.Config{
        Region: region},
    )

    svc := s3.New(sess)

    // Add policy to bucket
    input := &s3.PutBucketPolicyInput{
        Bucket: aws.String(bucket),
        Policy: aws.String(string(policy)),
    }

    _, err = svc.PutBucketPolicy(input)
    if err != nil {
        fmt.Println("Error adding policy to bucket:")
        fmt.Println(err.Error())
        os.Exit(1)
    }
}

func usage() {
    fmt.Println("Usage:")
    fmt.Println("go run S3AddBucketAccessPolicy -b BUCKET_NAME [-f ACCOUNT_FILE] [-p POLICY_ID] [-r REGION] [-t] [-h] [-d]")
    fmt.Println("")
    fmt.Println("BUCKET_NAME is required")
    fmt.Println("")
    fmt.Println("ACCOUNT_FILE is a file with a list of account IDs, one per line")
    fmt.Println("If omitted, defaults to 'WhiteList.txt'")
    fmt.Println("")
    fmt.Println("POLICY_ID is the value for the policy ID")
    fmt.Println("If omitted, defaults to 'beta'")
    fmt.Println("")
    fmt.Println("REGION is the region in which the bucket was created")
    fmt.Println("If omitted, defaults to 'us-west-2'")
    fmt.Println("")
    fmt.Println("-t runs the app in test mode, trying each account ID in turn, to check the validity")
    fmt.Println("   THIS RESULTS IN ONLY THE LAST ACCOUNT ID IN THE WHITELIST HAVING ACCESS!!!")
    fmt.Println("")
    fmt.Println("   The most effective way to use this feature is to batch up new account IDs")
    fmt.Println("   into a separate whitelist and then run -t against that list.")
    fmt.Println("   Once you've vetted the list, append it to the master list")
    fmt.Println("   and run this app, without -t, against the master list.")
    fmt.Println("")
    fmt.Println("-h prints this message and quits")
    fmt.Println("")
    fmt.Println("-d prints extra (debugging) information")
    fmt.Println("")
}

func main() {
    bucketPtr := flag.String("b", "", "")
    accountPtr := flag.String("f", "WhiteList.txt", "")
    policyIDPtr := flag.String("p", "beta", "")
    regionPtr := flag.String("r", "us-west-2", "")
    testPtr := flag.Bool("t", false, "")
    helpPtr := flag.Bool("h", false, "")
    debugPtr := flag.Bool("d", false, "")
    flag.Parse()
    bucket := *bucketPtr
    accountFile := *accountPtr
    policyID := *policyIDPtr
    test := *testPtr
    help := *helpPtr
    debug := *debugPtr

    if help {
        usage()
        os.Exit(0)
    }

    // Validate args
    if bucket == "" {
        fmt.Println("You must supply the name of the bucket")
        usage()
        os.Exit(1)
    }

    testPrint(test, "Running in test mode")

    file, err := os.Open(accountFile)
    if err != nil {
        fmt.Println("Got error opening file:")
        fmt.Println(err.Error())
        os.Exit(1)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    // We print out each successful account ID as we go
    // We know when we hit a bad account if the app quits BEFORE printing it out
    testPrint(test, "Account list:")

    numAccts := 0

    // Create list of accounts
    accountList := []string{}

    for scanner.Scan() {
        // Add account to list
        line := scanner.Text()

        // If the line is < 12 chars, prepend 0s until it's 12 chars
        if len(line) < 12 {
            debugPrint(debug, "Account "+line+" is < 12 chars")
        }

        for len(line) < 12 {
            line = "0" + line
        }

        if test {
            // So we can validate each account ID
            addAccountToBucket(regionPtr, bucket, line, policyID)
            testPrint(test, "  "+line)
        } else {
            accountList = append(accountList, line)
            debugPrint(debug, "  "+line)
        }

        numAccts++
    }
    if err := scanner.Err(); err != nil {
        fmt.Println("Got error reading file:")
        fmt.Println(err.Error())
        os.Exit(1)
    }

    if test {
        fmt.Println("Only one account was given access!!!")
    } else {
        addAccountsToBucket(regionPtr, bucket, accountList, policyID)
        fmt.Println("Added", numAccts, "accounts to bucket:", bucket)
    }
}
