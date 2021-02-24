#!/usr/bin/env node
import 'source-map-support/register';

import * as cdk from '@aws-cdk/core';
import * as s3 from '@aws-cdk/aws-s3';
import * as cloudtrail from '@aws-cdk/aws-cloudtrail';
import * as cw from '@aws-cdk/aws-cloudwatch';
import * as events from '@aws-cdk/aws-events';
import * as iam from '@aws-cdk/aws-iam';
import * as lambda from '@aws-cdk/aws-lambda';
import * as logs from '@aws-cdk/aws-logs';
import * as dynamodb from '@aws-cdk/aws-dynamodb';
import * as sfn from '@aws-cdk/aws-stepfunctions';
import * as tasks from '@aws-cdk/aws-stepfunctions-tasks';
import { Effect } from '@aws-cdk/aws-iam';

export class ImageStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Prefixes for images in s3
    const inPrefix = "uploads"
    const thumbPrefix = "thumbs"

    /* Use bucket event to execute a step function when an item uploaded to a bucket
     *   https://docs.aws.amazon.com/step-functions/latest/dg/tutorial-cloudwatch-events-s3.html
     *
     * 1: Create a bucket (Amazon S3)
     * 2: Create a trail (AWS CloudTrail)
     * 3: Create an events rule (AWS CloudWatch Events)
     */

    // Create Amazon Simple Storage Service (Amazon S3) bucket
    const myBucket = new s3.Bucket(this, 'doc-example-bucket');

    // Create trail to watch for events from bucket
    const myTrail = new cloudtrail.Trail(this, 'doc-example-trail');
    // Add an event selector to the trail so that
    // JPG or PNG files with 'uploads/' prefix
    // added to bucket are detected
    myTrail.addS3EventSelector([{
      bucket: myBucket,
      objectPrefix: inPrefix + '/',
    },]);

    // Create events rule
    const myRule = new events.Rule(this, 'rule', {
      eventPattern: {
        source: ['aws.s3'],
      },
    });

    // Create DynamoDB table for Lambda function to persist image info
    // Create Amazon DynamoDB table with primary key path (string)
    // that will be something like uploads/myPhoto.jpg
    const myTable = new dynamodb.Table(this, 'doc-example-table', {
      partitionKey: { name: 'path', type: dynamodb.AttributeType.STRING },
      stream: dynamodb.StreamViewType.NEW_IMAGE,
    });

    /* 
     * Define Lambda functions to:
     * 1. Add metadata from the photo to a Dynamodb table.     
     * 2. Call Amazon Rekognition to detect objects in the image file.
     * 3. Generate a thumbnail and store it in the S3 bucket with the **resized/** prefix
     */

    // Lambda function that:
    // 1. Receives notifications from Amazon S3 (ItemUpload)
    // 2. Gets metadata from the photo
    // 3. Saves the metadata in a DynamoDB table
    const saveMetadataFunction = new lambda.Function(this, 'doc-example-save-metadata', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: new lambda.AssetCode('src/save_metadata'), // Go source file is (relative to cdk.json): src/save_metadata/main.go
      environment: {
        tableName: myTable.tableName,
      },
    });

    // Add policy to Lambda function so it can call
    // GetObject on bucket and PutItem on table.
    /*
    const s3Policy = new iam.PolicyStatement({
      sid: "docexamples3dbstatement",
      actions: ["s3:GetObject", "dynamodb:PutItem"],
      effect: Effect.ALLOW,
      resources: [myBucket.bucketArn + "/*", myTable.tableArn + "/*"],
    })

    saveMetadataFunction.role?.addToPrincipalPolicy(s3Policy)
    */

    // Give Lambda function, which save ELIF data, write access to DynamoDB table and read access to S3 bucket
    myTable.grantWriteData(saveMetadataFunction.grantPrincipal)
    myBucket.grantRead(saveMetadataFunction.grantPrincipal)

    // Lambda function that:
    // 1. Calls Amazon Rekognition to detect objects in the image file
    // 2. Saves information about the objects in a Dynamodb table
    const saveObjectDataFunction = new lambda.Function(this, 'doc-example-save-object-data', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: new lambda.AssetCode('src/save_objectdata'), // Go source file is (relative to cdk.json): src/save_objectdata/main.go
      environment: {
        tableName: myTable.tableName,
      },
    });

    // Add policy to Lambda function so it can call
    // UpdateItem on table.
    // Do we need to add rekognition.DetectLabels???
    /*
    const dbPolicy = new iam.PolicyStatement({
      sid: "docexampledbstatement",
      actions: ["dynamodb:UpdateItem"],
      effect: Effect.ALLOW,
      resources: [myTable.tableArn + "/*"],
    })

    saveObjectDataFunction.role?.addToPrincipalPolicy(dbPolicy)
    */

    // Give Lambda function, which saves Rekognition data, write access to DynamoDB table
    myTable.grantWriteData(saveObjectDataFunction.grantPrincipal)

    // Lambda function that:
    // 1. Gets the photo from S3
    // 2. Creates a thumbnail of the photo
    // 3. Save the photo back into S3
    const createThumbnailFunction = new lambda.Function(this, 'doc-example-create-thumbnail', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: new lambda.AssetCode('src/create_thumbnail'), // Go source file is (relative to cdk.json): src/create_thumbnail/main.go
    });

    // Add policy to Lambda function so it can call
    // GetObject and PutObject on bucket.
    /*
    const s32Policy = new iam.PolicyStatement({
      sid: "docexamples3statement",
      actions: ["s3:GetObject", "s3:PutObject"],
      effect: Effect.ALLOW,
      resources: [myBucket.bucketArn + "/*"],
    })

    createThumbnailFunction.role?.addToPrincipalPolicy(s32Policy)
    */

    // Give Lambda function, which creates a thumbnail, read/write access to S3 bucket
    myBucket.grantReadWrite(createThumbnailFunction.grantPrincipal)

    // First task: save metadata from photo in S3 bucket to DynamoDB table
    const saveMetadataJob = new tasks.LambdaInvoke(this, 'Save Metadata Job', {
      lambdaFunction: saveMetadataFunction,
      //inputPath: '$', // Event from S3 notification (default)
      outputPath: '$.Payload',
    });

    // Second task: save image data from Rekognition to DynamoDB table
    const saveObjectDataJob = new tasks.LambdaInvoke(this, 'Save Object Data Job', {
      lambdaFunction: saveObjectDataFunction,
      inputPath: '$.Payload',
      outputPath: '$.Payload',
    });

    // Final task: create thumbnail of photo in S3 bucket
    const createThumbnailJob = new tasks.LambdaInvoke(this, 'Create Thumbnail Job', {
      lambdaFunction: createThumbnailFunction,
      inputPath: '$.Payload',
      outputPath: '$.Payload',
    });

    // Create state machine with one task, submitJob
    const definition = saveMetadataJob
      .next(saveObjectDataJob)
      .next(createThumbnailJob)

    // Create log group:
    const myLogGroup = new logs.LogGroup(this, 'MyLogGroup');

    const myStateMachine = new sfn.StateMachine(this, 'StateMachine', {
      definition,
      logs: {
        destination: myLogGroup,
        level: sfn.LogLevel.ALL,
      },
      timeout: cdk.Duration.minutes(5),
    });

    // Create role for Lambda function to call DynamoDB
    /*
    const dynamoDbRole = new iam.Role(this, 'doc-example-dynamodb-role', {
      roleName: 'doc-example-dynamodb',
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com')
    });

    // Let Lambda call these DynamoDB operations
    dynamoDbRole.addToPolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      resources: [myTable.tableArn],
      actions: [
        'dynamodb:PutItem'
      ]
    }));
    */

    // Display info about the resources.
    // You can see this information at any time by running:
    //   aws cloudformation describe-stacks --stack-name ImageStack --query Stacks[0].Outputs --output text
    new cdk.CfnOutput(this, 'Bucket name: ', { value: myBucket.bucketName });

    new cdk.CfnOutput(this, 'Save metadata function: ', { value: saveMetadataFunction.functionName });
    new cdk.CfnOutput(this, 'Save object data function: ', { value: saveObjectDataFunction.functionName });
    new cdk.CfnOutput(this, 'Create thumbnail function: ', { value: createThumbnailFunction.functionName });

    new cdk.CfnOutput(this, 'CloudTrail trail ARN: ', { value: myTrail.trailArn });

    new cdk.CfnOutput(this, 'Cloudwatch log group: ', { value: myLogGroup.logGroupName })

    new cdk.CfnOutput(this, 'Save ELIF data function CloudWatch log group: ', { value: saveMetadataFunction.logGroup.logGroupName });
    new cdk.CfnOutput(this, 'Save Rekognition function CloudWatch log group: ', { value: saveObjectDataFunction.logGroup.logGroupName });
    new cdk.CfnOutput(this, 'Create thumbnail function CloudWatch log group: ', { value: createThumbnailFunction.logGroup.logGroupName });

    new cdk.CfnOutput(this, 'Table name: ', { value: myTable.tableName });
    new cdk.CfnOutput(this, 'State machine: ', { value: myStateMachine.stateMachineName });
  }
}

const app = new cdk.App();
new ImageStack(app, 'ImageStack');
