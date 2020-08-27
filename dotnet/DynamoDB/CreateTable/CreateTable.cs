using System;
using System.Collections.Generic;
using System.Configuration;
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

        static async Task<CreateTableResponse> MakeTableAsync(IAmazonDynamoDB client, string table)
        {
            var response = await client.CreateTableAsync(new CreateTableRequest
            {
                TableName = table,
                AttributeDefinitions = new List<AttributeDefinition>
                {
                    new AttributeDefinition
                    {
                        AttributeName = "ID",
                        AttributeType = "S"
                    },
                    new AttributeDefinition
                    {
                        AttributeName = "Area",
                        AttributeType = "S"
                    }
                },
                KeySchema = new List<KeySchemaElement>
                {
                    new KeySchemaElement
                    {
                        AttributeName = "ID",
                        KeyType = "HASH"
                    },
                    new KeySchemaElement
                    {
                        AttributeName = "Area",
                        KeyType = "RANGE"
                    }
                },
                ProvisionedThroughput = new ProvisionedThroughput
                {
                    ReadCapacityUnits = 10,
                    WriteCapacityUnits = 5
                }
            });            

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
            Console.WriteLine("CreateTable.exe [-h] [-d]");
            Console.WriteLine("");
            Console.WriteLine("  -h prints this message and quits");
            Console.WriteLine("  -d prints additional (debugging) info");
        }

        static void Main(string[] args)
        {
            var debug = false;
            var configfile = "../../../../Config/app.config";
            var region = "";
            var table = "";

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
                    default:
                        break;
                }

                i++;
            }

            // Get default region and table from config file
            var efm = new ExeConfigurationFileMap
            {
                ExeConfigFilename = configfile
            };

            Configuration configuration = ConfigurationManager.OpenMappedExeConfiguration(efm, ConfigurationUserLevel.None);

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

            DebugPrint(debug, "Debugging enabled\n");            

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<bool> exists = DoesTableExistAsync(client, table);

            if (exists.Result)
            {
                Console.WriteLine("Table " + table + " already exists in region " + region);
                return;
            }

            Task<CreateTableResponse> response = MakeTableAsync(client, table);

            Console.WriteLine("Created table " + response.Result.TableDescription.TableName + " in region " + region);
        }
    }
}
