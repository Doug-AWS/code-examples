/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0.
 */

#[derive(Serialize, Deserialize)]
struct message {
    S: String,
}

#[derive(Serialize, Deserialize)]
struct username {
    S: String,
}

#[derive(Serialize, Deserialize)]
struct timestamp {
    S: String,
}

#[derive(Serialize, Deserialize)]
struct newimage {
    Timestamp: timestamp,
    Message: message,
    Username: username,
}

#[derive(Serialize, Deserialize)]
struct keys {
    Timestamp: timestamp,
    Username: username,
}

#[derive(Serialize, Deserialize)]
struct dynamoDB {
    ApproximateCreationDateTime: String,
    Keys: keys,
    NewImage: newimage,
    SequenceNumber: String,
    SizeBytes: u8,
    StreamViewType: String,
}

#[derive(Serialize, Deserialize)]
struct Record {
    eventID: String,
    eventName: String,
    eventSource: String,
    awsRegion: string,
    dynamodb: dynamoDB,
    eventSourceARN: String,
}

/// Displays a DynamoDB event.
#[tokio::main]
async fn main() -> Result<(), Error> {
    tracing_subscriber::fmt::init();

    let Opt { event, verbose } = Opt::from_args();

    println!();

    if verbose {
        println!("DynamoDB client version: {}", PKG_VERSION);
        println!("Event filename:          {}", &event);
        println!();
    }

    let data = r#"
    {
        "eventID": "7de3041dd709b024af6f29e4fa13d34c",
        "eventName": "INSERT",
        "eventVersion": "1.1",
        "eventSource": "aws:dynamodb",
        "awsRegion": "region",
        "dynamodb": {
            "ApproximateCreationDateTime": 1479499740,
            "Keys": {
                "Timestamp": {
                    "S": "2016-11-18:12:09:36"
                },
                "Username": {
                    "S": "John Doe"
                }
            },
            "NewImage": {
                "Timestamp": {
                    "S": "2016-11-18:12:09:36"
                },
                "Message": {
                    "S": "This is a bark from the Woofer social network"
                },
                "Username": {
                    "S": "John Doe"
                }
            },
            "SequenceNumber": "13021600000000001596893679",
            "SizeBytes": 112,
            "StreamViewType": "NEW_IMAGE"
        },
        "eventSourceARN": "arn:aws:dynamodb:region:123456789012:table/BarkTable/stream/2016-11-16T20:42:48.104"
    }"#;

    println!("Event info:");

    let r: Record = serde_json::from_str(data)?;

    println!(
        "The {} event was an {} from the {} region with the following message from {} at {}",
        r.eventSource,
        r.eventName,
        r.awsRegion,
        r.dynamodb.NewImage.Username,
        r.dynamodb.NewImage.Timestamp
    );
    println!("{}", r.dynamodb.NewImage.Message);

    Ok(())
}
