using System;
using System.Collections.Generic;
using System.Configuration;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class GetItem
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
            Console.WriteLine("GetItem.exe ITEM-ID [-d] [-h]");
            Console.WriteLine("");
            Console.WriteLine("  -h prints this message and quits");
            Console.WriteLine("  -d prints extra (debugging) info");
        }

        static async Task<QueryResponse> GetItemAsync(IAmazonDynamoDB client, string table, string id)
        {
            var response = await client.QueryAsync(new QueryRequest
            {
                TableName = table,
                KeyConditionExpression = "ID = :v_Id",
                ExpressionAttributeValues = new Dictionary<string, AttributeValue>
                {
                    {
                        ":v_Id", new AttributeValue
                        {
                            S = id 
                        }
                    }
                }
            });

            return response;
        }
                
        static void Main(string[] args)
        {
            var debug = false;
            var configfile = "../../../../Config/app.config";
            var region = "";
            var table = "";
            var id = "";

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
                        id = args[i];
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

            DebugPrint(debug, "Debugging enabled");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<QueryResponse> response =  GetItemAsync(client, table, id);
                        
            foreach (var item in response.Result.Items)
            {
                foreach (string attr in item.Keys)
                {
                    if (item[attr].S != null)
                    {
                        Console.WriteLine(attr + ": " + item[attr].S);
                    }
                    else if (item[attr].N != null)
                    {
                        Console.WriteLine(attr + ": " + item[attr].N.ToString());
                    }
                }

                Console.WriteLine("");
            }
        }
    }
}
