using System;
using System.Configuration;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.DocumentModel;

namespace DynamoDBCRUD
{
    class DeleteItem
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static async Task<bool> RemoveItemAsync(bool debug, IAmazonDynamoDB client, string table, string id, string area)
        {
            DebugPrint(debug, "Removing customer with ID " + id + " from " + table + " table ");

            var theTable = Table.LoadTable(client, table);
            var item = new Document();
            item["ID"] = id;
            item["Area"] = area;
            Document document = await theTable.DeleteItemAsync(item);
            
            return true;
        }

        /*
        static async Task<DeleteItemResponse> RemoveItemAsync(bool debug, IAmazonDynamoDB client, string table, string id, string area)
        {
            DebugPrint(debug, "Removing item with ID " + id + " and Area " + area + " from " + table + " table");

            var request = new DeleteItemRequest
            {
                TableName = table,
                Key = new Dictionary<string, AttributeValue>() 
                {
                    {
                        "ID",
                        new AttributeValue { S = id }
                    },
                    {
                        "Area",
                        new AttributeValue { S = area }
                    },
                }                
            };
            
            var response = await client.DeleteItemAsync(request);

            return response;
        }
        */

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("DeleteItem.exe -p PARTITION-KEY -s SORT-KEY [-h] [-d]");
            Console.WriteLine("");
            Console.WriteLine("Both PARTITION-KEY and SORT-KEY are required");
            Console.WriteLine("  -h prints this message and quits");
            Console.WriteLine("  -d prints extra (debugging) info");
        }

        static void Main(string[] args)
        {
            var debug = false;
            var configfile = "app.config";
            var region = "";
            var table = "";
            var partition = "";
            var sort = "";

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
                    Console.WriteLine("You must specify a Region and Table value in " + configfile);
                    return;
                }
            }
            else
            {
                Console.WriteLine("Could not find " + configfile);
                return;
            }

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
                    case "-p":
                        i++;
                        partition = args[i];
                        break;
                    case "-s":
                        i++;
                        sort = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            if ((partition == "") || (sort == ""))
            {
                Console.WriteLine("You must supply a partition key and sort key");
                Usage();
                return;
            }

            DebugPrint(debug, "Debugging enabled\n");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<bool> resp = RemoveItemAsync(debug, client, table, partition, sort);

            //    Task<DeleteItemResponse> response = RemoveItemAsync(debug, client, table, partition, sort);

            if (resp.Result)
            {
                Console.WriteLine("Removed item from " + table + " table in " + region + " region");
            }
        }
    }
}
