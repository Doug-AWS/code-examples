/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0.
 */

use serde::{Deserialize, Serialize};
use std::fs;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
struct Opt {
    /// The name of the JSON file containing the events.
    #[structopt(short, long)]
    json_file: String,

    /// Whether to display additional information.
    #[structopt(short, long)]
    verbose: bool,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Root {
    #[serde(rename = "Records")]
    pub records: Vec<Record>,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Record {
    #[serde(rename = "eventID")]
    pub event_id: String,
    pub event_name: String,
    pub event_version: String,
    pub event_source: String,
    pub aws_region: String,
    pub dynamodb: Dynamodb,
    #[serde(rename = "eventSourceARN")]
    pub event_source_arn: String,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Dynamodb {
    #[serde(rename = "ApproximateCreationDateTime")]
    pub approximate_creation_date_time: i64,
    #[serde(rename = "Keys")]
    pub keys: Keys,
    #[serde(rename = "NewImage")]
    pub new_image: NewImage,
    #[serde(rename = "SequenceNumber")]
    pub sequence_number: String,
    #[serde(rename = "SizeBytes")]
    pub size_bytes: i64,
    #[serde(rename = "StreamViewType")]
    pub stream_view_type: String,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Keys {
    #[serde(rename = "Timestamp")]
    pub timestamp: Timestamp,
    #[serde(rename = "Username")]
    pub username: Username,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Timestamp {
    #[serde(rename = "S")]
    pub s: String,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Username {
    #[serde(rename = "S")]
    pub s: String,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct NewImage {
    #[serde(rename = "Timestamp")]
    pub timestamp: Timestamp,
    #[serde(rename = "Message")]
    pub message: Message,
    #[serde(rename = "Username")]
    pub username: Username,
}

#[derive(Default, Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Message {
    #[serde(rename = "S")]
    pub s: String,
}

/// Displays a DynamoDB event.
fn main() {
    let Opt { json_file, verbose } = Opt::from_args();

    if verbose {
        println!("JSON filename: {}", &json_file);
        println!();
    }

    let contents = fs::read_to_string(json_file).expect("Something went wrong reading the file");

    println!("Event info:");

    let data: &str = &(*contents);

    let root: Root = serde_json::from_str(data).unwrap();

    for r in root.records {
        println!(
            "The {} event was an {} from the {} region with the following message from {} at {}:",
            r.event_source,
            r.event_name,
            r.aws_region,
            r.dynamodb.new_image.username.s,
            r.dynamodb.new_image.timestamp.s
        );
        println!("{}", r.dynamodb.new_image.message.s);
    }
}
