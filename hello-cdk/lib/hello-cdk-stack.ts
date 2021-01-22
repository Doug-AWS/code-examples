import * as cdk from '@aws-cdk/core';
import * as cfn from '@aws-cdk/aws-cloudformation';
import * as lambda from '@aws-cdk/aws-lambda';
import * as s3 from '@aws-cdk/aws-s3';
import * as nots from '@aws-cdk/aws-s3-notifications';
import * as path from 'path';
import { EventType } from '@aws-cdk/aws-s3';
import { CfnOutput } from '@aws-cdk/core';

export class HelloCdkStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create DynamoDB table with primary key id (string)

    // Create S3 bucket 
    const myBucket = new s3.Bucket(this, "MyBucket",);

    // Create Lambda function:
    
    // S3 Lambda function:
    const myS3Function = new lambda.Function(this, 'MyS3Function', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'handler', // call handler() in main.go
      code: new lambda.AssetCode('src/s3'), // Go source file is (relative to cdk.json): src/s3/main.go
    });

    myBucket.addEventNotification(EventType.OBJECT_CREATED, new nots.LambdaDestination(myS3Function))
        
    // Barf out info about the resources
    new CfnOutput(this, 'Bucket name:    ', {value: myBucket.bucketName});
    new CfnOutput(this, 'Function name:  ', {value: myS3Function.functionName});
    new CfnOutput(this, 'CloudWatch log: ', {value: myS3Function.logGroup.logGroupName});
  }
}
