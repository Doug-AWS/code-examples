/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0.
 */

use lambda_runtime::{handler_fn, Context, Error};
use log::LevelFilter;
use serde::{Deserialize, Serialize};
use simple_logger::SimpleLogger;

#[derive(Deserialize)]
struct Request {
    command: String,
}

/// A typical response structure.
/// The runtime serializes responses into JSON.
#[derive(Serialize)]
struct Response {
    req_id: String,
    msg: String,
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    // Enable CloudWatch error logging by the runtime.
    // You can replace it with any other method of initializing `log`.
    SimpleLogger::new()
        .with_level(LevelFilter::Info)
        .init()
        .unwrap();

    let func = handler_fn(my_handler);
    lambda_runtime::run(func).await?;
    Ok(())
}

async fn my_handler(event: Request, ctx: Context) -> Result<Response, Error> {
    // For now we are just displaying the event as a CloudWatch log entry.
    let resp = Response {
        req_id: ctx.request_id,
        msg: format!("Event: {}", event.command),
    };

    // Return `Response`, which will be serialized to JSON automatically by the runtime.
    Ok(resp)
}
