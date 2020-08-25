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

Read the
[Best Practices for Modeling Relational Data in DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/bp-relational-modeling.html)
topic in the Amazon DynamoDB Developer Guide for information about moving from a relational database to Amazon DynamoDB.

**IMPORTANT**

NoSQL design requires a different mindset than RDBMS design. 
For an RDBMS, you can create a normalized data model without thinking about access patterns. 
You can then extend it later when new questions and query requirements arise. 
By contrast, in Amazon DynamoDB, 
you shouldn't start designing your schema until you know the questions that it needs to answer. 
Understanding the business problems and the application use cases up front is absolutely essential.

### A simple example

Let's take a simple order entry system,
with just three tables: Customers, Orders, and Products.
- Customer data includes a unique ID, their name, address, email address.
- Orders data includes a unique ID, customer ID, product ID, order date, and status.
- Products data includes a unique ID, description, quantity, and cost.

You might end up with access patterns including the following:

- Get all orders for all customers within a given date range
- Get all orders of a given product for all customers
- Get all products below a given quantity

In a relational database, these might be satisfied by the following SQL queries:

```
select * from Orders where Order_Date between '2020-05-04 05:00:00' and '2020-08-13 09:00:00'
select * from Orders where Order_Product = 'a1b2c3'
select * from Products where Product_Quantity < '100'
```

## Modeling data in Amazon DynamoDB

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
  The table we create in these code examples will contain 
  a partition key that uniquely identifies a record,
  which can be a customer, an order, or a product.

  Therefore, we'll create some global seconday indices
  to query the table.
  
- Partition key and sort key.
  In this case, you need not have a unique partition key,
  however, the combination of partition key and sort key must be unique.

We'll show you how to create all of these when you create a table,
and how to use them when you access a table.

## Modeling Customers, Orders, and Products in Amazon DynamoDB

Your Amazon DynamoDB schema to model these tables might look like:

| Key | Data Type | Description |
| --- | --- | ---
| ID | Number | The unique ID of the item
| Customer_ID | Number | The unique ID of a customer
| Customer_Name | String | The name of the customer
| Customer_Address | String | The address of the customer
| Customer_Email | String | The email address of the customer
| Order_ID | Number | The unique ID of an order
| Order_Customer | Number | The Customer_ID of a customer
| Order_Product | Number | The Product_ID of a product
| Order_Date | Number | When the order was made
| Order_Status | String | The status (open, in delivery, etc.) of the order
| Product_ID | Number | The unique ID of a product
| Product_Description | String | The description of the product
| Product_Quantity | Number | How many are in the warehouse
| Product_Cost | Number | The cost, in cents, of one product

## Creating the example databases

We'll use three CSV (comma-separated value) files to define a set of customers,
orders, and products.
Then we'll load that data into a relational database and Amazon DynamoDB.
Finally, we'll run some SQL commands against the relational database,
and show you the corresponding queries or scan against Amazon DynamoDB.

The three sets of data are in:

- *customers.csv*, which defines six customers
- *orders.csv*, which defines 12 orders
- *products.csv*, which defines six products

## General code pattern

It's important that you understand the new asynch/await programming model in the
[AWS SDK for .NET](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide).

These code examples use the following NuGet packages:

- AWSSDK.Core, v3.5.0
- AWSSDK,DynamoDBv2, v3.5.0

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
        // Use the interface to facilitate unit testing    
        static async Task<APIResponse> DOSOMETHINGAsync(IAmazonDynamoDB client, string RESOURCE, ...)
        {
            var response = await client.APIAsync(...
            ...
	       
            return response;

        }

        static void Main(string[] args)
        {
            string region = "us-west-2";        
            string table = "";
            string RESOURCE = "";

            int i = 0;
            while (i < args.Length)
            {
                switch (args[i])
                {
                    case "-t":
                        i++;
                        table = args[i];
                        break;

                    case "-r":
                        i++;
                        region = args[i];
                        break;

                    case "-R":
                        i++;
                        RESOURCE = args[i];
                        break;

                    default:
                        break;
                }                

                i++;
            }

            if ((table == "") || (RESOURCE == ""))
            {
                Console.Writeline("You must supply a table name (-t TABLE) and RESOURCE (-r RESOURCE)");
                return;
            }                

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<APIResponse> response = DOSOMETHINGAsync(client, RESOURCE, ...);
        }
    }
}
```

## Testing your code

We use [moq4](https://github.com/moq/moq4) to create unit tests with mocked objects.

A typical unit test looks something like the following,
which tests a call to **PutTableAsync**:

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
            Logger.LogMessage("Created table {0}, tableName);
        }
    }
}
```

## Listing all of the tables in a region

Use the **ListTables** project to list all of the tables in a region.
By default it lists the tables in **us-west-2**,
but you can change that using:

```-r REGION```

where *REGION* is the name of a region, such as **us-east-1**.

## Creating a table

Use the **CreateTable** project to create a table.
By default this project creates the table **CustomersOrdersProducts**,
with a partition key, **ID**.

## Listing the items in a table

Use the **ListItems** project to list the items in a table.
By default this project lists the items in the table **CustomersOrdersProducts**.

## Adding an item to the table

Use the **AddItem** project to add an item to a table.
Again, by default this project adds an item to the table **CustomersOrdersProducts**.

## Uploading items to a table

The **AddItems** project incorportes data from three comma-separated value (CSV) files to populate a table.
Much other projects, it adds the items to the table **CustomersOrdersProducts**.

## Reading data from a table

You can read data from an Amazon DynamoDB table using a number of techniques.

- By the item's primary key
- By searching for a particular item or items based on some criteria

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

