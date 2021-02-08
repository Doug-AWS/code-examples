import * as cdk from '@aws-cdk/core';
// import * as codebuild from '@aws-cdk/aws-codebuild';
// import * as amplify from '@aws-cdk/aws-amplify';
import * as s3 from '@aws-cdk/aws-s3';
import * as cloudtrail from '@aws-cdk/aws-cloudtrail';
import * as events from '@aws-cdk/aws-events';
import * as targets from '@aws-cdk/aws-events-targets';
// import * as cloudwatch from '@aws-cdk/aws-cloudwatch';

// import * as nots from '@aws-cdk/aws-s3-notifications';
import * as iam from '@aws-cdk/aws-iam';
import * as lambda from '@aws-cdk/aws-lambda';
import * as dynamodb from '@aws-cdk/aws-dynamodb';
import * as sfn from '@aws-cdk/aws-stepfunctions';
import * as tasks from '@aws-cdk/aws-stepfunctions-tasks';

export class ImageRecogStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create Amazon Simple Storage Service (Amazon S3) bucket
    const myBucket = new s3.Bucket(this, 'doc-example-bucket');

    // Create CloudTrail trail to watch for events from bucket
    const myTrail = new cloudtrail.Trail(this, 'doc-example-trail');
    // Add an event selector to the trail so that
    // JPG or PNG files with 'uploads/' prefix
    // added to bucket are detected
    myTrail.addS3EventSelector([{
      bucket: myBucket,
      objectPrefix: 'uploads/',
    },]);

    // Create CloudWatch Events rule
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


    // Create Lambda function for state machine task to execute
    // Lambda function that receives notifications from Amazon S3 (ItemUpload)
    // and writes it to DynamoDB table
    const getMetadataFunction = new lambda.Function(this, 'doc-example-get-metadata', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: new lambda.AssetCode('src/get_metadata'), // Go source file is (relative to cdk.json): src/get_metadata/main.go
      environment: {
        tableName: myTable.tableName,
      },
    });

    // Create Lambda function to get status of uploaded data for state machine
    const getStatusLambda = new lambda.Function(this, 'doc-example-get-status', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: new lambda.AssetCode('src/get_status'), // Go source file is (relative to cdk.json): src/get_status/main.go
      environment: {
        tableName: myTable.tableName,
      },
    });

    // Create Step Functions state machine
    // First create a task for the state machine to execute
    // We'll start with one, that just echoes the bucket and key
    const submitJob = new tasks.LambdaInvoke(this, 'Submit Job', {
      lambdaFunction: getMetadataFunction,
      // Lambda's result is in the attribute `Payload`
      outputPath: '$.Payload',
    });

    const waitX = new sfn.Wait(this, 'Wait X Seconds', {
      time: sfn.WaitTime.secondsPath('$.waitSeconds'),
    });

    const getStatus = new tasks.LambdaInvoke(this, 'Get Job Status', {
      lambdaFunction: getStatusLambda,
      // Pass just the field named "guid" into the Lambda, put the
      // Lambda's result in a field called "status" in the response
      inputPath: '$.guid',
      outputPath: '$.Payload',
    });

    const jobFailed = new sfn.Fail(this, 'Job Failed', {
      cause: 'AWS Batch Job Failed',
      error: 'DescribeJob returned FAILED',
    });

    const finalStatus = new tasks.LambdaInvoke(this, 'Get Final Job Status', {
      lambdaFunction: getStatusLambda,
      // Use "guid" field as input
      inputPath: '$.guid',
      outputPath: '$.Payload',
    });

    // Create state machine with one task, submitJob
    const definition = submitJob
      .next(waitX)
      .next(getStatus)
      .next(new sfn.Choice(this, 'Job Complete?')
        // Look at the "status" field
        .when(sfn.Condition.stringEquals('$.status', 'FAILED'), jobFailed)
        .when(sfn.Condition.stringEquals('$.status', 'SUCCEEDED'), finalStatus)
        .otherwise(waitX));

    const duration = require('duration');

    const myStateMachine = new sfn.StateMachine(this, 'StateMachine', {
      definition,
      timeout: duration.minutes(5)
    });




    // Send S3 events to Step Functions state machine
    rule.addTarget(new targets.SfnStateMachine(myStateMachine));



    // Create role for Lambda function to call Step Functions
    const stepFuncRole = new iam.Role(this, 'doc-example-stepfunc-role', {
      roleName: 'doc-example-stepfunc',
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com')
    });

    // Let Lambda function call these Step Function operations
    stepFuncRole.addToPolicy(new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      resources: [],
    }))

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

    // Let Lambda call these 

    dynamoDbRole.addManagedPolicy(
      iam.ManagedPolicy.fromAwsManagedPolicyName(
        'service-role/AWSLambdaBasicExecutionRole'));

    // Configure Amazon S3 bucket to send notification events to step functions.
    // myBucket.addEventNotification(s3.EventType.OBJECT_CREATED, new nots.LambdaDestination(getMetadataFunction));


    // Display info about the resources.
    // You can see this information at any time by running:
    //   aws cloudformation describe-stacks --stack-name GoLambdaCdkStack --query Stacks[0].Outputs --output text
    new cdk.CfnOutput(this, 'Bucket name: ', { value: myBucket.bucketName });
    new cdk.CfnOutput(this, 'S3 function name: ', { value: getMetadataFunction.functionName });
    new cdk.CfnOutput(this, 'S3 function CloudWatch log group: ', { value: getMetadataFunction.logGroup.logGroupName });
  }
}
