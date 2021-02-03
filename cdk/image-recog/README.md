# Recognizing images serverlessly

This AWS CDK project creates an serverless workflow using AWS Step Functions, 
AWS Lambda, Amazon S3, Amazon DynamoDB and Amazon Rekognition.

Once the CDK app has been deployed, the user opens a Go app that starts the
workflow.

## Workflow

This workflow is as follows:

- The user opens the Go app.
- The app gets the fully-qualified path to a JPG or PNG photo.
- The app uploads the photo to an S3 bucket with the **upload/** prefix.
- The upload event triggers a Lambda function,
  which calls the following AWS Step Functions if the file is a JPG or PNG:
  1. Confirms that the photo is a JPG or PNG.
     If not, it logs an error and returns.
  1. Gets the photo from S3
  1. Extracts image metadata (format, EXIF data, size, etc.)   
  1. Calls:
     - Amazon Rekognition to detect objects in the image file. 
       If detected, store the tags in a DynamoDB table
     - Generates a thumbnail and stores it in the S3 bucket with the **resized/** prefix

