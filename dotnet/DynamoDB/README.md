# Amazon DynamoDB code examples in C#

This folder contains code examples for moving from SQL to NoSQL, 
specifically Amazon DynamoDB,
as described in the Amazon DynamoDB Developer Guide at
[From SQL to NoSQL](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/SQLtoNoSQL.html).

All of these code examples are written in C#, 
using the beta version of the AWS SDK for .NET.
Getting the 3.5 bits is straightforward using the command line 
from the same folder as your ```.csproj``` file.
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
topic in the Amazon DynamoDB Developer Guide for information about moving 
from a relational database to Amazon DynamoDB.

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
select * from Orders where Order_Product = '3'
select * from Products where Product_Quantity < '100'
```

Given the data in *customers.csv*, *orders.csv*, and *products.csv*,
these queries return (as CSV):

```
Order_ID,Order_Customer,Order_Product,Order_Date,Order_Status
1,1,6,"2020-07-04 12:00:00",pending
11,5,4,"2020-05-11 12:00:00",delivered
12,6,6,"2020-07-04 12:00:00",delivered

Order_ID,Order_Customer,Order_Product,Order_Date,Order_Status
4,4,3,"2020-04-01 12:00:00",backordered
8,2,3,"2019-01-01 12:00:00",backordered

Product_ID,Product_Description,Product_Quantity,Product_Cost
4,"2'x50' plastic sheeting",45,450
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
| ID | String | The unique ID of the item
| Type | String | Customer, Order, or Product
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

All of the projects that require a table name use the default table
**CustomersOrdersProducts**.

## General code pattern

It's important that you understand the new asynch/await programming model in the
[AWS SDK for .NET](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide).

These code examples use the following NuGet packages:

- AWSSDK.Core, v3.5.0
- AWSSDK,DynamoDBv2, v3.5.0

The projects use a configuration file to define the region, table name,
and files defining the customers, orders, and products for all of the examples.
Here's the current version:

```
<configuration>
 <appSettings>
    <add key="Region" value="us-west-2" />
    <add key="Table" value="CustomersOrdersProducts" />
    <add key="Customers" value="Customers.csv" />
    <add key="Orders" value="Orders.csv" />
    <add key="Products" value="Products.csv" />
 </appSettings>
</configuration>
```

