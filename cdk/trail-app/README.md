# Creating a sign-up for a hike app

This AWS CDK app creates the resources for an online app
where hikers can sign up for a hike.

## Resources

The app creates the following resources:

- An Amazon S3 bucket containing static web resources,
  including HTML, CSS, JavaScript, and image files
- An Amazon Cognito (user?) pool
- An Amazon API Gateway containing REST API interfaces to the Lambda functions
- AWS Lambda functions to:
  - Create a new hike (admin)
  - Add a new user (admin)
  - Make a user an admin (admin)
  - Show my hikes
  - Show upcoming hikes
  - Sign up for an upcoming hike
  - Unenroll from a hike
  - Delete a hike and notify all who signed up (admin)
- An Amazon DynamoDB table

All information, such as which hikes are available to join,
is obtained through API Gateway REST calls to a Lambda function.
The Lambda function sends data to and receives data from the DynamoDB table to persist data. 
Amazon Cognito provides user management and authentication functions to secure the backend API.
The initial admin is defined in **cdk.json**.

## User workflow

1. The user opens the web page
1. They are authenticated using Cognito
1. They are presented with the upcoming hikes to which they have registered
   and upcoming hikes.
1. If they want to join an upcoming hike or drop out of a hike,
   they select that hike from the list.
1. They confirm their decision or cancel.
1. When they are finished, they log off.

I'll add some screenshots of each of these decision points as I develop the app.

## Admin workflow

1. The user invokes the app
1. They are authenticated using Cognito
1. Since they have admin privileges, 
   the app asks them whether they want to perform admin actions
1. If they respond 'n',
   they are logged in as a regular user and go to step #3 in the previous workflow.
1. If they respond 'y', they are presented with a choice:
   a. Add a new hike
   a. delete an upcoming hike
1. If they chose Add a new hike, they enter information about that hike
   (date, time, location, difficulty, etc.)
1. If they chose Delete a hike, it asks for confirmation, 
   then deletes the hike and sends a notification to all who signed up.
   
I'll add some screenshots of each of these decision points as I develop the app.

## ???

What else?

## Commands

 * `npm run build`   compile typescript to js
 * `npm run watch`   watch for changes and compile
 * `npm run test`    perform the jest unit tests
 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template
