#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ImageRecogStack = void 0;
require("source-map-support/register");
const cdk = require("@aws-cdk/core");
const s3 = require("@aws-cdk/aws-s3");
const cloudtrail = require("@aws-cdk/aws-cloudtrail");
const events = require("@aws-cdk/aws-events");
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
        myTable.grantWriteData(saveMetadataFunction.grantPrincipal);
        myBucket.grantRead(saveMetadataFunction.grantPrincipal);
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
        myTable.grantWriteData(saveObjectDataFunction.grantPrincipal);
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
        myBucket.grantReadWrite(createThumbnailFunction.grantPrincipal);
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
            .next(createThumbnailJob);
        const myStateMachine = new sfn.StateMachine(this, 'StateMachine', {
            definition,
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
        //   aws cloudformation describe-stacks --stack-name ImageRecogStack --query Stacks[0].Outputs --output text
        new cdk.CfnOutput(this, 'Bucket name: ', { value: myBucket.bucketName });
        new cdk.CfnOutput(this, 'Save metadata function: ', { value: saveMetadataFunction.functionName });
        new cdk.CfnOutput(this, 'Save object data function: ', { value: saveObjectDataFunction.functionName });
        new cdk.CfnOutput(this, 'Create thumbnail function: ', { value: createThumbnailFunction.functionName });
        new cdk.CfnOutput(this, 'CloudTrail trail ARN: ', { value: myTrail.trailArn });
        new cdk.CfnOutput(this, 'Save ELIF data function CloudWatch log group: ', { value: saveMetadataFunction.logGroup.logGroupName });
        new cdk.CfnOutput(this, 'Save Rekognition function CloudWatch log group: ', { value: saveObjectDataFunction.logGroup.logGroupName });
        new cdk.CfnOutput(this, 'Create thumbnail function CloudWatch log group: ', { value: createThumbnailFunction.logGroup.logGroupName });
        new cdk.CfnOutput(this, 'Table name: ', { value: myTable.tableName });
        new cdk.CfnOutput(this, 'State machine: ', { value: myStateMachine.stateMachineName });
    }
}
exports.ImageRecogStack = ImageRecogStack;
const app = new cdk.App();
new ImageRecogStack(app, 'ImageRecogStack');
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoic2V0dXAuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi9zZXR1cC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7O0FBQ0EsdUNBQXFDO0FBRXJDLHFDQUFxQztBQUNyQyxzQ0FBc0M7QUFDdEMsc0RBQXNEO0FBQ3RELDhDQUE4QztBQUU5Qyw4Q0FBOEM7QUFDOUMsa0RBQWtEO0FBQ2xELGtEQUFrRDtBQUNsRCwwREFBMEQ7QUFHMUQsTUFBYSxlQUFnQixTQUFRLEdBQUcsQ0FBQyxLQUFLO0lBQzVDLFlBQVksS0FBb0IsRUFBRSxFQUFVLEVBQUUsS0FBc0I7UUFDbEUsS0FBSyxDQUFDLEtBQUssRUFBRSxFQUFFLEVBQUUsS0FBSyxDQUFDLENBQUM7UUFFeEI7Ozs7OztXQU1HO1FBRUgsMERBQTBEO1FBQzFELE1BQU0sUUFBUSxHQUFHLElBQUksRUFBRSxDQUFDLE1BQU0sQ0FBQyxJQUFJLEVBQUUsb0JBQW9CLENBQUMsQ0FBQztRQUUzRCwrQ0FBK0M7UUFDL0MsTUFBTSxPQUFPLEdBQUcsSUFBSSxVQUFVLENBQUMsS0FBSyxDQUFDLElBQUksRUFBRSxtQkFBbUIsQ0FBQyxDQUFDO1FBQ2hFLDZDQUE2QztRQUM3QywwQ0FBMEM7UUFDMUMsK0JBQStCO1FBQy9CLE9BQU8sQ0FBQyxrQkFBa0IsQ0FBQyxDQUFDO2dCQUMxQixNQUFNLEVBQUUsUUFBUTtnQkFDaEIsWUFBWSxFQUFFLFVBQVU7YUFDekIsRUFBRSxDQUFDLENBQUM7UUFFTCxxQkFBcUI7UUFDckIsTUFBTSxJQUFJLEdBQUcsSUFBSSxNQUFNLENBQUMsSUFBSSxDQUFDLElBQUksRUFBRSxNQUFNLEVBQUU7WUFDekMsWUFBWSxFQUFFO2dCQUNaLE1BQU0sRUFBRSxDQUFDLFFBQVEsQ0FBQzthQUNuQjtTQUNGLENBQUMsQ0FBQztRQUVILGtFQUFrRTtRQUNsRSw4REFBOEQ7UUFDOUQsa0RBQWtEO1FBQ2xELE1BQU0sT0FBTyxHQUFHLElBQUksUUFBUSxDQUFDLEtBQUssQ0FBQyxJQUFJ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QUFDO1FBRS9FLElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsZ0RBQWdELEVBQUUsRUFBRSxLQUFLLEVBQUUsb0JBQW9CLENBQUMsUUFBUSxDQUFDLFlBQVksRUFBRSxDQUFDLENBQUM7UUFDakksSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSxrREFBa0QsRUFBRSxFQUFFLEtBQUssRUFBRSxzQkFBc0IsQ0FBQyxRQUFRLENBQUMsWUFBWSxFQUFFLENBQUMsQ0FBQztRQUNySSxJQUFJLEdBQUcsQ0FBQyxTQUFTLENBQUMsSUFBSSxFQUFFLGtEQUFrRCxFQUFFLEVBQUUsS0FBSyxFQUFFLHVCQUF1QixDQUFDLFFBQVEsQ0FBQyxZQUFZLEVBQUUsQ0FBQyxDQUFDO1FBRXRJLElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsY0FBYyxFQUFFLEVBQUUsS0FBSyxFQUFFLE9BQU8sQ0FBQyxTQUFTLEVBQUUsQ0FBQyxDQUFDO1FBQ3RFLElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUsaUJBQWlCLEVBQUUsRUFBRSxLQUFLLEVBQUUsY0FBYyxDQUFDLGdCQUFnQixFQUFFLENBQUMsQ0FBQztJQUN6RixDQUFDO0NBQ0Y7QUF0TUQsMENBc01DO0FBRUQsTUFBTSxHQUFHLEdBQUcsSUFBSSxHQUFHLENBQUMsR0FBRyxFQUFFLENBQUM7QUFDMUIsSUFBSSxlQUFlLENBQUMsR0FBRyxFQUFFLGlCQUFpQixDQUFDLENBQUMiLCJzb3VyY2VzQ29udGVudCI6WyIjIS91c3IvYmluL2VudiBub2RlXHJcbmltcG9ydCAnc291cmNlLW1hcC1zdXBwb3J0L3JlZ2lzdGVyJztcclxuXHJcbmltcG9ydCAqIGFzIGNkayBmcm9tICdAYXdzLWNkay9jb3JlJztcclxuaW1wb3J0ICogYXMgczMgZnJvbSAnQGF3cy1jZGsvYXdzLXMzJztcclxuaW1wb3J0ICogYXMgY2xvdWR0cmFpbCBmcm9tICdAYXdzLWNkay9hd3MtY2xvdWR0cmFpbCc7XHJcbmltcG9ydCAqIGFzIGV2ZW50cyBmcm9tICdAYXdzLWNkay9hd3MtZXZlbnRzJztcclxuaW1wb3J0ICogYXMgaWFtIGZyb20gJ0Bhd3MtY2RrL2F3cy1pYW0nO1xyXG5pbXBvcnQgKiBhcyBsYW1iZGEgZnJvbSAnQGF3cy1jZGsvYXdzLWxhbWJkYSc7XHJcbmltcG9ydCAqIGFzIGR5bmFtb2RiIGZyb20gJ0Bhd3MtY2RrL2F3cy1keW5hbW9kYic7XHJcbmltcG9ydCAqIGFzIHNmbiBmcm9tICdAYXdzLWNkay9hd3Mtc3RlcGZ1bmN0aW9ucyc7XHJcbmltcG9ydCAqIGFzIHRhc2tzIGZyb20gJ0Bhd3MtY2RrL2F3cy1zdGVwZnVuY3Rpb25zLXRhc2tzJztcclxuaW1wb3J0IHsgRWZmZWN0IH0gZnJvbSAnQGF3cy1jZGsvYXdzLWlhbSc7XHJcblxyXG5leHBvcnQgY2xhc3MgSW1hZ2VSZWNvZ1N0YWNrIGV4dGVuZHMgY2RrLlN0YWNrIHtcclxuICBjb25zdHJ1Y3RvcihzY29wZTogY2RrLkNvbnN0cnVjdCwgaWQ6IHN0cmluZywgcHJvcHM/OiBjZGsuU3RhY2tQcm9wcykge1xyXG4gICAgc3VwZXIoc2NvcGUsIGlkLCBwcm9wcyk7XHJcblxyXG4gICAgLyogVXNlIGJ1Y2tldCBldmVudCB0byBleGVjdXRlIGEgc3RlcCBmdW5jdGlvbiB3aGVuIGFuIGl0ZW0gdXBsb2FkZWQgdG8gYSBidWNrZXRcclxuICAgICAqICAgaHR0cHM6Ly9kb2NzLmF3cy5hbWF6b24uY29tL3N0ZXAtZnVuY3Rpb25zL2xhdGVzdC9kZy90dXRvcmlhbC1jbG91ZHdhdGNoLWV2ZW50cy1zMy5odG1sXHJcbiAgICAgKlxyXG4gICAgICogMTogQ3JlYXRlIGEgYnVja2V0IChBbWF6b24gUzMpXHJcbiAgICAgKiAyOiBDcmVhdGUgYSB0cmFpbCAoQVdTIENsb3VkVHJhaWwpXHJcbiAgICAgKiAzOiBDcmVhdGUgYW4gZXZlbnRzIHJ1bGUgKEFXUyBDbG91ZFdhdGNoIEV2ZW50cylcclxuICAgICAqL1xyXG5cclxuICAgIC8vIENyZWF0ZSBBbWF6b24gU2ltcGxlIFN0b3JhZ2UgU2VydmljZSAoQW1hem9uIFMzKSBidWNrZXRcclxuICAgIGNvbnN0IG15QnVja2V0ID0gbmV3IHMzLkJ1Y2tldCh0aGlzLCAnZG9jLWV4YW1wbGUtYnVja2V0Jyk7XHJcblxyXG4gICAgLy8gQ3JlYXRlIHRyYWlsIHRvIHdhdGNoIGZvciBldmVudHMgZnJvbSBidWNrZXRcclxuICAgIGNvbnN0IG15VHJhaWwgPSBuZXcgY2xvdWR0cmFpbC5UcmFpbCh0aGlzLCAnZG9jLWV4YW1wbGUtdHJhaWwnKTtcclxuICAgIC8vIEFkZCBhbiBldmVudCBzZWxlY3RvciB0byB0aGUgdHJhaWwgc28gdGhhdFxyXG4gICAgLy8gSlBHIG9yIFBORyBmaWxlcyB3aXRoICd1cGxvYWRzLycgcHJlZml4XHJcbiAgICAvLyBhZGRlZCB0byBidWNrZXQgYXJlIGRldGVjdGVkXHJcbiAgICBteVRyYWlsLmFkZFMzRXZlbnRTZWxlY3Rvcihbe1xyXG4gICAgICBidWNrZXQ6IG15QnVja2V0LFxyXG4gICAgICBvYmplY3RQcmVmaXg6ICd1cGxvYWRzLycsXHJcbiAgICB9LF0pO1xyXG5cclxuICAgIC8vIENyZWF0ZSBldmVudHMgcnVsZVxyXG4gICAgY29uc3QgcnVsZSA9IG5ldyBldmVudHMuUnVsZSh0aGlzLCAncnVsZScsIHtcclxuICAgICAgZXZlbnRQYXR0ZXJuOiB7XHJcbiAgICAgICAgc291cmNlOiBbJ2F3cy5zMyddLFxyXG4gICAgICB9LFxyXG4gICAgfSk7XHJcblxyXG4gICAgLy8gQ3JlYXRlIER5bmFtb0RCIHRhYmxlIGZvciBMYW1iZGEgZnVuY3Rpb24gdG8gcGVyc2lzdCBpbWFnZSBpbmZvXHJcbiAgICAvLyBDcmVhdGUgQW1hem9uIER5bmFtb0RCIHRhYmxlIHdpdGggcHJpbWFyeSBrZXkgcGF0aCAoc3RyaW5nKVxyXG4gICAgLy8gdGhhdCB3aWxsIGJlIHNvbWV0aGluZyBsaWtlIHVwbG9hZHMvbXlQaG90by5qcGdcclxuICAgIGNvbnN0IG15VGFibGUgPSBuZXcgZHluYW1vZGIuVGFibGUodGhpcywgJ2RvYy1leGFtcGxlLXRhYmxlJywge1xyXG4gICAgICBwYXJ0aXRpb25LZXk6IHsgbmFtZTogJ3BhdGgnLCB0eXBlOiBkeW5hbW9kYi5BdHRyaWJ1dGVUeXBlLlNUUklORyB9LFxyXG4gICAgICBzdHJlYW06IGR5bmFtb2RiLlN0cmVhbVZpZXdUeXBlLk5FV19JTUFHRSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8qIFxyXG4gICAgICogRGVmaW5lIExhbWJkYSBmdW5jdGlvbnMgdG86XHJcbiAgICAgKiAxLiBBZGQgbWV0YWRhdGEgZnJvbSB0aGUgcGhvdG8gdG8gYSBEeW5hbW9kYiB0YWJsZS4gICAgIFxyXG4gICAgICogMi4gQ2FsbCBBbWF6b24gUmVrb2duaXRpb24gdG8gZGV0ZWN0IG9iamVjdHMgaW4gdGhlIGltYWdlIGZpbGUuXHJcbiAgICAgKiAzLiBHZW5lcmF0ZSBhIHRodW1ibmFpbCBhbmQgc3RvcmUgaXQgaW4gdGhlIFMzIGJ1Y2tldCB3aXRoIHRoZSAqKnJlc2l6ZWQvKiogcHJlZml4XHJcbiAgICAgKi9cclxuXHJcbiAgICAvLyBMYW1iZGEgZnVuY3Rpb24gdGhhdDpcclxuICAgIC8vIDEuIFJlY2VpdmVzIG5vdGlmaWNhdGlvbnMgZnJvbSBBbWF6b24gUzMgKEl0ZW1VcGxvYWQpXHJcbiAgICAvLyAyLiBHZXRzIG1ldGFkYXRhIGZyb20gdGhlIHBob3RvXHJcbiAgICAvLyAzLiBTYXZlcyB0aGUgbWV0YWRhdGEgaW4gYSBEeW5hbW9EQiB0YWJsZVxyXG4gICAgY29uc3Qgc2F2ZU1ldGFkYXRhRnVuY3Rpb24gPSBuZXcgbGFtYmRhLkZ1bmN0aW9uKHRoaXMsICdkb2MtZXhhbXBsZS1zYXZlLW1ldGFkYXRhJywge1xyXG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5HT18xX1gsXHJcbiAgICAgIGhhbmRsZXI6ICdtYWluJyxcclxuICAgICAgY29kZTogbmV3IGxhbWJkYS5Bc3NldENvZGUoJ3NyYy9zYXZlX21ldGFkYXRhJyksIC8vIEdvIHNvdXJjZSBmaWxlIGlzIChyZWxhdGl2ZSB0byBjZGsuanNvbik6IHNyYy9zYXZlX21ldGFkYXRhL21haW4uZ29cclxuICAgICAgZW52aXJvbm1lbnQ6IHtcclxuICAgICAgICB0YWJsZU5hbWU6IG15VGFibGUudGFibGVOYW1lLFxyXG4gICAgICB9LFxyXG4gICAgfSk7XHJcblxyXG4gICAgLy8gQWRkIHBvbGljeSB0byBMYW1iZGEgZnVuY3Rpb24gc28gaXQgY2FuIGNhbGxcclxuICAgIC8vIEdldE9iamVjdCBvbiBidWNrZXQgYW5kIFB1dEl0ZW0gb24gdGFibGUuXHJcbiAgICAvKlxyXG4gICAgY29uc3QgczNQb2xpY3kgPSBuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIHNpZDogXCJkb2NleGFtcGxlczNkYnN0YXRlbWVudFwiLFxyXG4gICAgICBhY3Rpb25zOiBbXCJzMzpHZXRPYmplY3RcIiwgXCJkeW5hbW9kYjpQdXRJdGVtXCJdLFxyXG4gICAgICBlZmZlY3Q6IEVmZmVjdC5BTExPVyxcclxuICAgICAgcmVzb3VyY2VzOiBbbXlCdWNrZXQuYnVja2V0QXJuICsgXCIvKlwiLCBteVRhYmxlLnRhYmxlQXJuICsgXCIvKlwiXSxcclxuICAgIH0pXHJcblxyXG4gICAgc2F2ZU1ldGFkYXRhRnVuY3Rpb24ucm9sZT8uYWRkVG9QcmluY2lwYWxQb2xpY3koczNQb2xpY3kpXHJcbiAgICAqL1xyXG5cclxuICAgIC8vIEdpdmUgTGFtYmRhIGZ1bmN0aW9uLCB3aGljaCBzYXZlIEVMSUYgZGF0YSwgd3JpdGUgYWNjZXNzIHRvIER5bmFtb0RCIHRhYmxlIGFuZCByZWFkIGFjY2VzcyB0byBTMyBidWNrZXRcclxuICAgIG15VGFibGUuZ3JhbnRXcml0ZURhdGEoc2F2ZU1ldGFkYXRhRnVuY3Rpb24uZ3JhbnRQcmluY2lwYWwpXHJcbiAgICBteUJ1Y2tldC5ncmFudFJlYWQoc2F2ZU1ldGFkYXRhRnVuY3Rpb24uZ3JhbnRQcmluY2lwYWwpXHJcblxyXG4gICAgLy8gTGFtYmRhIGZ1bmN0aW9uIHRoYXQ6XHJcbiAgICAvLyAxLiBDYWxscyBBbWF6b24gUmVrb2duaXRpb24gdG8gZGV0ZWN0IG9iamVjdHMgaW4gdGhlIGltYWdlIGZpbGVcclxuICAgIC8vIDIuIFNhdmVzIGluZm9ybWF0aW9uIGFib3V0IHRoZSBvYmplY3RzIGluIGEgRHluYW1vZGIgdGFibGVcclxuICAgIGNvbnN0IHNhdmVPYmplY3REYXRhRnVuY3Rpb24gPSBuZXcgbGFtYmRhLkZ1bmN0aW9uKHRoaXMsICdkb2MtZXhhbXBsZS1zYXZlLW9iamVjdC1kYXRhJywge1xyXG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5HT18xX1gsXHJcbiAgICAgIGhhbmRsZXI6ICdtYWluJyxcclxuICAgICAgY29kZTogbmV3IGxhbWJkYS5Bc3NldENvZGUoJ3NyYy9zYXZlX29iamVjdGRhdGEnKSwgLy8gR28gc291cmNlIGZpbGUgaXMgKHJlbGF0aXZlIHRvIGNkay5qc29uKTogc3JjL3NhdmVfb2JqZWN0ZGF0YS9tYWluLmdvXHJcbiAgICAgIGVudmlyb25tZW50OiB7XHJcbiAgICAgICAgdGFibGVOYW1lOiBteVRhYmxlLnRhYmxlTmFtZSxcclxuICAgICAgfSxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIEFkZCBwb2xpY3kgdG8gTGFtYmRhIGZ1bmN0aW9uIHNvIGl0IGNhbiBjYWxsXHJcbiAgICAvLyBVcGRhdGVJdGVtIG9uIHRhYmxlLlxyXG4gICAgLy8gRG8gd2UgbmVlZCB0byBhZGQgcmVrb2duaXRpb24uRGV0ZWN0TGFiZWxzPz8/XHJcbiAgICAvKlxyXG4gICAgY29uc3QgZGJQb2xpY3kgPSBuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIHNpZDogXCJkb2NleGFtcGxlZGJzdGF0ZW1lbnRcIixcclxuICAgICAgYWN0aW9uczogW1wiZHluYW1vZGI6VXBkYXRlSXRlbVwiXSxcclxuICAgICAgZWZmZWN0OiBFZmZlY3QuQUxMT1csXHJcbiAgICAgIHJlc291cmNlczogW215VGFibGUudGFibGVBcm4gKyBcIi8qXCJdLFxyXG4gICAgfSlcclxuXHJcbiAgICBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uLnJvbGU/LmFkZFRvUHJpbmNpcGFsUG9saWN5KGRiUG9saWN5KVxyXG4gICAgKi9cclxuXHJcbiAgICAvLyBHaXZlIExhbWJkYSBmdW5jdGlvbiwgd2hpY2ggc2F2ZXMgUmVrb2duaXRpb24gZGF0YSwgd3JpdGUgYWNjZXNzIHRvIER5bmFtb0RCIHRhYmxlXHJcbiAgICBteVRhYmxlLmdyYW50V3JpdGVEYXRhKHNhdmVPYmplY3REYXRhRnVuY3Rpb24uZ3JhbnRQcmluY2lwYWwpXHJcblxyXG4gICAgLy8gTGFtYmRhIGZ1bmN0aW9uIHRoYXQ6XHJcbiAgICAvLyAxLiBHZXRzIHRoZSBwaG90byBmcm9tIFMzXHJcbiAgICAvLyAyLiBDcmVhdGVzIGEgdGh1bWJuYWlsIG9mIHRoZSBwaG90b1xyXG4gICAgLy8gMy4gU2F2ZSB0aGUgcGhvdG8gYmFjayBpbnRvIFMzXHJcbiAgICBjb25zdCBjcmVhdGVUaHVtYm5haWxGdW5jdGlvbiA9IG5ldyBsYW1iZGEuRnVuY3Rpb24odGhpcywgJ2RvYy1leGFtcGxlLWNyZWF0ZS10aHVtYm5haWwnLCB7XHJcbiAgICAgIHJ1bnRpbWU6IGxhbWJkYS5SdW50aW1lLkdPXzFfWCxcclxuICAgICAgaGFuZGxlcjogJ21haW4nLFxyXG4gICAgICBjb2RlOiBuZXcgbGFtYmRhLkFzc2V0Q29kZSgnc3JjL2NyZWF0ZV90aHVtYm5haWwnKSwgLy8gR28gc291cmNlIGZpbGUgaXMgKHJlbGF0aXZlIHRvIGNkay5qc29uKTogc3JjL2NyZWF0ZV90aHVtYm5haWwvbWFpbi5nb1xyXG4gICAgfSk7XHJcblxyXG4gICAgLy8gQWRkIHBvbGljeSB0byBMYW1iZGEgZnVuY3Rpb24gc28gaXQgY2FuIGNhbGxcclxuICAgIC8vIEdldE9iamVjdCBhbmQgUHV0T2JqZWN0IG9uIGJ1Y2tldC5cclxuICAgIC8qXHJcbiAgICBjb25zdCBzMzJQb2xpY3kgPSBuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIHNpZDogXCJkb2NleGFtcGxlczNzdGF0ZW1lbnRcIixcclxuICAgICAgYWN0aW9uczogW1wiczM6R2V0T2JqZWN0XCIsIFwiczM6UHV0T2JqZWN0XCJdLFxyXG4gICAgICBlZmZlY3Q6IEVmZmVjdC5BTExPVyxcclxuICAgICAgcmVzb3VyY2VzOiBbbXlCdWNrZXQuYnVja2V0QXJuICsgXCIvKlwiXSxcclxuICAgIH0pXHJcblxyXG4gICAgY3JlYXRlVGh1bWJuYWlsRnVuY3Rpb24ucm9sZT8uYWRkVG9QcmluY2lwYWxQb2xpY3koczMyUG9saWN5KVxyXG4gICAgKi9cclxuXHJcbiAgICAvLyBHaXZlIExhbWJkYSBmdW5jdGlvbiwgd2hpY2ggY3JlYXRlcyBhIHRodW1ibmFpbCwgcmVhZC93cml0ZSBhY2Nlc3MgdG8gUzMgYnVja2V0XHJcbiAgICBteUJ1Y2tldC5ncmFudFJlYWRXcml0ZShjcmVhdGVUaHVtYm5haWxGdW5jdGlvbi5ncmFudFByaW5jaXBhbClcclxuXHJcbiAgICAvLyBGaXJzdCB0YXNrOiBzYXZlIG1ldGFkYXRhIGZyb20gcGhvdG8gaW4gUzMgYnVja2V0IHRvIER5bmFtb0RCIHRhYmxlXHJcbiAgICBjb25zdCBzYXZlTWV0YWRhdGFKb2IgPSBuZXcgdGFza3MuTGFtYmRhSW52b2tlKHRoaXMsICdTYXZlIE1ldGFkYXRhIEpvYicsIHtcclxuICAgICAgbGFtYmRhRnVuY3Rpb246IHNhdmVNZXRhZGF0YUZ1bmN0aW9uLFxyXG4gICAgICAvL2lucHV0UGF0aDogJyQnLCAvLyBFdmVudCBmcm9tIFMzIG5vdGlmaWNhdGlvbiAoZGVmYXVsdClcclxuICAgICAgb3V0cHV0UGF0aDogJyQuUGF5bG9hZCcsXHJcbiAgICB9KTtcclxuXHJcbiAgICAvLyBTZWNvbmQgdGFzazogc2F2ZSBpbWFnZSBkYXRhIGZyb20gUmVrb2duaXRpb24gdG8gRHluYW1vREIgdGFibGVcclxuICAgIGNvbnN0IHNhdmVPYmplY3REYXRhSm9iID0gbmV3IHRhc2tzLkxhbWJkYUludm9rZSh0aGlzLCAnU2F2ZSBPYmplY3QgRGF0YSBKb2InLCB7XHJcbiAgICAgIGxhbWJkYUZ1bmN0aW9uOiBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLlBheWxvYWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5QYXlsb2FkJyxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIEZpbmFsIHRhc2s6IGNyZWF0ZSB0aHVtYm5haWwgb2YgcGhvdG8gaW4gUzMgYnVja2V0XHJcbiAgICBjb25zdCBjcmVhdGVUaHVtYm5haWxKb2IgPSBuZXcgdGFza3MuTGFtYmRhSW52b2tlKHRoaXMsICdDcmVhdGUgVGh1bWJuYWlsIEpvYicsIHtcclxuICAgICAgbGFtYmRhRnVuY3Rpb246IGNyZWF0ZVRodW1ibmFpbEZ1bmN0aW9uLFxyXG4gICAgICBpbnB1dFBhdGg6ICckLlBheWxvYWQnLFxyXG4gICAgICBvdXRwdXRQYXRoOiAnJC5QYXlsb2FkJyxcclxuICAgIH0pO1xyXG5cclxuICAgIC8vIENyZWF0ZSBzdGF0ZSBtYWNoaW5lIHdpdGggb25lIHRhc2ssIHN1Ym1pdEpvYlxyXG4gICAgY29uc3QgZGVmaW5pdGlvbiA9IHNhdmVNZXRhZGF0YUpvYlxyXG4gICAgICAubmV4dChzYXZlT2JqZWN0RGF0YUpvYilcclxuICAgICAgLm5leHQoY3JlYXRlVGh1bWJuYWlsSm9iKVxyXG5cclxuICAgIGNvbnN0IG15U3RhdGVNYWNoaW5lID0gbmV3IHNmbi5TdGF0ZU1hY2hpbmUodGhpcywgJ1N0YXRlTWFjaGluZScsIHtcclxuICAgICAgZGVmaW5pdGlvbixcclxuICAgICAgdGltZW91dDogY2RrLkR1cmF0aW9uLm1pbnV0ZXMoNSksXHJcbiAgICB9KTtcclxuXHJcbiAgICAvLyBDcmVhdGUgcm9sZSBmb3IgTGFtYmRhIGZ1bmN0aW9uIHRvIGNhbGwgRHluYW1vREJcclxuICAgIC8qXHJcbiAgICBjb25zdCBkeW5hbW9EYlJvbGUgPSBuZXcgaWFtLlJvbGUodGhpcywgJ2RvYy1leGFtcGxlLWR5bmFtb2RiLXJvbGUnLCB7XHJcbiAgICAgIHJvbGVOYW1lOiAnZG9jLWV4YW1wbGUtZHluYW1vZGInLFxyXG4gICAgICBhc3N1bWVkQnk6IG5ldyBpYW0uU2VydmljZVByaW5jaXBhbCgnbGFtYmRhLmFtYXpvbmF3cy5jb20nKVxyXG4gICAgfSk7XHJcblxyXG4gICAgLy8gTGV0IExhbWJkYSBjYWxsIHRoZXNlIER5bmFtb0RCIG9wZXJhdGlvbnNcclxuICAgIGR5bmFtb0RiUm9sZS5hZGRUb1BvbGljeShuZXcgaWFtLlBvbGljeVN0YXRlbWVudCh7XHJcbiAgICAgIGVmZmVjdDogaWFtLkVmZmVjdC5BTExPVyxcclxuICAgICAgcmVzb3VyY2VzOiBbbXlUYWJsZS50YWJsZUFybl0sXHJcbiAgICAgIGFjdGlvbnM6IFtcclxuICAgICAgICAnZHluYW1vZGI6UHV0SXRlbSdcclxuICAgICAgXVxyXG4gICAgfSkpO1xyXG4gICAgKi9cclxuXHJcbiAgICAvLyBEaXNwbGF5IGluZm8gYWJvdXQgdGhlIHJlc291cmNlcy5cclxuICAgIC8vIFlvdSBjYW4gc2VlIHRoaXMgaW5mb3JtYXRpb24gYXQgYW55IHRpbWUgYnkgcnVubmluZzpcclxuICAgIC8vICAgYXdzIGNsb3VkZm9ybWF0aW9uIGRlc2NyaWJlLXN0YWNrcyAtLXN0YWNrLW5hbWUgSW1hZ2VSZWNvZ1N0YWNrIC0tcXVlcnkgU3RhY2tzWzBdLk91dHB1dHMgLS1vdXRwdXQgdGV4dFxyXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ0J1Y2tldCBuYW1lOiAnLCB7IHZhbHVlOiBteUJ1Y2tldC5idWNrZXROYW1lIH0pO1xyXG5cclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdTYXZlIG1ldGFkYXRhIGZ1bmN0aW9uOiAnLCB7IHZhbHVlOiBzYXZlTWV0YWRhdGFGdW5jdGlvbi5mdW5jdGlvbk5hbWUgfSk7XHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnU2F2ZSBvYmplY3QgZGF0YSBmdW5jdGlvbjogJywgeyB2YWx1ZTogc2F2ZU9iamVjdERhdGFGdW5jdGlvbi5mdW5jdGlvbk5hbWUgfSk7XHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnQ3JlYXRlIHRodW1ibmFpbCBmdW5jdGlvbjogJywgeyB2YWx1ZTogY3JlYXRlVGh1bWJuYWlsRnVuY3Rpb24uZnVuY3Rpb25OYW1lIH0pO1xyXG5cclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdDbG91ZFRyYWlsIHRyYWlsIEFSTjogJywgeyB2YWx1ZTogbXlUcmFpbC50cmFpbEFybiB9KTtcclxuXHJcbiAgICBuZXcgY2RrLkNmbk91dHB1dCh0aGlzLCAnU2F2ZSBFTElGIGRhdGEgZnVuY3Rpb24gQ2xvdWRXYXRjaCBsb2cgZ3JvdXA6ICcsIHsgdmFsdWU6IHNhdmVNZXRhZGF0YUZ1bmN0aW9uLmxvZ0dyb3VwLmxvZ0dyb3VwTmFtZSB9KTtcclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdTYXZlIFJla29nbml0aW9uIGZ1bmN0aW9uIENsb3VkV2F0Y2ggbG9nIGdyb3VwOiAnLCB7IHZhbHVlOiBzYXZlT2JqZWN0RGF0YUZ1bmN0aW9uLmxvZ0dyb3VwLmxvZ0dyb3VwTmFtZSB9KTtcclxuICAgIG5ldyBjZGsuQ2ZuT3V0cHV0KHRoaXMsICdDcmVhdGUgdGh1bWJuYWlsIGZ1bmN0aW9uIENsb3VkV2F0Y2ggbG9nIGdyb3VwOiAnLCB7IHZhbHVlOiBjcmVhdGVUaHVtYm5haWxGdW5jdGlvbi5sb2dHcm91cC5sb2dHcm91cE5hbWUgfSk7XHJcblxyXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1RhYmxlIG5hbWU6ICcsIHsgdmFsdWU6IG15VGFibGUudGFibGVOYW1lIH0pO1xyXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1N0YXRlIG1hY2hpbmU6ICcsIHsgdmFsdWU6IG15U3RhdGVNYWNoaW5lLnN0YXRlTWFjaGluZU5hbWUgfSk7XHJcbiAgfVxyXG59XHJcblxyXG5jb25zdCBhcHAgPSBuZXcgY2RrLkFwcCgpO1xyXG5uZXcgSW1hZ2VSZWNvZ1N0YWNrKGFwcCwgJ0ltYWdlUmVjb2dTdGFjaycpO1xyXG4iXX0=