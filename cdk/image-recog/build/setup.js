#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ImageRecogStack = void 0;
require("source-map-support/register");
const cdk = require("@aws-cdk/core");
// import { Duration } from "@aws-cdk/core";
// import * as codebuild from '@aws-cdk/aws-codebuild';
// import * as amplify from '@aws-cdk/aws-amplify';
const s3 = require("@aws-cdk/aws-s3");
const cloudtrail = require("@aws-cdk/aws-cloudtrail");
const events = require("@aws-cdk/aws-events");
// import * as cloudwatch from '@aws-cdk/aws-cloudwatch';
// import * as nots from '@aws-cdk/aws-s3-notifications';
const iam = require("@aws-cdk/aws-iam");
const lambda = require("@aws-cdk/aws-lambda");
const dynamodb = require("@aws-cdk/aws-dynamodb");
const sfn = require("@aws-cdk/aws-stepfunctions");
const tasks = require("@aws-cdk/aws-stepfunctions-tasks");
class ImageRecogStack extends cdk.Stack {
    constructor(scope, id, props) {
        super(scope, id, props);
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
                objectPrefix: 'uploads/',
            },]);
        // Create events rule
        const rule = new events.Rule(this, 'rule', {
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
            code: new lambda.AssetCode('src/save_metadata'),
            environment: {
                tableName: myTable.tableName,
            },
        });
        // Lambda function that:
        // 1. Calls Amazon Rekognition to detect objects in the image file
        // 2. Saves information about the objects in a Dynamodb table
        const saveObjectDataFunction = new lambda.Function(this, 'doc-example-save-object-data', {
            runtime: lambda.Runtime.GO_1_X,
            handler: 'main',
            code: new lambda.AssetCode('src/save_objectdata'),
            environment: {
                tableName: myTable.tableName,
            },
        });
        // Lambda function that:
        // 1. Gets the photo from S3
        // 2. Creates a thumbnail of the photo
        // 3. Save the photo back into S3
        const createThumbnailFunction = new lambda.Function(this, 'doc-example-create-thumbnail', {
            runtime: lambda.Runtime.GO_1_X,
            handler: 'main',
            code: new lambda.AssetCode('src/create_thumbnail'),
        });
        // Create Lambda function to get status of uploaded data for state machine
        /*
        const getStatusLambda = new lambda.Function(this, 'doc-example-get-status', {
          runtime: lambda.Runtime.GO_1_X,
          handler: 'main',
          code: new lambda.AssetCode('src/get_status'), // Go source file is (relative to cdk.json): src/get_status/main.go
          environment: {
            tableName: myTable.tableName,
          },
        });
        */
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
        /*
        const waitX = new sfn.Wait(this, 'Wait X Seconds', {
          time: sfn.WaitTime.duration(cdk.Duration.seconds(5))       //.secondsPath('$.Payload.waitSeconds'),
        });
    
        const getStatus = new tasks.LambdaInvoke(this, 'Get Job Status', {
          lambdaFunction: getStatusLambda,
          inputPath: '$.guid',
          outputPath: '$.status',
        });
    
        const jobFailed = new sfn.Fail(this, 'Job Failed', {
          cause: 'AWS Batch Job Failed',
          error: 'DescribeJob returned FAILED',
        });
    
        const finalStatus = new tasks.LambdaInvoke(this, 'Get Final Job Status', {
          lambdaFunction: getStatusLambda,
          inputPath: '$.guid',
          outputPath: '$.status',
        });
        */
        // Create state machine with one task, submitJob
        const definition = saveMetadataJob
            .next(saveObjectDataJob)
            .next(createThumbnailJob);
        //      .next(waitX)
        //      .next(getStatus)
        //      .next(new sfn.Choice(this, 'Job Complete?')
        // Look at the "status" field
        //        .when(sfn.Condition.stringEquals('$.status', 'FAILED'), jobFailed)
        //        .when(sfn.Condition.stringEquals('$.status', 'SUCCEEDED'), finalStatus)
        //        .otherwise(waitX));
        const myStateMachine = new sfn.StateMachine(this, 'StateMachine', {
            definition,
            timeout: cdk.Duration.minutes(5),
        });
        // Send S3 events to Step Functions state machine
        //   rule.addTarget(new targets.SfnStateMachine(myStateMachine));
        /*
        // Create role for Lambda function to call Step Functions
        const stepFuncRole = new iam.Role(this, 'doc-example-stepfunc-role', {
          roleName: 'doc-example-stepfunc',
          assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com')
        });
    
        // Let Lambda function call these Step Function operations
        stepFuncRole.addToPolicy(new iam.PolicyStatement({
          actions: ["???", ""],
          effect: iam.Effect.ALLOW,
          resources: [],
        }))
        */
        // Create role for Lambda function to call DynamoDB
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
        /*
        dynamoDbRole.addManagedPolicy(
          iam.ManagedPolicy.fromAwsManagedPolicyName(
            'service-role/AWSLambdaBasicExecutionRole'));
        */
        // Configure Amazon S3 bucket to send notification events to step functions.
        // myBucket.addEventNotification(s3.EventType.OBJECT_CREATED, new nots.LambdaDestination(getMetadataFunction));
        // Display info about the resources.
        // You can see this information at any time by running:
        //   aws cloudformation describe-stacks --stack-name ImageRecogStack --query Stacks[0].Outputs --output text
        new cdk.CfnOutput(this, 'Bucket name: ', { value: myBucket.bucketName });
        new cdk.CfnOutput(this, 'S3 function name: ', { value: saveMetadataFunction.functionName });
        new cdk.CfnOutput(this, 'S3 function CloudWatch log group: ', { value: saveMetadataFunction.logGroup.logGroupName });
        // new cdk.CfnOutput(this, 'Status function: ', { value: getStatusLambda.functionName });
        new cdk.CfnOutput(this, 'Table name: ', { value: myTable.tableName });
        new cdk.CfnOutput(this, 'State machine: ', { value: myStateMachine.stateMachineName });
    }
}
exports.ImageRecogStack = ImageRecogStack;
const app = new cdk.App();
new ImageRecogStack(app, 'ImageRecogStack');
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoic2V0dXAuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi9zZXR1cC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7O0FBQ0EsdUNBQXFDO0FBRXJDLHFDQUFxQztBQUNyQyw0Q0FBNEM7QUFDNUMsdURBQXVEO0FBQ3ZELG1EQUFtRDtBQUNuRCxzQ0FBc0M7QUFDdEMsc0RBQXNEO0FBQ3RELDhDQUE4QztBQUU5Qyx5REFBeUQ7QUFFekQseURBQXlEO0FBQ3pELHdDQUF3QztBQUN4Qyw4Q0FBOEM7QUFDOUMsa0RBQWtEO0FBQ2xELGtEQUFrRDtBQUVsRCwwREFBMEQ7QUFFMUQsTUFBYSxlQUFnQixTQUFRLEdBQUcsQ0FBQyxLQUFLO0lBQzVDLFlBQVksS0FBb0IsRUFBRSxFQUFVLEVBQUUsS0FBc0I7UUFDbEUsS0FBSyxDQUFDLEtBQUssRUFBRSxFQUFFLEVBQUUsS0FBSyxDQUFDLENBQUM7UUFFeEI7Ozs7OztXQU1HO1FBRUgsMERBQTBEO1FBQzFELE1BQU0sUUFBUSxHQUFHLElBQUksRUFBRSxDQUFDLE1BQU0sQ0FBQyxJQUFJLEVBQUUsb0JBQW9CLENBQUMsQ0FBQztRQUUzRCwrQ0FBK0M7UUFDL0MsTUFBTSxPQUFPLEdBQUcsSUFBSSxVQUFVLENBQUMsS0FBSyxDQUFDLElBQUksRUFBRSxtQkFBbUIsQ0FBQyxDQUFDO1FBQ2hFLDZDQUE2QztRQUM3QywwQ0FBMEM7UUFDMUMsK0JBQStCO1FBQy9CLE9BQU8sQ0FBQyxrQkFBa0IsQ0FBQyxDQUFDO2dCQUMxQixNQUFNLEVBQUUsUUFBUTtnQkFDaEIsWUFBWSxFQUFFLFVBQVU7YUFDekIsRUFBRSxDQUFDLENBQUM7UUFFTCxxQkFBcUI7UUFDckIsTUFBTSxJQUFJLEdBQUcsSUFBSSxNQUFNLENBQUMsSUFBSSxDQUFDLElBQUksRUFBRSxNQUFNLEVBQUU7WUFDekMsWUFBWSxFQUFFO2dCQUNaLE1BQU0sRUFBRSxDQUFDLFFBQVEsQ0FBQzthQUNuQjtTQUNGLENBQUMsQ0FBQztRQUVILGtFQUFr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SSxFQUFFLG9DQUFvQyxFQUFFLEVBQUUsS0FBSyxFQUFFLG9CQUFvQixDQUFDLFFBQVEsQ0FBQyxZQUFZLEVBQUUsQ0FBQyxDQUFDO1FBRXJILHlGQUF5RjtRQUN6RixJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLGNBQWMsRUFBRSxFQUFFLEtBQUssRUFBRSxPQUFPLENBQUMsU0FBUyxFQUFFLENBQUMsQ0FBQztRQUV0RSxJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLGlCQUFpQixFQUFFLEVBQUUsS0FBSyxFQUFFLGNBQWMsQ0FBQyxnQkFBZ0IsRUFBRSxDQUFDLENBQUM7SUFDekYsQ0FBQztDQUNGO0FBcE5ELDBDQW9OQztBQUVELE1BQU0sR0FBRyxHQUFHLElBQUksR0FBRyxDQUFDLEdBQUcsRUFBRSxDQUFDO0FBQzFCLElBQUksZUFBZSxDQUFDLEdBQUcsRUFBRSxpQkFBaUIsQ0FBQyxDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiIyEvdXNyL2Jpbi9lbnYgbm9kZVxyXG5pbXBvcnQgJ3NvdXJjZS1tYXAtc3VwcG9ydC9yZWdpc3Rlcic7XHJcblxyXG5pbXBvcnQgKiBhcyBjZGsgZnJvbSAnQGF3cy1jZGsvY29yZSc7XHJcbi8vIGltcG9ydCB7IER1cmF0aW9uIH0gZnJvbSBcIkBhd3MtY2RrL2NvcmVcIjtcclxuLy8gaW1wb3J0ICogYXMgY29kZWJ1aWxkIGZyb20gJ0Bhd3MtY2RrL2F3cy1jb2RlYnVpbGQnO1xyXG4vLyBpbXBvcnQgKiBhcyBhbXBsaWZ5IGZyb20gJ0Bhd3MtY2RrL2F3cy1hbXBsaWZ5JztcclxuaW1wb3J0ICogYXMgczMgZnJvbSAnQGF3cy1jZGsvYXdzLXMzJztcclxuaW1wb3J0ICogYXMgY2xvdWR0cmFpbCBmcm9tICdAYXdzLWNkay9hd3MtY2xvdWR0cmFpbCc7XHJcbmltcG9ydCAqIGFzIGV2ZW50cyBmcm9tICdAYXdzLWNkay9hd3MtZXZlbnRzJztcclxuaW1wb3J0ICogYXMgdGFyZ2V0cyBmcm9tICdAYXdzLWNkay9hd3MtZXZlbnRzLXRhcmdldHMnO1xyXG4vLyBpbXBvcnQgKiBhcyBjbG91ZHdhdGNoIGZyb20gJ0Bhd3MtY2RrL2F3cy1jbG91ZHdhdGNoJztcclxuXHJcbi8vIGltcG9ydCAqIGFzIG5vdHMgZnJvbSAnQGF3cy1jZGsvYXdzLXMzLW5vdGlmaWNhdGlvbnMnO1xyXG5pbXBvcnQgKiBhcyBpYW0gZnJvbSAnQGF3cy1jZGsvYXdzLWlhbSc7XHJcbmltcG9ydCAqIGFzIGxhbWJkYSBmcm9tICdAYXdzLWNkay9hd3MtbGFtYmRhJztcclxuaW1wb3J0ICogYXMgZHluYW1vZGIgZnJvbSAnQGF3cy1jZGsvYXdzLWR5bmFtb2RiJztcclxuaW1wb3J0ICogYXMgc2ZuIGZyb20gJ0Bhd3MtY2RrL2F3cy1zdGVwZnVuY3Rpb25zJztcclxuaW1wb3J0IHsgV2FpdFRpbWUgfSBmcm9tIFwiQGF3cy1jZGsvYXdzLXN0ZXBmdW5jdGlvbnNcIjtcclxuaW1wb3J0ICogYXMgdGFza3MgZnJvbSAnQGF3cy1jZGsvYXdzLXN0ZXBmdW5jdGlvbnMtdGFza3MnO1xyXG5cclxuZXhwb3J0IGNsYXNzIEltYWdlUmVjb2dTdGFjayBleHRlbmRzIGNkay5TdGFjayB7XHJcbiAgY29uc3RydWN0b3Ioc2NvcGU6IGNkay5Db25zdHJ1Y3QsIGlkOiBzdHJpbmcsIHByb3BzPzogY2RrLlN0YWNrUHJvcHMpIHtcclxuICAgIHN1cGVyKHNjb3BlLCBpZCwgcHJvcHMpO1xyXG5cclxuICAgIC8qIFVzZSBidWNrZXQgZXZlbnQgdG8gZXhlY3V0ZSBhIHN0ZXAgZnVuY3Rpb24gd2hlbiBhbiBpdGVtIHVwbG9hZGVkIHRvIGEgYnVja2V0XHJcbiAgICAgKiAgIGh0dHBzOi8vZG9jcy5hd3MuYW1hem9uLmNvbS9zdGVwLWZ1bmN0aW9ucy9sYXRlc3QvZGcvdHV0b3JpYWwtY2xvdWR3YXRjaC1ldmVudHMtczMuaHRtbFxyXG4gICAgICpcclxuICAgICAqIDE6IENyZWF0ZSBhIGJ1Y2tldCAoQW1hem9uIFMzKVxyXG4gICAgICogMjogQ3JlYXRlIGEgdHJhaWwgKEFXUyBDbG91ZFRyYWlsKVxyXG4gICAgICogMzogQ3JlYXRlIGFuIGV2ZW50cyBydWxlIChBV1MgQ2xvdWRXYXRjaCBFdmVudHMpXHJcbiAgICAgKi9cclxuXHJcbiAgICAvLyBDcmVhdGUgQW1hem9uIFNpbXBsZSBTdG9yYWdlIFNlcnZpY2UgKEFtYXpvbiBTMykgYnVja2V0XHJcbiAgICBjb25zdCBteUJ1Y2tldCA9IG5ldyBzMy5CdWNrZXQodGhpcywgJ2RvYy1leGFtcGxlLWJ1Y2tldCcpO1xyXG5cclxuICAgIC8vIENyZWF0ZSB0cmFpbCB0byB3YXRjaCBmb3IgZXZlbnRzIGZyb20gYnVja2V0XHJcbiAgICBjb25zdCBteVRyYWlsID0gbmV3IGNsb3VkdHJhaWwuVHJhaWwodGhpcywgJ2RvYy1leGFtcGxlLXRyYWlsJyk7XHJcbiAgICAvLyBBZGQgYW4gZXZlbnQgc2VsZWN0b3IgdG8gdGhlIHRyYWlsIHNvIHRoYXRcclxuICAgIC8vIEpQRyBvciBQTkcgZmlsZXMgd2l0aCAndXBsb2Fkcy8nIHByZWZpeFxyXG4gICAgLy8gYWRkZWQgdG8gYnVja2V0IGFyZSBkZXRlY3RlZFxyXG4gICAgbXlUcmFpbC5hZGRTM0V2ZW50U2VsZWN0b3IoW3tcclxuICAgICAgYnVja2V0OiBteUJ1Y2tldCxcclxuICAgICAgb2JqZWN0UHJlZml4OiAndXBsb2Fkcy8nLFxyXG4gICAgfSxdKTtcclxuXHJcbiAgICAvLyBDcmVhdGUgZXZlbnRzIHJ1bGVcclxuICAgIGNvbnN0IHJ1bGUgPSBuZXcgZXZlbnRzLlJ1bGUodGhpcywgJ3J1bGUnLCB7XHJcbiAgICAgIGV2ZW50UGF0dGVybjoge1xyXG4gICAgICAgIHNvdXJjZTogWydhd3MuczMnXSxcclxuICAgICAgfSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIENyZWF0ZSBEeW5hbW9EQiB0YWJsZSBmb3IgTGFtYmRhIGZ1bmN0aW9uIHRvIHBlcnNpc3QgaW1hZ2UgaW5mb1xyXG4gICAgLy8gQ3JlYXRlIEFtYXpvbiBEeW5hbW9EQiB0YWJsZSB3aXRoIHByaW1hcnkga2V5IHBhdGggKHN0cmluZylcclxuICAgIC8vIHRoYXQgd2lsbCBiZSBzb21ldGhpbmcgbGlrZSB1cGxvYWRzL215UGhvdG8uanBnXHJcbiAgICBjb25zdCBteVRhYmxlID0gbmV3IGR5bmFtb2RiLlRhYmxlKHRoaXMsICdkb2MtZXhhbXBsZS10YWJsZScsIHtcclxuICAgICAgcGFydGl0aW9uS2V5OiB7IG5hbWU6ICdwYXRoJywgdHlwZTogZHluYW1vZGIuQXR0cmlidXRlVHlwZS5TVFJJTkcgfSxcclxuICAgICAgc3RyZWFtOiBkeW5hbW9kYi5TdHJlYW1WaWV3VHlwZS5ORVdfSU1BR0UsXHJcbiAgICB9KTtcclxuXHJcbiAgICAvKiBcclxuICAgICAqIERlZmluZSBMYW1iZGEgZnVuY3Rpb25zIHRvOlxyXG4gICAgICogMS4gQWRkIG1ldGFkYXRhIGZyb20gdGhlIHBob3RvIHRvIGEgRHluYW1vZGIgdGFibGUuICAgICBcclxuICAgICAqIDIuIENhbGwgQW1hem9uIFJla29nbml0aW9uIHRvIGRldGVjdCBvYmplY3RzIGluIHRoZSBpbWFnZSBmaWxlLlxyXG4gICAgICogMy4gR2VuZXJhdGUgYSB0aHVtYm5haWwgYW5kIHN0b3JlIGl0IGluIHRoZSBTMyBidWNrZXQgd2l0aCB0aGUgKipyZXNpemVkLyoqIHByZWZpeFxyXG4gICAgICovXHJcblxyXG4gICAgLy8gTGFtYmRhIGZ1bmN0aW9uIHRoYXQ6XHJcbiAgICAvLyAxLiBSZWNlaXZlcyBub3RpZmljYXRpb25zIGZyb20gQW1hem9uIFMzIChJdGVtVXBsb2FkKVxyXG4gICAgLy8gMi4gR2V0cyBtZXRhZGF0YSBmcm9tIHRoZSBwaG90b1xyXG4gICAgLy8gMy4gU2F2ZXMgdGhlIG1ldGFkYXRhIGluIGEgRHluYW1vREIgdGFibGVcclxuICAgIGNvbnN0IHNhdmVNZXRhZGF0YUZ1bmN0aW9uID0gbmV3IGxhbWJkYS5GdW5jdGlvbih0aGlzLCAnZG9jLWV4YW1wbGUtc2F2ZS1tZXRhZGF0YScsIHtcclxuICAgICAgcnVudGltZTogbGFtYmRhLlJ1bnRpbWUuR09fMV9YLFxyXG4gICAgICBoYW5kbGVyOiAnbWFpbicsXHJcbiAgICAgIGNvZGU6IG5ldyBsYW1iZGEuQXNzZXRDb2RlKCdzcmMvc2F2ZV9tZXRhZGF0YScpLCAvLyBHbyBzb3VyY2UgZmlsZSBpcyAocmVsYXRpdmUgdG8gY2RrLmpzb24pOiBzcmMvc2F2ZV9tZXRhZGF0YS9tYWluLmdvXHJcbiAgICAgIGVudmlyb25tZW50OiB7XHJcbiAgICAgICAgdGFibGVOYW1lOiBteVRhYmxlLnRhYmxlTmFtZSxcclxuICAgICAgfSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIExhbWJkYSBmdW5jdGlvbiB0aGF0OlxyXG4gICAgLy8gMS4gQ2FsbHMgQW1hem9uIFJla29nbml0aW9uIHRvIGRldGVjdCBvYmplY3RzIGluIHRoZSBpbWFnZSBmaWxlXHJcbiAgICAvLyAyLiBTYXZlcyBpbmZvcm1hdGlvbiBhYm91dCB0aGUgb2JqZWN0cyBpbiBhIER5bmFtb2RiIHRhYmxlXHJcbiAgICBjb25zdCBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uID0gbmV3IGxhbWJkYS5GdW5jdGlvbih0aGlzLCAnZG9jLWV4YW1wbGUtc2F2ZS1vYmplY3QtZGF0YScsIHtcclxuICAgICAgcnVudGltZTogbGFtYmRhLlJ1bnRpbWUuR09fMV9YLFxyXG4gICAgICBoYW5kbGVyOiAnbWFpbicsXHJcbiAgICAgIGNvZGU6IG5ldyBsYW1iZGEuQXNzZXRDb2RlKCdzcmMvc2F2ZV9vYmplY3RkYXRhJyksIC8vIEdvIHNvdXJjZSBmaWxlIGlzIChyZWxhdGl2ZSB0byBjZGsuanNvbik6IHNyYy9zYXZlX29iamVjdGRhdGEvbWFpbi5nb1xyXG4gICAgICBlbnZpcm9ubWVudDoge1xyXG4gICAgICAgIHRhYmxlTmFtZTogbXlUYWJsZS50YWJsZU5hbWUsXHJcbiAgICAgIH0sXHJcbiAgICB9KTtcclxuXHJcbiAgICAvLyBMYW1iZGEgZnVuY3Rpb24gdGhhdDpcclxuICAgIC8vIDEuIEdldHMgdGhlIHBob3RvIGZyb20gUzNcclxuICAgIC8vIDIuIENyZWF0ZXMgYSB0aHVtYm5haWwgb2YgdGhlIHBob3RvXHJcbiAgICAvLyAzLiBTYXZlIHRoZSBwaG90byBiYWNrIGludG8gUzNcclxuICAgIGNvbnN0IGNyZWF0ZVRodW1ibmFpbEZ1bmN0aW9uID0gbmV3IGxhbWJkYS5GdW5jdGlvbih0aGlzLCAnZG9jLWV4YW1wbGUtY3JlYXRlLXRodW1ibmFpbCcsIHtcclxuICAgICAgcnVudGltZTogbGFtYmRhLlJ1bnRpbWUuR09fMV9YLFxyXG4gICAgICBoYW5kbGVyOiAnbWFpbicsXHJcbiAgICAgIGNvZGU6IG5ldyBsYW1iZGEuQXNzZXRDb2RlKCdzcmMvY3JlYXRlX3RodW1ibmFpbCcpLCAvLyBHbyBzb3VyY2UgZmlsZSBpcyAocmVsYXRpdmUgdG8gY2RrLmpzb24pOiBzcmMvY3JlYXRlX3RodW1ibmFpbC9tYWluLmdvXHJcbiAgICB9KTtcclxuXHJcblxyXG4gICAgLy8gQ3JlYXRlIExhbWJkYSBmdW5jdGlvbiB0byBnZXQgc3RhdHVzIG9mIHVwbG9hZGVkIGRhdGEgZm9yIHN0YXRlIG1hY2hpbmVcclxuICAgIC8qXHJcbiAgICBjb25zdCBnZXRTdGF0dXNMYW1iZGEgPSBuZXcgbGFtYmRhLkZ1bmN0aW9uKHRoaXMsICdkb2MtZXhhbXBsZS1nZXQtc3RhdHVzJywge1xyXG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5HT18xX1gsXHJcbiAgICAgIGhhbmRsZXI6ICdtYWluJyxcclxuICAgICAgY29kZTogbmV3IGxhbWJkYS5Bc3NldENvZGUoJ3NyYy9nZXRfc3RhdHVzJyksIC8vIEdvIHNvdXJjZSBmaWxlIGlzIChyZWxhdGl2ZSB0byBjZGsuanNvbik6IHNyYy9nZXRfc3RhdHVzL21haW4uZ29cclxuICAgICAgZW52aXJvbm1lbnQ6IHtcclxuICAgICAgICB0YWJsZU5hbWU6IG15VGFibGUudGFibGVOYW1lLFxyXG4gICAgICB9LFxyXG4gICAgfSk7XHJcbiAgICAqL1xyXG5cclxuXHJcbiAgICAvLyBGaXJzdCB0YXNrOiBzYXZlIG1ldGFkYXRhIGZyb20gcGhvdG8gaW4gUzMgYnVja2V0IHRvIER5bmFtb0RCIHRhYmxlXHJcbiAgICBjb25zdCBzYXZlTWV0YWRhdGFKb2IgPSBuZXcgdGFza3MuTGFtYmRhSW52b2tlKHRoaXMsICdTYXZlIE1ldGFkYXRhIEpvYicsIHtcclxuICAgICAgbGFtYmRhRnVuY3Rpb246IHNhdmVNZXRhZGF0YUZ1bmN0aW9uLFxyXG4gICAgICAvL2lucHV0UGF0aDogJyQnLCAvLyBFdmVudCBmcm9tIFMzIG5vdGlmaWNhdGlvbiAoZGVmYXVsdClcclxuICAgICAgb3V0cHV0UGF0aDogJyQuUGF5bG9hZCcsXHJcbiAgICB9KTtcclxuXHJcbiAgICAvLyBTZWNvbmQgdGFzazogc2F2ZSBpbWFnZSBkYXRhIGZyb20gUmVrb2duaXRpb24gdG8gRHluYW1vREIgdGFibGVcclxuICAgIGNvbnN0IHNhdmVPYmplY3REYXRhSm9iID0gbmV3IHRhc2tzLkxhbWJkYUludm9rZSh0aGlzLCAnU2F2ZSBPYmplY3QgRGF0YSBKb2InLCB7XHJcbiAgICAgIGxhbWJkYUZ1bmN0aW9uOiBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLlBheWxvYWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5QYXlsb2FkJyxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIEZpbmFsIHRhc2s6IGNyZWF0ZSB0aHVtYm5haWwgb2YgcGhvdG8gaW4gUzMgYnVja2V0XHJcbiAgICBjb25zdCBjcmVhdGVUaHVtYm5haWxKb2IgPSBuZXcgdGFza3MuTGFtYmRhSW52b2tlKHRoaXMsICdDcmVhdGUgVGh1bWJuYWlsIEpvYicsIHtcclxuICAgICAgbGFtYmRhRnVuY3Rpb246IGNyZWF0ZVRodW1ibmFpbEZ1bmN0aW9uLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLlBheWxvYWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5QYXlsb2FkJyxcclxuICAgIH0pO1xyXG5cclxuICAgIC8qXHJcbiAgICBjb25zdCB3YWl0WCA9IG5ldyBzZm4uV2FpdCh0aGlzLCAnV2FpdCBYIFNlY29uZHMnLCB7XHJcbiAgICAgIHRpbWU6IHNmbi5XYWl0VGltZS5kdXJhdGlvbihjZGsuRHVyYXRpb24uc2Vjb25kcyg1KSkgICAgICAgLy8uc2Vjb25kc1BhdGgoJyQuUGF5bG9hZC53YWl0U2Vjb25kcycpLFxyXG4gICAgfSk7XHJcblxyXG4gICAgY29uc3QgZ2V0U3RhdHVzID0gbmV3IHRhc2tzLkxhbWJkYUludm9rZSh0aGlzLCAnR2V0IEpvYiBTdGF0dXMnLCB7XHJcbiAgICAgIGxhbWJkYUZ1bmN0aW9uOiBnZXRTdGF0dXNMYW1iZGEsXHJcbiAgICAgIGlucHV0UGF0aDogJyQuZ3VpZCcsXHJcbiAgICAgIG91dHB1dFBhdGg6ICckLnN0YXR1cycsXHJcbiAgICB9KTtcclxuXHJcbiAgICBjb25zdCBqb2JGYWlsZWQgPSBuZXcgc2ZuLkZhaWwodGhpcywgJ0pvYiBGYWlsZWQnLCB7XHJcbiAgICAgIGNhdXNlOiAnQVdTIEJhdGNoIEpvYiBGYWlsZWQnLFxyXG4gICAgICBlcnJvcjogJ0Rlc2NyaWJlSm9iIHJldHVybmVkIEZBSUxFRCcsXHJcbiAgICB9KTtcclxuXHJcbiAgICBjb25zdCBmaW5hbFN0YXR1cyA9IG5ldyB0YXNrcy5MYW1iZGFJbnZva2UodGhpcywgJ0dldCBGaW5hbCBKb2IgU3RhdHVzJywge1xyXG4gICAgICBsYW1iZGFGdW5jdGlvbjogZ2V0U3RhdHVzTGFtYmRhLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLmd1aWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5zdGF0dXMnLFxyXG4gICAgfSk7XHJcbiAgICAqL1xyXG5cclxuICAgIC8vIENyZWF0ZSBzdGF0ZSBtYWNoaW5lIHdpdGggb25lIHRhc2ssIHN1Ym1pdEpvYlxyXG4gICAgY29uc3QgZGVmaW5pdGlvbiA9IHNhdmVNZXRhZGF0YUpvYlxyXG4gICAgICAubmV4dChzYXZlT2JqZWN0RGF0YUpvYilcclxuICAgICAgLm5leHQoY3JlYXRlVGh1bWJuYWlsSm9iKVxyXG4gICAgLy8gICAgICAubmV4dCh3YWl0WClcclxuICAgIC8vICAgICAgLm5leHQoZ2V0U3RhdHVzKVxyXG4gICAgLy8gICAgICAubmV4dChuZXcgc2ZuLkNob2ljZSh0aGlzLCAnSm9iIENvbXBsZXRlPycpXHJcbiAgICAvLyBMb29rIGF0IHRoZSBcInN0YXR1c1wiIGZpZWxkXHJcbiAgICAvLyAgICAgICAgLndoZW4oc2ZuLkNvbmRpdGlvbi5zdHJpbmdFcXVhbHMoJyQuc3RhdHVzJywgJ0ZBSUxFRCcpLCBqb2JGYWlsZWQpXHJcbiAgICAvLyAgICAgICAgLndoZW4oc2ZuLkNvbmRpdGlvbi5zdHJpbmdFcXVhbHMoJyQuc3RhdHVzJywgJ1NVQ0NFRURFRCcpLCBmaW5hbFN0YXR1cylcclxuICAgIC8vICAgICAgICAub3RoZXJ3aXNlKHdhaXRYKSk7XHJcblxyXG4gICAgY29uc3QgbXlTdGF0ZU1hY2hpbmUgPSBuZXcgc2ZuLlN0YXRlTWFjaGluZSh0aGlzLCAnU3RhdGVNYWNoaW5lJywge1xyXG4gICAgICBkZWZpbml0aW9uLFxyXG4gICAgICB0aW1lb3V0OiBjZGsuRHVyYXRpb24ubWludXRlcyg1KSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIFNlbmQgUzMgZXZlbnRzIHRvIFN0ZXAgRnVuY3Rpb25zIHN0YXRlIG1hY2hpbmVcclxuICAgIC8vICAgcnVsZS5hZGRUYXJnZXQobmV3IHRhcmdldHMuU2ZuU3RhdGVNYWNoaW5lKG15U3RhdGVNYWNoaW5lKSk7XHJcblxyXG4gICAgLypcclxuICAgIC8vIENyZWF0ZSByb2xlIGZvciBMYW1iZGEgZnVuY3Rpb24gdG8gY2FsbCBTdGVwIEZ1bmN0aW9uc1xyXG4gICAgY29uc3Qgc3RlcEZ1bmNSb2xlID0gbmV3IGlhbS5Sb2xlKHRoaXMsICdkb2MtZXhhbXBsZS1zdGVwZnVuYy1yb2xlJywge1xyXG4gICAgICByb2xlTmFtZTogJ2RvYy1leGFtcGxlLXN0ZXBmdW5jJyxcclxuICAgICAgYXNzdW1lZEJ5OiBuZXcgaWFtLlNlcnZpY2VQcmluY2lwYWwoJ2xhbWJkYS5hbWF6b25hd3MuY29tJylcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIExldCBMYW1iZGEgZnVuY3Rpb24gY2FsbCB0aGVzZSBTdGVwIEZ1bmN0aW9uIG9wZXJhdGlvbnNcclxuICAgIHN0ZXBGdW5jUm9sZS5hZGRUb1BvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIGFjdGlvbnM6IFtcIj8/P1wiLCBcIlwiXSxcclxuICAgICAgZWZmZWN0OiBpYW0uRWZmZWN0LkFMTE9XLFxyXG4gICAgICByZXNvdXJjZXM6IFtdLFxyXG4gICAgfSkpXHJcbiAgICAqL1xyXG5cclxuICAgIC8vIENyZWF0ZSByb2xlIGZvciBMYW1iZGEgZnVuY3Rpb24gdG8gY2FsbCBEeW5hbW9EQlxyXG4gICAgY29uc3QgZHluYW1vRGJSb2xlID0gbmV3IGlhbS5Sb2xlKHRoaXMsICdkb2MtZXhhbXBsZS1keW5hbW9kYi1yb2xlJywge1xyXG4gICAgICByb2xlTmFtZTogJ2RvYy1leGFtcGxlLWR5bmFtb2RiJyxcclxuICAgICAgYXNzdW1lZEJ5OiBuZXcgaWFtLlNlcnZpY2VQcmluY2lwYWwoJ2xhbWJkYS5hbWF6b25hd3MuY29tJylcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIExldCBMYW1iZGEgY2FsbCB0aGVzZSBEeW5hbW9EQiBvcGVyYXRpb25zXHJcbiAgICBkeW5hbW9EYlJvbGUuYWRkVG9Qb2xpY3kobmV3IGlhbS5Qb2xpY3lTdGF0ZW1lbnQoe1xyXG4gICAgICBlZmZlY3Q6IGlhbS5FZmZlY3QuQUxMT1csXHJcbiAgICAgIHJlc291cmNlczogW215VGFibGUudGFibGVBcm5dLFxyXG4gICAgICBhY3Rpb25zOiBbXHJcbiAgICAgICAgJ2R5bmFtb2RiOlB1dEl0ZW0nXHJcbiAgICAgIF1cclxuICAgIH0pKTtcclxuXHJcbiAgICAvKlxyXG4gICAgZHluYW1vRGJSb2xlLmFkZE1hbmFnZWRQb2xpY3koXHJcbiAgICAgIGlhbS5NYW5hZ2VkUG9saWN5LmZyb21Bd3NNYW5hZ2VkUG9saWN5TmFtZShcclxuICAgICAgICAnc2VydmljZS1yb2xlL0FXU0xhbWJkYUJhc2ljRXhlY3V0aW9uUm9sZScpKTtcclxuICAgICovXHJcblxyXG4gICAgLy8gQ29uZmlndXJlIEFtYXpvbiBTMyBidWNrZXQgdG8gc2VuZCBub3RpZmljYXRpb24gZXZlbnRzIHRvIHN0ZXAgZnVuY3Rpb25zLlxyXG4gICAgLy8gbXlCdWNrZXQuYWRkRXZlbnROb3RpZmljYXRpb24oczMuRXZlbnRUeXBlLk9CSkVDVF9DUkVBVEVELCBuZXcgbm90cy5MYW1iZGFEZXN0aW5hdGlvbihnZXRNZXRhZGF0YUZ1bmN0aW9uKSk7XHJcblxyXG5cclxuICAgIC8vIERpc3BsYXkgaW5mbyBhYm91dCB0aGUgcmVzb3VyY2VzLlxyXG4gICAgLy8gWW91IGNhbiBzZWUgdGhpcyBpbmZvcm1hdGlvbiBhdCBhbnkgdGltZSBieSBydW5uaW5nOlxyXG4gICAgLy8gICBhd3MgY2xvdWRmb3JtYXRpb24gZGVzY3JpYmUtc3RhY2tzIC0tc3RhY2stbmFtZSBJbWFnZVJlY29nU3RhY2sgLS1xdWVyeSBTdGFja3NbMF0uT3V0cHV0cyAtLW91dHB1dCB0ZXh0XHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnQnVja2V0IG5hbWU6ICcsIHsgdmFsdWU6IG15QnVja2V0LmJ1Y2tldE5hbWUgfSk7XHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnUzMgZnVuY3Rpb24gbmFtZTogJywgeyB2YWx1ZTogc2F2ZU1ldGFkYXRhRnVuY3Rpb24uZnVuY3Rpb25OYW1lIH0pO1xyXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1MzIGZ1bmN0aW9uIENsb3VkV2F0Y2ggbG9nIGdyb3VwOiAnLCB7IHZhbHVlOiBzYXZlTWV0YWRhdGFGdW5jdGlvbi5sb2dHcm91cC5sb2dHcm91cE5hbWUgfSk7XHJcblxyXG4gICAgLy8gbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1N0YXR1cyBmdW5jdGlvbjogJywgeyB2YWx1ZTogZ2V0U3RhdHVzTGFtYmRhLmZ1bmN0aW9uTmFtZSB9KTtcclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdUYWJsZSBuYW1lOiAnLCB7IHZhbHVlOiBteVRhYmxlLnRhYmxlTmFtZSB9KTtcclxuXHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnU3RhdGUgbWFjaGluZTogJywgeyB2YWx1ZTogbXlTdGF0ZU1hY2hpbmUuc3RhdGVNYWNoaW5lTmFtZSB9KTtcclxuICB9XHJcbn1cclxuXHJcbmNvbnN0IGFwcCA9IG5ldyBjZGsuQXBwKCk7XHJcbm5ldyBJbWFnZVJlY29nU3RhY2soYXBwLCAnSW1hZ2VSZWNvZ1N0YWNrJyk7Il19