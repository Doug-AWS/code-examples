# Amazon DynamoDB code examples in C#

This folder contains code examples for moving from SQL to NoSQL, specifically Amazon DynamoDB,
as described in the Amazon DynamoDB Developer Guide at
[From SQL to NoSQL](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/SQLtoNoSQL.html).

All of these code examples are written in C#, using the beta version of the AWS SDK for .NET.
Getting the 3.5 bits is straightforward using the command line from the same folder as your ```.csproj``` file.
For example, to load the beta version of the Amazon DynamoDB bits:

```
dotnet add package AWSSDK.DynamoDBv2 --version 3.5.0-beta
```

## Using asynch/await

Read the 
[Migrating to Version 3.5 of the AWS SDK for .NET](https://docs.aws.amazon.com/sdk-for-net/v3/developer-guide/net-dg-v35.html) 
topic for details.

## Before you write any code

Amazon DynamoDB supports the following data types,
so you might have to create a new data model:

- Scalar Types

  A scalar type can represent exactly one value.
  The scalar types are number, string, binary, Boolean, and null.

- Document Types
 
  A document type can represent a complex structure with nested attributes,
  such as you would find in a JSON document.
  The document types are list and map.

- Set Types

  A set type can represent multiple scalar values.
  The set types are string set, number set, and binary set.
  
Figure out how you want to access your data.
Many, if not most, stored procedures can be implemented using
[AWS Lambda](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.Lambda.BestPracticesWithDynamoDB.html).

Determine the type of primary key you want:

- Partition key, which is a unique identifier for the item in the table.
  If you use a partition key, every key must be unique.
  
- Partition key and sort key.
  In this case, you need not have a unique partition key,
  however, the combination of partition key and sort key must be unique.
  The table we create in these code examples will contain information about songs,
  so the partition key will be a hash of the artist's name,
  and the sort key will be a string containing the title of the song.
  
Consider creating seconday indices.
These give you additional flexibility when querying the table.
Remember, Amazon DynamoDB does not use SQL.

We'll show you how to create all of these when you create a table,
and how to use them when you access a table.

## General code pattern

It's important that you understand the new asynch/await programming model in the
[AWS SDK for .NET](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide).

These code examples use the following NuGet packages:

- AWSSDK.Core, v3.5.0-beta
- AWSSDK,DynamoDBv2, v3.5.0-beta

All of the following sections contain a static method to implement the stated objective.
To reduce the amount of code in each section,
each uses the following template.
*NOTE*: anything in ALL CAPS (API, RESOURCE) is a placeholder.

```
using System;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class Program
    {
        // Static method goes here
        /* static async Task<APIResponse> DoSomethingAsync(IAmazonDynamoDB client, string RESOURCE, ...)
           {
               var response = await client.APIAsync(...
               ...
               return response;

           }
        */

        static void Main(string[] args)
        {
            string table = "";
            string artist = "";
            string title = "";            

            int i = 0;
            while (i < args.Length)
            {
                switch (args[i])
                {
                    case "-t":
                        i++;
                        table = args[i];
                        break;

                    case "-a":
                        i++;
                        artist = args[i];
                        break;

                    case "-s":
                        i++;
                        title = args[i];
                        break;

                    default:
                        break;
                }                

                i++;
            }

            if ((table == "") || (artist == "") || (title == ""))
            {
                Console.Writeline("You must supply a table name (-t TABLE), artist name (-a ARTIST), and song title (-s TITLE)");
                return;
            }                

            IAmazonDynamoDB client = new AmazonDynamoDBClient();

            Task<APIResponse> response = DoSomethingAsync(client, RESOURCE, ...);
        }
    }
}
```

## Testing your code

We use [moq4](https://github.com/moq/moq4) to create unit tests with mocked objects.

A typical unit test looks something like the following:

```
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

using Microsoft.VisualStudio.TestTools.UnitTesting;
using Microsoft.VisualStudio.TestTools.UnitTesting.Logging;

using Moq;

using System.Threading.Tasks;
using System.Threading;

namespace DotNetCoreConsoleTemplate
{

    [TestClass]
    public class CreateTableTest
    {
        string tableName = "testtable";

        private IAmazonDynamodDB CreateMockDynamoDBClient()
        {
            var mockDynamoDBClient = new Mock<IAmazonDynamoDB>();
             
            mockDynamoDBClient.Setup(client => client.PutTableAsync(It.IsAny<PutTableRequest>(), It.IsAny<CancellationToken>()))
                .Callback<PutTableRequest, CancellationToken>((request, token) =>
                {
                    if(!string.IsNullOrEmpty(tableName))
                    {
                        Assert.AreEqual(tableName, request.TableName);
                    }
                })
                .Returns((PutTableRequest r, CancellationToken token) =>
                {
                    return Task.FromResult(new PutTableResponse());
                });

            return mockDynamoDBClient.Object;
        }

        [TestMethod]
        public async Task CheckCreateTable()
        {
            IAmazonDynamodDB client = CreateMockDynamoDBClient();

            var result = await CreateTable.MakeTable(client, tableName);
            // log result
            /*Microsoft.VisualStudio.TestTools.UnitTesting.Logging.*/Logger.LogMessage("Created table {0}, tableName);
        }
    }
}
```

## Creating a table

```
public static async Task<PutTableResponse> MakeTable(IAmazonDynamoDB client, string table)
{
    using (client)
    {
        var response = await client.PutTableAsync(TableName: table);
        return response;
    }
}
```

## Getting information about a table

## Writing data to a table

## Reading data from a table

### Reading an item using its primary key

### Querying a table

### Scanning a table

## Managing indexes

### Creating an index

### Querying an index

### Scanning an index

## Modifying data in a table

## Deleting data from a table

## Removing a table

