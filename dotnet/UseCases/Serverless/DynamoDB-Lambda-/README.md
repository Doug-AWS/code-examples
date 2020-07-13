# Serverless code examples using the AWS SDK for .NET

This folder contains a use case that uses Amazon DynamoDB to store questions and answers.
You pick a question and decide which answer is correct.

Interaction with the database is through AWS Lambda function calls.

This interaction is exposed through Amazon API Gateway,
which uses Amazon Cognito to specify who can call these APIs.

It also uses CloudWatch and AWS X-Ray to monitor the operations.

Finally, it includes a React application to present the questions
and answers to users.

## Creating the application

To create the initial code for the app:

- Create a new folder.
  ```sh
  mkdir Serverless
  ```
- Navigate to that folder.
  ```sh
  cd Serverless
  ```
- Run the following command to create a new console application:
  ```sh
  dotnet new console --name DynamoDB-Lambda-
  ```
- Navigate to that folder.
  ```sh
  cd DynamoDB-Lambda-
  ```
- Run the following commands to add the core .NET and DynamoDB,
  and JSON serializer
  NuGet packages to the application.
  We'll add more packages as we need them.
  ```sh
  dotnet add package AWSSDK.Core
  dotnet add package AWSSDK.DYNAMODBv2
  dotnet add package Newtonsoft.Json
  ```
- Use the following command to see these references in your project.
  ```sh
  dotnet list package
  ```
- Create a *config.json* file,
  with the following content,
  that we use to save the names of our resources.
  Feel free to change the value, but not the key.
  We'll add more entries to the configuration file as we create our application.
  ```sh
  {
    "Table": "MyDynamoDBTable"
  }
  ```
- Open *Program.cs* and add the following class to store our configuration data.
  We'll modify the class as we go to add more configuration settings.
  ```cs
  public class MyConfig
  {
      string Table;

      MyConfig(string table)
      {
          Table = table;
      }
  }
  ```
- Replace the single line of code in the **main** routine with the following:
  ```cs
  