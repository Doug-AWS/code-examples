using System;
using System.Collections.Generic;
using System.Text;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace CreateIndex
{
    class CreateIndex
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("CreateIndex.exe -i INDEX-NAME -m MAIN-KEY -s SECONDARY-KEY -p PROJECTIONS [-t TABLE] [-r REGION] [-h] [-d]");
            Console.WriteLine("");
            Console.WriteLine("  INDEX-NAME is required");
            Console.WriteLine("  MAIN-KEY is the partition key");
            Console.WriteLine("  SECONDARY-KEY is the sort key");
            Console.WriteLine("  PROJECTIONS are the keys we want to return values for");
            Console.WriteLine("");
            Console.WriteLine("  TABLE is optional, and defaults to CustomersOrdersProducts");
            Console.WriteLine("  REGION is optional, and defaults to us-west-2");
            Console.WriteLine("  -h prints this message and quits");
            Console.WriteLine("  -d print some extra (debugging) info");
        }

        static async Task<UpdateTableResponse> AddIndexAsync(IAmazonDynamoDB client, string table, string indexname, string partitionkey, string sortkey, string projections)
        {
            // Prepare projections
            string[] parts = projections.Split(" ");
            List<string> nonkeyattributes = new List<string>();

            for (int i = 0; i < parts.Length; i++)
            {
                nonkeyattributes.Add(parts[i]);
            }

            var newIndex = new CreateGlobalSecondaryIndexAction()
            {
                IndexName = indexname,
                ProvisionedThroughput = 
                {
                    ReadCapacityUnits = 1L,
                    WriteCapacityUnits = 1L
                },
                KeySchema = {
                new KeySchemaElement {
                    AttributeName = partitionkey, KeyType = "HASH"
                },
                new KeySchemaElement {
                    AttributeName = sortkey, KeyType = "RANGE"
                }
            },
                Projection = new Projection
                {
                    ProjectionType = "INCLUDE",
                    NonKeyAttributes = nonkeyattributes
                }
            };

            GlobalSecondaryIndexUpdate update = new GlobalSecondaryIndexUpdate
            {
                Create = newIndex
            };

            List<GlobalSecondaryIndexUpdate> updates = new List<GlobalSecondaryIndexUpdate>();

            var response = await client.UpdateTableAsync(new UpdateTableRequest
            {
                GlobalSecondaryIndexUpdates = updates
            });

            return response;
        }
        

        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";
            string table = "CustomersOrdersProducts";
            string indexname = "";
            string mainkey = "";
            string secondarykey = "";
            string projections = "";

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
                    case "-i":
                        i++;
                        indexname = args[i];
                        break;
                    case "-m":
                        i++;
                        mainkey = args[i];
                        break;
                    case "-p":
                        i++;
                        projections = args[i];
                        break;
                    case "-r":
                        i++;
                        region = args[i];
                        break;
                    case "-s":
                        i++;
                        secondarykey = args[i];
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

            if ((region == "") || (table == "") || (indexname == "") || (mainkey == "") || (secondarykey == "") || (projections == ""))            
            {
                Console.WriteLine("You must supply a non-empty table name (-t TABLE), region (-r REGIoN), index name (-i INDEX), main and secondary keys (-m MAIN -s SECONDARY) and projections (-p PROJECTIONS)");
                return;
            }

            DebugPrint(debug, "Debugging enabled");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<UpdateTableResponse> response = AddIndexAsync(client, table, indexname, mainkey, secondarykey, projections);
        }
    }
}
