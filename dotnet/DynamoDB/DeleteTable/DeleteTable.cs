using System;
using System.Configuration;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class DeleteTable
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static async Task<DeleteTableResponse> RemoveTableAsync(bool debug, IAmazonDynamoDB client, string table)
        {
            DebugPrint(debug, "Removing " + table + " table ");

            var response = await client.DeleteTableAsync(new DeleteTableRequest
            {
                TableName = table
            });

            return response;
        }

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("DeleteTable.exe [-h] [-d]");
            Console.WriteLine("");
            Console.WriteLine("  -h prints this message and quits");
            Console.WriteLine("  -d prints extra (debugging) info");
        }

        static void Main(string[] args)
        {
            var debug = false;
            var configfile = "app.config";
            var region = "";
            var table = "";

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
                    default:
                        break;
                }

                i++;
            }

            DebugPrint(debug, "Debugging enabled\n");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<DeleteTableResponse> response = RemoveTableAsync(debug, client, table);

            if (response.Result.HttpStatusCode == System.Net.HttpStatusCode.OK)
            {
                Console.WriteLine("Removed " + table + " table in " + region + " region");
            }
            else
            {
                Console.WriteLine("Could not remove " + table + " table");
            }
        }
    }
}
