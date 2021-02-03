import * as cdk from '@aws-cdk/core';
import * as codebuild from '@aws-cdk/aws-codebuild';
import * as amplify from '@aws-cdk/aws-amplify';
import * as s3 from '@aws-cdk/aws-s3';
import * as nots from '@aws-cdk/aws-s3-notifications';
import * as iam from '@aws-cdk/aws-iam';
import * as lambda from '@aws-cdk/aws-lambda';
import * as dynamodb from '@aws-cdk/aws-dynamodb';
import * as sfn from '@aws-cdk/aws-stepfunctions';
import * as tasks from '@aws-cdk/aws-stepfunctions-tasks';

export class ImageRecogStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const amplifyApp = new amplify.App(this, 'MyApp', {
      sourceCodeProvider: new amplify.GitHubSourceCodeProvider({
        owner: '<user>',
        repository: '<repo>',
        oauthToken: cdk.SecretValue.secretsManager('my-github-token')
      }),
      buildSpec: codebuild.BuildSpec.fromObject({ // Alternatively add a `amplify.yml` to the repo
        version: '1.0',
        frontend: {
          phases: {
            preBuild: {
              commands: [
                'yarn'
              ]
            },
            build: {
              commands: [
                'yarn build'
              ]
            }
          },
          artifacts: {
            baseDirectory: 'public',
            files: '**/*'
          }
        }
      })
    });
        
  

    // Create Amazon DynamoDB table with primary key id (string)
    const myTable = new dynamodb.Table(this, 'doc-example-table', {
      partitionKey: { name: 'id', type: dynamodb.AttributeType.STRING },
      stream: dynamodb.StreamViewType.NEW_IMAGE,
    });

    // Lambda function that receives notifications from Amazon S3 (ItemUpload)
    // and writes it to DynamoDB table
    const getMetadataFunction = new lambda.Function(this, 'doc-example-get-metadata', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: new lambda.AssetCode('src/get_metadata'), // Go source file is (relative to cdk.json): src/get_metadata/main.go
      environment: {
        tableName: myTable.tableName,
      }
    });

    // Create Amazon Simple Storage Service (Amazon S3) bucket
    const myBucket = new s3.Bucket(this, "doc-example-bucket");

    // Create Step Functions
    // We'll start with one, that just echoes the bucket and key
    const submitJob = new tasks.LambdaInvoke(this, 'Submit Job', {
      lambdaFunction: submitLambda,
      // Lambda's result is in the attribute `Payload`
      outputPath: '$.Payload',
    });


    

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
            "service-role/AWSLambdaBasicExecutionRole"));

    // Configure Amazon S3 bucket to send notification events to step functions.
    // myBucket.addEventNotification(s3.EventType.OBJECT_CREATED, new nots.LambdaDestination(getMetadataFunction));
    

    // Display info about the resources.
      // You can see this information at any time by running:
      //   aws cloudformation describe-stacks --stack-name GoLambdaCdkStack --query Stacks[0].Outputs --output text
      new cdk.CfnOutput(this, 'Bucket name: ', {value: myBucket.bucketName});
      new cdk.CfnOutput(this, 'S3 function name: ', {value: getMetadataFunction.functionName});
      new cdk.CfnOutput(this, 'S3 function CloudWatch log group: ', {value: getMetadataFunction.logGroup.logGroupName});
  }
}