So if you want to use a different table name or region, it's easy to do so.

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
            var configfile = "../../../../Config/app.config";
            
            // Get default region and table from config file
            var efm = new ExeConfigurationFileMap
            {
                ExeConfigFilename = configfile
            };

            Configuration configuration = ConfigurationManager.OpenMappedExeConfiguration(efm, 
                ConfigurationUserLevel.None);

            if (configuration.HasFile)
            {
                AppSettingsSection appSettings = configuration.AppSettings;
                region = appSettings.Settings["Region"].Value;
                table = appSettings.Settings["Table"].Value;

                if ((region == "") || (table == ""))
                {
                    Console.WriteLine("You must specify Region and Table values in " + configfile);                    
                    return;
                }
            }
            else
            {
                Console.WriteLine("Could not find " + configfile);
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
             
            mockDynamoDBClient.Setup(client => client.PutTableAsync(
                It.IsAny<PutTableRequest>(), 
                It.IsAny<CancellationToken>()))
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

## Creating a table

Use the **CreateTable** project to create a table
with the partition key, **ID**.

## Listing the items in a table

Use the **ListItems** project to list the items in a table.

## Adding an item to the table

Use the **AddItem** project to add an item to a table.
It requires the following options:

- ```-k``` *KEYS*, where *KEYS* is a a comma-separated list of keys
- ```-v``` *VALUES*, where *VALUES* is a a comma-separated list of values

There must be the same number of keys as values.

- If the item is a customer, the schema should match that in *customers.csv*.
- If the item is an order, the schema should match that in *orders.csv*.
- If the item is a product, the schema should match that in *products.csv*.

## Uploading items to a table

The **AddItems** project incorporates data from three comma-separated value 
(CSV) files to populate a table.
It gets the values of these files from the configuration file 
as described in the previous *General code pattern* section.

## Reading data from a table

You can read data from an Amazon DynamoDB table using a number of techniques.

- By the item's primary key
- By searchng for a particular item or items based on the value of one or more keys

### Reading an item using its primary key

Use the **GetItem** project 
to retrieve information about the customer, order, 
or product with the given primary key.
It requires the following option:

- ```-i``` *ID*, which is the partition ID of the item in the table.

### Scanning a table

Use the **ScanTable** project to perform the same queries we showed in the
*A simple example* section:

- Get all orders for all customers within a given date range
- Get all orders of a given product for all customers
- Get all products below a given quantity

To replicate the SQL queries in the
*A simple example* section, call **ScanTable.exe** as shown:

- ```ScanTable.exe -r```, which invokes the **GetOrdersInDateRangeAsync** method
- ```ScanTable.exe -p```, which invokes the **GetProductOrdersAsync** method
- ```ScanTable.exe -l```, which invokes the **GetLowStockAsync** method

#### GetOrdersInDateRangeAsync
```
static async Task<ScanResponse> GetOrdersInDateRangeAsync(IAmazonDynamoDB client, string table, string start, string end)
{
    // Convert start and end strings to longs
    var StartDateTime = DateTime.ParseExact(start, "yyyy-MM-dd HH:mm:ss", CultureInfo.InvariantCulture);
    var EndDateTime = DateTime.ParseExact(end, "yyyy-MM-dd HH:mm:ss", CultureInfo.InvariantCulture);

    TimeSpan startTimeSpan = StartDateTime - new DateTime(1970, 1, 1, 0, 0, 0);
    TimeSpan endTimeSpan = EndDateTime - new DateTime(1970, 1, 1, 0, 0, 0);

    var begin = (long)startTimeSpan.TotalSeconds;
    var finish = (long)endTimeSpan.TotalSeconds;

    var response = await client.ScanAsync(new ScanRequest
    {
        TableName = table,
        ExpressionAttributeValues = new Dictionary<string, AttributeValue> {
            {":startval", new AttributeValue { N = begin.ToString() } },
            {":endval", new AttributeValue { N = finish.ToString()} }
        },
        FilterExpression = "Order_Date > :startval AND Order_Date < :endval",
        ProjectionExpression = "Order_ID, Order_Customer, Order_Product, Order_Date, Order_Status"
    });

    return response;
}
```

#### GetProductOrdersAsync
```
static async Task<ScanResponse> GetProductOrdersAsync(bool debug, IAmazonDynamoDB client, string table, string productId)
{
    var response = await client.ScanAsync(new ScanRequest
    {
        TableName = table,
        ExpressionAttributeValues = new Dictionary<string, AttributeValue> {
            {":val", new AttributeValue { N = productId }}
        },
        FilterExpression = "Order_Product = :val",
        ProjectionExpression = "Order_ID,Order_Customer,Order_Product,Order_Date,Order_Status"
    });

    return response;
}
```

#### GetLowStockAsync
```
static async Task<ScanResponse> GetLowStockAsync(bool debug, IAmazonDynamoDB client, string table, string volume)
{
    var response = await client.ScanAsync(new ScanRequest
    {
        TableName = table,
        ExpressionAttributeValues = new Dictionary<string, AttributeValue> {
            {":val", new AttributeValue { N = volume }}
        },
        FilterExpression = "Product_Quantity < :val",
        ProjectionExpression = "Product_ID, Product_Description, Product_Quantity, Product_Cost"
    });

    return response;
}
```

## Managing indexes

Global secondary indices give you the ability to treat a set of 
Amazon DynamoDB table keys as if they were a separate table.

### Creating an index

Use the **CreateIndex** project to create an index.
It requires the following command-line options:

- ```-i``` *INDEX-NAME*, where *INDEX-NAME* is the name of the index
- ```-m``` *MAIN-KEY*, where *MAIN-KEY* is the partition key of the index
- ```-k``` *MAIN-KEY-TYPE*, where *MAIN-KEY-TYPE* is one of string or number
- ```-s``` *SECONDARY-KEY*, where *SECONDARY-KEY* is the sort key of the index
- ```-t``` *SECONDARY-KEY-TYPE*, where *SECONDARY-KEY-TYPE* is one of string or number

To create a global secondary index (GSI) for customers, orders, and products,
execute the following commands, 
where the **-m** flag defines the main (partition key) value 
and the **-s** flag defines the secondary (sort key) value:

```
CreateIndex.exe -i Customers -m Customer_ID -k number -s Customer_Email   -t string
CreateIndex.exe -i Orders    -m Order_ID    -k number -s Order_Date       -t number
CreateIndex.exe -i Products  -m Product_ID  -k number -s Product_Quantity -t number
```

## Modifying a table item

Use the **UpdateItem** project to modify the status of an order in the table.
It takes the following options:

- ```-i``` *ID*, where *ID* is the value of the order's ORDER_ID attribute
- ```-s``` *STATUS*, where *STATUS* is the new status value (backordered, delivered, delivering, pending)C

## Deleting an item from a table

## Removing a table

