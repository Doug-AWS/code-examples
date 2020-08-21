using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class CreateTable
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static async Task<bool> DoesTableExistAsync(IAmazonDynamoDB client, string table)
        {
            var resp = await client.ListTablesAsync(new ListTablesRequest { });

            foreach (var t in resp.TableNames)
            {
                if (t == table)
                {
                    return true;
                }
            }

            return false;
        }

        static async Task<CreateTableResponse> MakeTableAsync(bool debug, IAmazonDynamoDB client, string table)
        {
            var response = await client.CreateTableAsync(new CreateTableRequest
            {
                TableName = table,
                AttributeDefinitions = new List<AttributeDefinition>
                {
                    new AttributeDefinition
                    {
                        AttributeName = "Artist",
                        AttributeType = "S"
                    },
                    new AttributeDefinition
                    {
                        AttributeName = "SongTitle",
                        AttributeType = "S"
                    }
                },
                KeySchema = new List<KeySchemaElement>
                {
                    new KeySchemaElement
                    {
                        AttributeName = "Artist",
                        KeyType = "HASH"
                    },
                    new KeySchemaElement
                    {
                        AttributeName = "SongTitle",
                        KeyType = "RANGE"
                    },
                },
                ProvisionedThroughput = new ProvisionedThroughput
                {
                    ReadCapacityUnits = 10,
                    WriteCapacityUnits = 5
                }
            });

            if (debug)
            {
                Console.WriteLine("CreateTable response:");
                Console.WriteLine(response);
            }

            // Wait for table to be created
            bool ready = false;
            int wait = 1; // Milliseconds to wait

            while (!ready)
            {
                Thread.Sleep(wait);

                var resp = await client.DescribeTableAsync(new DescribeTableRequest
                {
                    TableName = table
                });

                ready = (resp.Table.TableStatus == TableStatus.ACTIVE);
                wait *= 2;
            }

            return response;
        }

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("CreateTable.exe [-t TABLE] [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine(" TABLE is optional, and defaults to Music");
            Console.WriteLine(" REGION is optional, and defaults to us-west-2");
            Console.WriteLine(" -h prints this message and quits");
        }

        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";
            string table = "Music";

            int i = 0;
            while (i < args.Length)
            {
                switch (args[i])
                {
                    case "-h":
                        Usage();
                        return;
                    case "-d":
                        debug = true;
                        break;
                    case "-r":
                        i++;
                        region = args[i];
                        break;
                    case "-t":
                        i++;
                        table = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            if (table == "")
            {
                Console.WriteLine("You must supply a non-empty table name (-t TABLE)");
                return;
            }

            DebugPrint(debug, "Debugging enabled\n");

            DebugPrint(debug, "Table  == " + table + "\n");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<bool> exists = DoesTableExistAsync(client, table);

            if (exists.Result)
            {
                Console.WriteLine("Table " + table + " already exists in region " + region);
                return;
            }

            Task<CreateTableResponse> response = MakeTableAsync(debug, client, table);

            Console.WriteLine("Created table " + response.Result.TableDescription.TableName + " in region " + region);
        }
    }
}
