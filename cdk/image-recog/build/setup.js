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
const aws_iam_1 = require("@aws-cdk/aws-iam");
class ImageRecogStack extends cdk.Stack {
    constructor(scope, id, props) {
        var _a, _b, _c;
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
        // Add policy to Lambda function so it can call
        // GetObject on bucket and PutItem on table.
        const s3Policy = new iam.PolicyStatement({
            sid: "doc-example-s3-statement",
            actions: ["s3:GetObject", "dynamodb:PutItem"],
            effect: aws_iam_1.Effect.ALLOW,
            resources: [myBucket.bucketArn + "/*", myTable.tableArn + "/*"],
        });
        (_a = saveMetadataFunction.role) === null || _a === void 0 ? void 0 : _a.addToPrincipalPolicy(s3Policy);
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
        // Add policy to Lambda function so it can call
        // PutItem on table.
        const dbPolicy = new iam.PolicyStatement({
            sid: "doc-example-s3-statement",
            actions: ["dynamodb:PutItem"],
            effect: aws_iam_1.Effect.ALLOW,
            resources: [myTable.tableArn + "/*"],
        });
        (_b = saveObjectDataFunction.role) === null || _b === void 0 ? void 0 : _b.addToPrincipalPolicy(dbPolicy);
        // Lambda function that:
        // 1. Gets the photo from S3
        // 2. Creates a thumbnail of the photo
        // 3. Save the photo back into S3
        const createThumbnailFunction = new lambda.Function(this, 'doc-example-create-thumbnail', {
            runtime: lambda.Runtime.GO_1_X,
            handler: 'main',
            code: new lambda.AssetCode('src/create_thumbnail'),
        });
        // Add policy to Lambda function so it can call
        // GetObject and PutObject on bucket.
        const s32Policy = new iam.PolicyStatement({
            sid: "doc-example-s3-statement",
            actions: ["s3:GetObject", "s3:PutObject"],
            effect: aws_iam_1.Effect.ALLOW,
            resources: [myBucket.bucketArn + "/*"],
        });
        (_c = createThumbnailFunction.role) === null || _c === void 0 ? void 0 : _c.addToPrincipalPolicy(s32Policy);
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
        new cdk.CfnOutput(this, 'Save metadata function: ', { value: saveMetadataFunction.functionName });
        new cdk.CfnOutput(this, 'Save object data function: ', { value: saveObjectDataFunction.functionName });
        new cdk.CfnOutput(this, 'Create thumbnail function: ', { value: createThumbnailFunction.functionName });
        new cdk.CfnOutput(this, 'S3 function CloudWatch log group: ', { value: saveMetadataFunction.logGroup.logGroupName });
        // new cdk.CfnOutput(this, 'Status function: ', { value: getStatusLambda.functionName });
        new cdk.CfnOutput(this, 'Table name: ', { value: myTable.tableName });
        new cdk.CfnOutput(this, 'State machine: ', { value: myStateMachine.stateMachineName });
    }
}
exports.ImageRecogStack = ImageRecogStack;
const app = new cdk.App();
new ImageRecogStack(app, 'ImageRecogStack');
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoic2V0dXAuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi9zZXR1cC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7O0FBQ0EsdUNBQXFDO0FBRXJDLHFDQUFxQztBQUNyQyw0Q0FBNEM7QUFDNUMsdURBQXVEO0FBQ3ZELG1EQUFtRDtBQUNuRCxzQ0FBc0M7QUFDdEMsc0RBQXNEO0FBQ3RELDhDQUE4QztBQUU5Qyx5REFBeUQ7QUFFekQseURBQXlEO0FBQ3pELHdDQUF3QztBQUN4Qyw4Q0FBOEM7QUFDOUMsa0RBQWtEO0FBQ2xELGtEQUFrRDtBQUVsRCwwREFBMEQ7QUFDMUQsOENBQTBDO0FBRTFDLE1BQWEsZUFBZ0IsU0FBUSxHQUFHLENBQUMsS0FBSztJQUM1QyxZQUFZLEtBQW9CLEVBQUUsRUFBVSxFQUFFLEtBQXNCOztRQUNsRSxLQUFLLENBQUMsS0FBSyxFQUFFLEVBQUUsRUFBRSxLQUFLLENBQUMsQ0FBQztRQUV4Qjs7Ozs7O1dBTUc7UUFFSCwwREFBMEQ7UUFDMUQsTUFBTSxRQUFRLEdBQUcsSUFBSSxFQUFFLENBQUMsTUFBTSxDQUFDLElBQUksRUFBRSxvQkFBb0IsQ0FBQyxDQUFDO1FBRTNELCtDQUErQztRQUMvQyxNQUFNLE9BQU8sR0FBRyxJQUFJLFVBQVUsQ0FBQyxLQUFLLENBQUMsSUFBSSxFQUFFLG1CQUFtQixDQUFDLENBQUM7UUFDaEUsNkNBQTZDO1FBQzdDLDBDQUEwQztRQUMxQywrQkFBK0I7UUFDL0IsT0FBTyxDQUFDLGtCQUFrQixDQUFDLENBQUM7Z0JBQzFCLE1BQU0sRUFBRSxRQUFRO2dCQUNoQixZQUFZLEVBQUUsVUFBVTthQUN6QixFQUFFLENBQUMsQ0FBQztRQUVMLHFCQUFxQjtRQUNyQixNQUFNLElBQUksR0FBRyxJQUFJLE1BQU0sQ0FBQyxJQUFJLENBQUMsSUFBSSxFQUFFLE1BQU0sRUFBRTtZQUN6QyxZQUFZLEVBQUU7Z0JBQ1osTUFBTSxFQUFFLENBQUMsUUFBUSxDQUFDO2FBQ25CO1NBQ0YsQ0FBQyxD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Q0FBQyxJQUFJLEVBQUUsNkJBQTZCLEVBQUUsRUFBRSxLQUFLLEVBQUUsc0JBQXNCLENBQUMsWUFBWSxFQUFFLENBQUMsQ0FBQztRQUN2RyxJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLDZCQUE2QixFQUFFLEVBQUUsS0FBSyxFQUFFLHVCQUF1QixDQUFDLFlBQVksRUFBRSxDQUFDLENBQUM7UUFFeEcsSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSxvQ0FBb0MsRUFBRSxFQUFFLEtBQUssRUFBRSxvQkFBb0IsQ0FBQyxRQUFRLENBQUMsWUFBWSxFQUFFLENBQUMsQ0FBQztRQUVySCx5RkFBeUY7UUFDekYsSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSxjQUFjLEVBQUUsRUFBRSxLQUFLLEVBQUUsT0FBTyxDQUFDLFNBQVMsRUFBRSxDQUFDLENBQUM7UUFFdEUsSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSxpQkFBaUIsRUFBRSxFQUFFLEtBQUssRUFBRSxjQUFjLENBQUMsZ0JBQWdCLEVBQUUsQ0FBQyxDQUFDO0lBQ3pGLENBQUM7Q0FDRjtBQTFQRCwwQ0EwUEM7QUFFRCxNQUFNLEdBQUcsR0FBRyxJQUFJLEdBQUcsQ0FBQyxHQUFHLEVBQUUsQ0FBQztBQUMxQixJQUFJLGVBQWUsQ0FBQyxHQUFHLEVBQUUsaUJBQWlCLENBQUMsQ0FBQyIsInNvdXJjZXNDb250ZW50IjpbIiMhL3Vzci9iaW4vZW52IG5vZGVcclxuaW1wb3J0ICdzb3VyY2UtbWFwLXN1cHBvcnQvcmVnaXN0ZXInO1xyXG5cclxuaW1wb3J0ICogYXMgY2RrIGZyb20gJ0Bhd3MtY2RrL2NvcmUnO1xyXG4vLyBpbXBvcnQgeyBEdXJhdGlvbiB9IGZyb20gXCJAYXdzLWNkay9jb3JlXCI7XHJcbi8vIGltcG9ydCAqIGFzIGNvZGVidWlsZCBmcm9tICdAYXdzLWNkay9hd3MtY29kZWJ1aWxkJztcclxuLy8gaW1wb3J0ICogYXMgYW1wbGlmeSBmcm9tICdAYXdzLWNkay9hd3MtYW1wbGlmeSc7XHJcbmltcG9ydCAqIGFzIHMzIGZyb20gJ0Bhd3MtY2RrL2F3cy1zMyc7XHJcbmltcG9ydCAqIGFzIGNsb3VkdHJhaWwgZnJvbSAnQGF3cy1jZGsvYXdzLWNsb3VkdHJhaWwnO1xyXG5pbXBvcnQgKiBhcyBldmVudHMgZnJvbSAnQGF3cy1jZGsvYXdzLWV2ZW50cyc7XHJcbmltcG9ydCAqIGFzIHRhcmdldHMgZnJvbSAnQGF3cy1jZGsvYXdzLWV2ZW50cy10YXJnZXRzJztcclxuLy8gaW1wb3J0ICogYXMgY2xvdWR3YXRjaCBmcm9tICdAYXdzLWNkay9hd3MtY2xvdWR3YXRjaCc7XHJcblxyXG4vLyBpbXBvcnQgKiBhcyBub3RzIGZyb20gJ0Bhd3MtY2RrL2F3cy1zMy1ub3RpZmljYXRpb25zJztcclxuaW1wb3J0ICogYXMgaWFtIGZyb20gJ0Bhd3MtY2RrL2F3cy1pYW0nO1xyXG5pbXBvcnQgKiBhcyBsYW1iZGEgZnJvbSAnQGF3cy1jZGsvYXdzLWxhbWJkYSc7XHJcbmltcG9ydCAqIGFzIGR5bmFtb2RiIGZyb20gJ0Bhd3MtY2RrL2F3cy1keW5hbW9kYic7XHJcbmltcG9ydCAqIGFzIHNmbiBmcm9tICdAYXdzLWNkay9hd3Mtc3RlcGZ1bmN0aW9ucyc7XHJcbmltcG9ydCB7IFdhaXRUaW1lIH0gZnJvbSBcIkBhd3MtY2RrL2F3cy1zdGVwZnVuY3Rpb25zXCI7XHJcbmltcG9ydCAqIGFzIHRhc2tzIGZyb20gJ0Bhd3MtY2RrL2F3cy1zdGVwZnVuY3Rpb25zLXRhc2tzJztcclxuaW1wb3J0IHsgRWZmZWN0IH0gZnJvbSAnQGF3cy1jZGsvYXdzLWlhbSc7XHJcblxyXG5leHBvcnQgY2xhc3MgSW1hZ2VSZWNvZ1N0YWNrIGV4dGVuZHMgY2RrLlN0YWNrIHtcclxuICBjb25zdHJ1Y3RvcihzY29wZTogY2RrLkNvbnN0cnVjdCwgaWQ6IHN0cmluZywgcHJvcHM/OiBjZGsuU3RhY2tQcm9wcykge1xyXG4gICAgc3VwZXIoc2NvcGUsIGlkLCBwcm9wcyk7XHJcblxyXG4gICAgLyogVXNlIGJ1Y2tldCBldmVudCB0byBleGVjdXRlIGEgc3RlcCBmdW5jdGlvbiB3aGVuIGFuIGl0ZW0gdXBsb2FkZWQgdG8gYSBidWNrZXRcclxuICAgICAqICAgaHR0cHM6Ly9kb2NzLmF3cy5hbWF6b24uY29tL3N0ZXAtZnVuY3Rpb25zL2xhdGVzdC9kZy90dXRvcmlhbC1jbG91ZHdhdGNoLWV2ZW50cy1zMy5odG1sXHJcbiAgICAgKlxyXG4gICAgICogMTogQ3JlYXRlIGEgYnVja2V0IChBbWF6b24gUzMpXHJcbiAgICAgKiAyOiBDcmVhdGUgYSB0cmFpbCAoQVdTIENsb3VkVHJhaWwpXHJcbiAgICAgKiAzOiBDcmVhdGUgYW4gZXZlbnRzIHJ1bGUgKEFXUyBDbG91ZFdhdGNoIEV2ZW50cylcclxuICAgICAqL1xyXG5cclxuICAgIC8vIENyZWF0ZSBBbWF6b24gU2ltcGxlIFN0b3JhZ2UgU2VydmljZSAoQW1hem9uIFMzKSBidWNrZXRcclxuICAgIGNvbnN0IG15QnVja2V0ID0gbmV3IHMzLkJ1Y2tldCh0aGlzLCAnZG9jLWV4YW1wbGUtYnVja2V0Jyk7XHJcblxyXG4gICAgLy8gQ3JlYXRlIHRyYWlsIHRvIHdhdGNoIGZvciBldmVudHMgZnJvbSBidWNrZXRcclxuICAgIGNvbnN0IG15VHJhaWwgPSBuZXcgY2xvdWR0cmFpbC5UcmFpbCh0aGlzLCAnZG9jLWV4YW1wbGUtdHJhaWwnKTtcclxuICAgIC8vIEFkZCBhbiBldmVudCBzZWxlY3RvciB0byB0aGUgdHJhaWwgc28gdGhhdFxyXG4gICAgLy8gSlBHIG9yIFBORyBmaWxlcyB3aXRoICd1cGxvYWRzLycgcHJlZml4XHJcbiAgICAvLyBhZGRlZCB0byBidWNrZXQgYXJlIGRldGVjdGVkXHJcbiAgICBteVRyYWlsLmFkZFMzRXZlbnRTZWxlY3Rvcihbe1xyXG4gICAgICBidWNrZXQ6IG15QnVja2V0LFxyXG4gICAgICBvYmplY3RQcmVmaXg6ICd1cGxvYWRzLycsXHJcbiAgICB9LF0pO1xyXG5cclxuICAgIC8vIENyZWF0ZSBldmVudHMgcnVsZVxyXG4gICAgY29uc3QgcnVsZSA9IG5ldyBldmVudHMuUnVsZSh0aGlzLCAncnVsZScsIHtcclxuICAgICAgZXZlbnRQYXR0ZXJuOiB7XHJcbiAgICAgICAgc291cmNlOiBbJ2F3cy5zMyddLFxyXG4gICAgICB9LFxyXG4gICAgfSk7XHJcblxyXG4gICAgLy8gQ3JlYXRlIER5bmFtb0RCIHRhYmxlIGZvciBMYW1iZGEgZnVuY3Rpb24gdG8gcGVyc2lzdCBpbWFnZSBpbmZvXHJcbiAgICAvLyBDcmVhdGUgQW1hem9uIER5bmFtb0RCIHRhYmxlIHdpdGggcHJpbWFyeSBrZXkgcGF0aCAoc3RyaW5nKVxyXG4gICAgLy8gdGhhdCB3aWxsIGJlIHNvbWV0aGluZyBsaWtlIHVwbG9hZHMvbXlQaG90by5qcGdcclxuICAgIGNvbnN0IG15VGFibGUgPSBuZXcgZHluYW1vZGIuVGFibGUodGhpcywgJ2RvYy1leGFtcGxlLXRhYmxlJywge1xyXG4gICAgICBwYXJ0aXRpb25LZXk6IHsgbmFtZTogJ3BhdGgnLCB0eXBlOiBkeW5hbW9kYi5BdHRyaWJ1dGVUeXBlLlNUUklORyB9LFxyXG4gICAgICBzdHJlYW06IGR5bmFtb2RiLlN0cmVhbVZpZXdUeXBlLk5FV19JTUFHRSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8qIFxyXG4gICAgICogRGVmaW5lIExhbWJkYSBmdW5jdGlvbnMgdG86XHJcbiAgICAgKiAxLiBBZGQgbWV0YWRhdGEgZnJvbSB0aGUgcGhvdG8gdG8gYSBEeW5hbW9kYiB0YWJsZS4gICAgIFxyXG4gICAgICogMi4gQ2FsbCBBbWF6b24gUmVrb2duaXRpb24gdG8gZGV0ZWN0IG9iamVjdHMgaW4gdGhlIGltYWdlIGZpbGUuXHJcbiAgICAgKiAzLiBHZW5lcmF0ZSBhIHRodW1ibmFpbCBhbmQgc3RvcmUgaXQgaW4gdGhlIFMzIGJ1Y2tldCB3aXRoIHRoZSAqKnJlc2l6ZWQvKiogcHJlZml4XHJcbiAgICAgKi9cclxuXHJcbiAgICAvLyBMYW1iZGEgZnVuY3Rpb24gdGhhdDpcclxuICAgIC8vIDEuIFJlY2VpdmVzIG5vdGlmaWNhdGlvbnMgZnJvbSBBbWF6b24gUzMgKEl0ZW1VcGxvYWQpXHJcbiAgICAvLyAyLiBHZXRzIG1ldGFkYXRhIGZyb20gdGhlIHBob3RvXHJcbiAgICAvLyAzLiBTYXZlcyB0aGUgbWV0YWRhdGEgaW4gYSBEeW5hbW9EQiB0YWJsZVxyXG4gICAgY29uc3Qgc2F2ZU1ldGFkYXRhRnVuY3Rpb24gPSBuZXcgbGFtYmRhLkZ1bmN0aW9uKHRoaXMsICdkb2MtZXhhbXBsZS1zYXZlLW1ldGFkYXRhJywge1xyXG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5HT18xX1gsXHJcbiAgICAgIGhhbmRsZXI6ICdtYWluJyxcclxuICAgICAgY29kZTogbmV3IGxhbWJkYS5Bc3NldENvZGUoJ3NyYy9zYXZlX21ldGFkYXRhJyksIC8vIEdvIHNvdXJjZSBmaWxlIGlzIChyZWxhdGl2ZSB0byBjZGsuanNvbik6IHNyYy9zYXZlX21ldGFkYXRhL21haW4uZ29cclxuICAgICAgZW52aXJvbm1lbnQ6IHtcclxuICAgICAgICB0YWJsZU5hbWU6IG15VGFibGUudGFibGVOYW1lLFxyXG4gICAgICB9LFxyXG4gICAgfSk7XHJcblxyXG4gICAgLy8gQWRkIHBvbGljeSB0byBMYW1iZGEgZnVuY3Rpb24gc28gaXQgY2FuIGNhbGxcclxuICAgIC8vIEdldE9iamVjdCBvbiBidWNrZXQgYW5kIFB1dEl0ZW0gb24gdGFibGUuXHJcbiAgICBjb25zdCBzM1BvbGljeSA9IG5ldyBpYW0uUG9saWN5U3RhdGVtZW50KHtcclxuICAgICAgc2lkOiBcImRvYy1leGFtcGxlLXMzLXN0YXRlbWVudFwiLFxyXG4gICAgICBhY3Rpb25zOiBbXCJzMzpHZXRPYmplY3RcIiwgXCJkeW5hbW9kYjpQdXRJdGVtXCJdLFxyXG4gICAgICBlZmZlY3Q6IEVmZmVjdC5BTExPVyxcclxuICAgICAgcmVzb3VyY2VzOiBbbXlCdWNrZXQuYnVja2V0QXJuICsgXCIvKlwiLCBteVRhYmxlLnRhYmxlQXJuICsgXCIvKlwiXSxcclxuICAgIH0pXHJcblxyXG4gICAgc2F2ZU1ldGFkYXRhRnVuY3Rpb24ucm9sZT8uYWRkVG9QcmluY2lwYWxQb2xpY3koczNQb2xpY3kpXHJcblxyXG4gICAgLy8gTGFtYmRhIGZ1bmN0aW9uIHRoYXQ6XHJcbiAgICAvLyAxLiBDYWxscyBBbWF6b24gUmVrb2duaXRpb24gdG8gZGV0ZWN0IG9iamVjdHMgaW4gdGhlIGltYWdlIGZpbGVcclxuICAgIC8vIDIuIFNhdmVzIGluZm9ybWF0aW9uIGFib3V0IHRoZSBvYmplY3RzIGluIGEgRHluYW1vZGIgdGFibGVcclxuICAgIGNvbnN0IHNhdmVPYmplY3REYXRhRnVuY3Rpb24gPSBuZXcgbGFtYmRhLkZ1bmN0aW9uKHRoaXMsICdkb2MtZXhhbXBsZS1zYXZlLW9iamVjdC1kYXRhJywge1xyXG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5HT18xX1gsXHJcbiAgICAgIGhhbmRsZXI6ICdtYWluJyxcclxuICAgICAgY29kZTogbmV3IGxhbWJkYS5Bc3NldENvZGUoJ3NyYy9zYXZlX29iamVjdGRhdGEnKSwgLy8gR28gc291cmNlIGZpbGUgaXMgKHJlbGF0aXZlIHRvIGNkay5qc29uKTogc3JjL3NhdmVfb2JqZWN0ZGF0YS9tYWluLmdvXHJcbiAgICAgIGVudmlyb25tZW50OiB7XHJcbiAgICAgICAgdGFibGVOYW1lOiBteVRhYmxlLnRhYmxlTmFtZSxcclxuICAgICAgfSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIEFkZCBwb2xpY3kgdG8gTGFtYmRhIGZ1bmN0aW9uIHNvIGl0IGNhbiBjYWxsXHJcbiAgICAvLyBQdXRJdGVtIG9uIHRhYmxlLlxyXG4gICAgY29uc3QgZGJQb2xpY3kgPSBuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIHNpZDogXCJkb2MtZXhhbXBsZS1zMy1zdGF0ZW1lbnRcIixcclxuICAgICAgYWN0aW9uczogW1wiZHluYW1vZGI6UHV0SXRlbVwiXSxcclxuICAgICAgZWZmZWN0OiBFZmZlY3QuQUxMT1csXHJcbiAgICAgIHJlc291cmNlczogW215VGFibGUudGFibGVBcm4gKyBcIi8qXCJdLFxyXG4gICAgfSlcclxuXHJcbiAgICBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uLnJvbGU/LmFkZFRvUHJpbmNpcGFsUG9saWN5KGRiUG9saWN5KVxyXG5cclxuICAgIC8vIExhbWJkYSBmdW5jdGlvbiB0aGF0OlxyXG4gICAgLy8gMS4gR2V0cyB0aGUgcGhvdG8gZnJvbSBTM1xyXG4gICAgLy8gMi4gQ3JlYXRlcyBhIHRodW1ibmFpbCBvZiB0aGUgcGhvdG9cclxuICAgIC8vIDMuIFNhdmUgdGhlIHBob3RvIGJhY2sgaW50byBTM1xyXG4gICAgY29uc3QgY3JlYXRlVGh1bWJuYWlsRnVuY3Rpb24gPSBuZXcgbGFtYmRhLkZ1bmN0aW9uKHRoaXMsICdkb2MtZXhhbXBsZS1jcmVhdGUtdGh1bWJuYWlsJywge1xyXG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5HT18xX1gsXHJcbiAgICAgIGhhbmRsZXI6ICdtYWluJyxcclxuICAgICAgY29kZTogbmV3IGxhbWJkYS5Bc3NldENvZGUoJ3NyYy9jcmVhdGVfdGh1bWJuYWlsJyksIC8vIEdvIHNvdXJjZSBmaWxlIGlzIChyZWxhdGl2ZSB0byBjZGsuanNvbik6IHNyYy9jcmVhdGVfdGh1bWJuYWlsL21haW4uZ29cclxuICAgIH0pO1xyXG5cclxuICAgIC8vIEFkZCBwb2xpY3kgdG8gTGFtYmRhIGZ1bmN0aW9uIHNvIGl0IGNhbiBjYWxsXHJcbiAgICAvLyBHZXRPYmplY3QgYW5kIFB1dE9iamVjdCBvbiBidWNrZXQuXHJcbiAgICBjb25zdCBzMzJQb2xpY3kgPSBuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIHNpZDogXCJkb2MtZXhhbXBsZS1zMy1zdGF0ZW1lbnRcIixcclxuICAgICAgYWN0aW9uczogW1wiczM6R2V0T2JqZWN0XCIsIFwiczM6UHV0T2JqZWN0XCJdLFxyXG4gICAgICBlZmZlY3Q6IEVmZmVjdC5BTExPVyxcclxuICAgICAgcmVzb3VyY2VzOiBbbXlCdWNrZXQuYnVja2V0QXJuICsgXCIvKlwiXSxcclxuICAgIH0pXHJcblxyXG4gICAgY3JlYXRlVGh1bWJuYWlsRnVuY3Rpb24ucm9sZT8uYWRkVG9QcmluY2lwYWxQb2xpY3koczMyUG9saWN5KVxyXG5cclxuXHJcblxyXG4gICAgLy8gQ3JlYXRlIExhbWJkYSBmdW5jdGlvbiB0byBnZXQgc3RhdHVzIG9mIHVwbG9hZGVkIGRhdGEgZm9yIHN0YXRlIG1hY2hpbmVcclxuICAgIC8qXHJcbiAgICBjb25zdCBnZXRTdGF0dXNMYW1iZGEgPSBuZXcgbGFtYmRhLkZ1bmN0aW9uKHRoaXMsICdkb2MtZXhhbXBsZS1nZXQtc3RhdHVzJywge1xyXG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5HT18xX1gsXHJcbiAgICAgIGhhbmRsZXI6ICdtYWluJyxcclxuICAgICAgY29kZTogbmV3IGxhbWJkYS5Bc3NldENvZGUoJ3NyYy9nZXRfc3RhdHVzJyksIC8vIEdvIHNvdXJjZSBmaWxlIGlzIChyZWxhdGl2ZSB0byBjZGsuanNvbik6IHNyYy9nZXRfc3RhdHVzL21haW4uZ29cclxuICAgICAgZW52aXJvbm1lbnQ6IHtcclxuICAgICAgICB0YWJsZU5hbWU6IG15VGFibGUudGFibGVOYW1lLFxyXG4gICAgICB9LFxyXG4gICAgfSk7XHJcbiAgICAqL1xyXG5cclxuXHJcbiAgICAvLyBGaXJzdCB0YXNrOiBzYXZlIG1ldGFkYXRhIGZyb20gcGhvdG8gaW4gUzMgYnVja2V0IHRvIER5bmFtb0RCIHRhYmxlXHJcbiAgICBjb25zdCBzYXZlTWV0YWRhdGFKb2IgPSBuZXcgdGFza3MuTGFtYmRhSW52b2tlKHRoaXMsICdTYXZlIE1ldGFkYXRhIEpvYicsIHtcclxuICAgICAgbGFtYmRhRnVuY3Rpb246IHNhdmVNZXRhZGF0YUZ1bmN0aW9uLFxyXG4gICAgICAvL2lucHV0UGF0aDogJyQnLCAvLyBFdmVudCBmcm9tIFMzIG5vdGlmaWNhdGlvbiAoZGVmYXVsdClcclxuICAgICAgb3V0cHV0UGF0aDogJyQuUGF5bG9hZCcsXHJcbiAgICB9KTtcclxuXHJcbiAgICAvLyBTZWNvbmQgdGFzazogc2F2ZSBpbWFnZSBkYXRhIGZyb20gUmVrb2duaXRpb24gdG8gRHluYW1vREIgdGFibGVcclxuICAgIGNvbnN0IHNhdmVPYmplY3REYXRhSm9iID0gbmV3IHRhc2tzLkxhbWJkYUludm9rZSh0aGlzLCAnU2F2ZSBPYmplY3QgRGF0YSBKb2InLCB7XHJcbiAgICAgIGxhbWJkYUZ1bmN0aW9uOiBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLlBheWxvYWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5QYXlsb2FkJyxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIEZpbmFsIHRhc2s6IGNyZWF0ZSB0aHVtYm5haWwgb2YgcGhvdG8gaW4gUzMgYnVja2V0XHJcbiAgICBjb25zdCBjcmVhdGVUaHVtYm5haWxKb2IgPSBuZXcgdGFza3MuTGFtYmRhSW52b2tlKHRoaXMsICdDcmVhdGUgVGh1bWJuYWlsIEpvYicsIHtcclxuICAgICAgbGFtYmRhRnVuY3Rpb246IGNyZWF0ZVRodW1ibmFpbEZ1bmN0aW9uLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLlBheWxvYWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5QYXlsb2FkJyxcclxuICAgIH0pO1xyXG5cclxuICAgIC8qXHJcbiAgICBjb25zdCB3YWl0WCA9IG5ldyBzZm4uV2FpdCh0aGlzLCAnV2FpdCBYIFNlY29uZHMnLCB7XHJcbiAgICAgIHRpbWU6IHNmbi5XYWl0VGltZS5kdXJhdGlvbihjZGsuRHVyYXRpb24uc2Vjb25kcyg1KSkgICAgICAgLy8uc2Vjb25kc1BhdGgoJyQuUGF5bG9hZC53YWl0U2Vjb25kcycpLFxyXG4gICAgfSk7XHJcblxyXG4gICAgY29uc3QgZ2V0U3RhdHVzID0gbmV3IHRhc2tzLkxhbWJkYUludm9rZSh0aGlzLCAnR2V0IEpvYiBTdGF0dXMnLCB7XHJcbiAgICAgIGxhbWJkYUZ1bmN0aW9uOiBnZXRTdGF0dXNMYW1iZGEsXHJcbiAgICAgIGlucHV0UGF0aDogJyQuZ3VpZCcsXHJcbiAgICAgIG91dHB1dFBhdGg6ICckLnN0YXR1cycsXHJcbiAgICB9KTtcclxuXHJcbiAgICBjb25zdCBqb2JGYWlsZWQgPSBuZXcgc2ZuLkZhaWwodGhpcywgJ0pvYiBGYWlsZWQnLCB7XHJcbiAgICAgIGNhdXNlOiAnQVdTIEJhdGNoIEpvYiBGYWlsZWQnLFxyXG4gICAgICBlcnJvcjogJ0Rlc2NyaWJlSm9iIHJldHVybmVkIEZBSUxFRCcsXHJcbiAgICB9KTtcclxuXHJcbiAgICBjb25zdCBmaW5hbFN0YXR1cyA9IG5ldyB0YXNrcy5MYW1iZGFJbnZva2UodGhpcywgJ0dldCBGaW5hbCBKb2IgU3RhdHVzJywge1xyXG4gICAgICBsYW1iZGFGdW5jdGlvbjogZ2V0U3RhdHVzTGFtYmRhLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLmd1aWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5zdGF0dXMnLFxyXG4gICAgfSk7XHJcbiAgICAqL1xyXG5cclxuICAgIC8vIENyZWF0ZSBzdGF0ZSBtYWNoaW5lIHdpdGggb25lIHRhc2ssIHN1Ym1pdEpvYlxyXG4gICAgY29uc3QgZGVmaW5pdGlvbiA9IHNhdmVNZXRhZGF0YUpvYlxyXG4gICAgICAubmV4dChzYXZlT2JqZWN0RGF0YUpvYilcclxuICAgICAgLm5leHQoY3JlYXRlVGh1bWJuYWlsSm9iKVxyXG4gICAgLy8gICAgICAubmV4dCh3YWl0WClcclxuICAgIC8vICAgICAgLm5leHQoZ2V0U3RhdHVzKVxyXG4gICAgLy8gICAgICAubmV4dChuZXcgc2ZuLkNob2ljZSh0aGlzLCAnSm9iIENvbXBsZXRlPycpXHJcbiAgICAvLyBMb29rIGF0IHRoZSBcInN0YXR1c1wiIGZpZWxkXHJcbiAgICAvLyAgICAgICAgLndoZW4oc2ZuLkNvbmRpdGlvbi5zdHJpbmdFcXVhbHMoJyQuc3RhdHVzJywgJ0ZBSUxFRCcpLCBqb2JGYWlsZWQpXHJcbiAgICAvLyAgICAgICAgLndoZW4oc2ZuLkNvbmRpdGlvbi5zdHJpbmdFcXVhbHMoJyQuc3RhdHVzJywgJ1NVQ0NFRURFRCcpLCBmaW5hbFN0YXR1cylcclxuICAgIC8vICAgICAgICAub3RoZXJ3aXNlKHdhaXRYKSk7XHJcblxyXG4gICAgY29uc3QgbXlTdGF0ZU1hY2hpbmUgPSBuZXcgc2ZuLlN0YXRlTWFjaGluZSh0aGlzLCAnU3RhdGVNYWNoaW5lJywge1xyXG4gICAgICBkZWZpbml0aW9uLFxyXG4gICAgICB0aW1lb3V0OiBjZGsuRHVyYXRpb24ubWludXRlcyg1KSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIFNlbmQgUzMgZXZlbnRzIHRvIFN0ZXAgRnVuY3Rpb25zIHN0YXRlIG1hY2hpbmVcclxuICAgIC8vICAgcnVsZS5hZGRUYXJnZXQobmV3IHRhcmdldHMuU2ZuU3RhdGVNYWNoaW5lKG15U3RhdGVNYWNoaW5lKSk7XHJcblxyXG4gICAgLypcclxuICAgIC8vIENyZWF0ZSByb2xlIGZvciBMYW1iZGEgZnVuY3Rpb24gdG8gY2FsbCBTdGVwIEZ1bmN0aW9uc1xyXG4gICAgY29uc3Qgc3RlcEZ1bmNSb2xlID0gbmV3IGlhbS5Sb2xlKHRoaXMsICdkb2MtZXhhbXBsZS1zdGVwZnVuYy1yb2xlJywge1xyXG4gICAgICByb2xlTmFtZTogJ2RvYy1leGFtcGxlLXN0ZXBmdW5jJyxcclxuICAgICAgYXNzdW1lZEJ5OiBuZXcgaWFtLlNlcnZpY2VQcmluY2lwYWwoJ2xhbWJkYS5hbWF6b25hd3MuY29tJylcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIExldCBMYW1iZGEgZnVuY3Rpb24gY2FsbCB0aGVzZSBTdGVwIEZ1bmN0aW9uIG9wZXJhdGlvbnNcclxuICAgIHN0ZXBGdW5jUm9sZS5hZGRUb1BvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIGFjdGlvbnM6IFtcIj8/P1wiLCBcIlwiXSxcclxuICAgICAgZWZmZWN0OiBpYW0uRWZmZWN0LkFMTE9XLFxyXG4gICAgICByZXNvdXJjZXM6IFtdLFxyXG4gICAgfSkpXHJcbiAgICAqL1xyXG5cclxuICAgIC8vIENyZWF0ZSByb2xlIGZvciBMYW1iZGEgZnVuY3Rpb24gdG8gY2FsbCBEeW5hbW9EQlxyXG4gICAgY29uc3QgZHluYW1vRGJSb2xlID0gbmV3IGlhbS5Sb2xlKHRoaXMsICdkb2MtZXhhbXBsZS1keW5hbW9kYi1yb2xlJywge1xyXG4gICAgICByb2xlTmFtZTogJ2RvYy1leGFtcGxlLWR5bmFtb2RiJyxcclxuICAgICAgYXNzdW1lZEJ5OiBuZXcgaWFtLlNlcnZpY2VQcmluY2lwYWwoJ2xhbWJkYS5hbWF6b25hd3MuY29tJylcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIExldCBMYW1iZGEgY2FsbCB0aGVzZSBEeW5hbW9EQiBvcGVyYXRpb25zXHJcbiAgICBkeW5hbW9EYlJvbGUuYWRkVG9Qb2xpY3kobmV3IGlhbS5Qb2xpY3lTdGF0ZW1lbnQoe1xyXG4gICAgICBlZmZlY3Q6IGlhbS5FZmZlY3QuQUxMT1csXHJcbiAgICAgIHJlc291cmNlczogW215VGFibGUudGFibGVBcm5dLFxyXG4gICAgICBhY3Rpb25zOiBbXHJcbiAgICAgICAgJ2R5bmFtb2RiOlB1dEl0ZW0nXHJcbiAgICAgIF1cclxuICAgIH0pKTtcclxuXHJcbiAgICAvKlxyXG4gICAgZHluYW1vRGJSb2xlLmFkZE1hbmFnZWRQb2xpY3koXHJcbiAgICAgIGlhbS5NYW5hZ2VkUG9saWN5LmZyb21Bd3NNYW5hZ2VkUG9saWN5TmFtZShcclxuICAgICAgICAnc2VydmljZS1yb2xlL0FXU0xhbWJkYUJhc2ljRXhlY3V0aW9uUm9sZScpKTtcclxuICAgICovXHJcblxyXG4gICAgLy8gQ29uZmlndXJlIEFtYXpvbiBTMyBidWNrZXQgdG8gc2VuZCBub3RpZmljYXRpb24gZXZlbnRzIHRvIHN0ZXAgZnVuY3Rpb25zLlxyXG4gICAgLy8gbXlCdWNrZXQuYWRkRXZlbnROb3RpZmljYXRpb24oczMuRXZlbnRUeXBlLk9CSkVDVF9DUkVBVEVELCBuZXcgbm90cy5MYW1iZGFEZXN0aW5hdGlvbihnZXRNZXRhZGF0YUZ1bmN0aW9uKSk7XHJcblxyXG5cclxuICAgIC8vIERpc3BsYXkgaW5mbyBhYm91dCB0aGUgcmVzb3VyY2VzLlxyXG4gICAgLy8gWW91IGNhbiBzZWUgdGhpcyBpbmZvcm1hdGlvbiBhdCBhbnkgdGltZSBieSBydW5uaW5nOlxyXG4gICAgLy8gICBhd3MgY2xvdWRmb3JtYXRpb24gZGVzY3JpYmUtc3RhY2tzIC0tc3RhY2stbmFtZSBJbWFnZVJlY29nU3RhY2sgLS1xdWVyeSBTdGFja3NbMF0uT3V0cHV0cyAtLW91dHB1dCB0ZXh0XHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnQnVja2V0IG5hbWU6ICcsIHsgdmFsdWU6IG15QnVja2V0LmJ1Y2tldE5hbWUgfSk7XHJcblxyXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1NhdmUgbWV0YWRhdGEgZnVuY3Rpb246ICcsIHsgdmFsdWU6IHNhdmVNZXRhZGF0YUZ1bmN0aW9uLmZ1bmN0aW9uTmFtZSB9KTtcclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdTYXZlIG9iamVjdCBkYXRhIGZ1bmN0aW9uOiAnLCB7IHZhbHVlOiBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uLmZ1bmN0aW9uTmFtZSB9KTtcclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdDcmVhdGUgdGh1bWJuYWlsIGZ1bmN0aW9uOiAnLCB7IHZhbHVlOiBjcmVhdGVUaHVtYm5haWxGdW5jdGlvbi5mdW5jdGlvbk5hbWUgfSk7XHJcblxyXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1MzIGZ1bmN0aW9uIENsb3VkV2F0Y2ggbG9nIGdyb3VwOiAnLCB7IHZhbHVlOiBzYXZlTWV0YWRhdGFGdW5jdGlvbi5sb2dHcm91cC5sb2dHcm91cE5hbWUgfSk7XHJcblxyXG4gICAgLy8gbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1N0YXR1cyBmdW5jdGlvbjogJywgeyB2YWx1ZTogZ2V0U3RhdHVzTGFtYmRhLmZ1bmN0aW9uTmFtZSB9KTtcclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdUYWJsZSBuYW1lOiAnLCB7IHZhbHVlOiBteVRhYmxlLnRhYmxlTmFtZSB9KTtcclxuXHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnU3RhdGUgbWFjaGluZTogJywgeyB2YWx1ZTogbXlTdGF0ZU1hY2hpbmUuc3RhdGVNYWNoaW5lTmFtZSB9KTtcclxuICB9XHJcbn1cclxuXHJcbmNvbnN0IGFwcCA9IG5ldyBjZGsuQXBwKCk7XHJcbm5ldyBJbWFnZVJlY29nU3RhY2soYXBwLCAnSW1hZ2VSZWNvZ1N0YWNrJyk7Il19