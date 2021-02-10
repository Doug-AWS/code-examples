# Recognizing images serverlessly

This AWS CDK project creates an serverless workflow using AWS Step Functions, 
AWS Lambda, Amazon S3, Amazon DynamoDB and Amazon Rekognition.

Once the CDK app has been deployed, the user opens a Go app that starts the
workflow.

## Workflow

This workflow is as follows:

- The user calls the Go app with the fully-qualified path to a JPG or PNG photo.
- The app uploads the photo to an S3 bucket with the **upload/** prefix.
- The upload event triggers a Step Function workflow with the following steps as Lambda functions:
  1. Adds metadata from the photo to a Dynamodb table.     
  1. Calls Amazon Rekognition to detect objects in the image file.
  1. Generates a thumbnail and stores it in the S3 bucket with the **resized/** prefix

